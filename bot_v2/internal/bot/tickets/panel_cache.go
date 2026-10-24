package tickets

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"golang.org/x/sync/singleflight"
)

// Configuration writers exclude admissions and the final channel-send check.
// This also prevents a cache fill racing a committed dashboard edit.
var panelPolicyMu sync.RWMutex

type questionSnapshot struct {
	title     string
	questions []QuestionData
	version   string
}

func questionVersion(title string, questions []QuestionData) string {
	encoded, _ := json.Marshal(questionSnapshotForHash{title, questions})
	sum := sha256.Sum256(encoded)
	return fmt.Sprintf("%x", sum[:8])
}

type questionSnapshotForHash struct {
	Title     string
	Questions []QuestionData
}

type questionCache struct {
	mu      sync.Mutex
	entries map[int32]questionSnapshot
	loads   singleflight.Group
	load    func(context.Context, int32) (questionSnapshot, error)
}

var modalQuestions = questionCache{entries: make(map[int32]questionSnapshot), load: func(ctx context.Context, panelID int32) (questionSnapshot, error) {
	title, questions, err := loadPanelQuestions(ctx, panelID)
	return questionSnapshot{title: title, questions: questions}, err
}}

func (c *questionCache) get(ctx context.Context, panelID int32) (questionSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return questionSnapshot{}, err
	}
	c.mu.Lock()
	cached, ok := c.entries[panelID]
	c.mu.Unlock()
	if ok {
		return cached, nil
	}
	key := strconv.Itoa(int(panelID))
	result, err, _ := c.loads.Do(key, func() (any, error) {
		c.mu.Lock()
		cached, ok := c.entries[panelID]
		c.mu.Unlock()
		if ok {
			return cached, nil
		}
		loadCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		value, err := c.load(loadCtx, panelID)
		if err != nil {
			return questionSnapshot{}, err
		}
		value.version = questionVersion(value.title, value.questions)
		c.mu.Lock()
		if len(c.entries) >= 256 {
			for id := range c.entries {
				delete(c.entries, id)
				break
			}
		}
		c.entries[panelID] = value
		c.mu.Unlock()
		return value, nil
	})
	// Wait for the shared fill before releasing panelPolicyMu, even on timeout.
	// Otherwise an old fill could repopulate the cache after invalidation.
	if ctx.Err() != nil {
		return questionSnapshot{}, ctx.Err()
	}
	if err != nil {
		return questionSnapshot{}, err
	}
	return result.(questionSnapshot), nil
}

func (c *questionCache) invalidate(panelID int32) {
	c.mu.Lock()
	delete(c.entries, panelID)
	c.mu.Unlock()
	c.loads.Forget(strconv.Itoa(int(panelID)))
}

type admissionSnapshot struct {
	questionSnapshot
	window db.PanelScheduleWindow
}

var admissionReads singleflight.Group
var admissionSlots = make(chan struct{}, 4)

// Caller holds panelPolicyMu.RLock. Schedule results are shared only by overlapping
// requests, never retained; eligibility is always computed against the current clock.
func loadAdmission(ctx context.Context, serverID int64, panelID int32) (admissionSnapshot, error) {
	result, err, _ := admissionReads.Do(fmt.Sprintf("%d:%d", serverID, panelID), func() (any, error) {
		select {
		case admissionSlots <- struct{}{}:
			defer func() { <-admissionSlots }()
		default:
			return nil, errors.New("panel reads busy")
		}
		if _, err := queries.GetPanelConfigByID(ctx, serverID, panelID); err != nil {
			return nil, err
		}
		windows, err := queries.ListScheduleWindowsByPanel(ctx, panelID)
		if err != nil {
			return nil, err
		}
		questions, err := modalQuestions.get(ctx, panelID)
		return struct {
			questions questionSnapshot
			windows   []db.PanelScheduleWindow
		}{questions, windows}, err
	})
	if err != nil {
		return admissionSnapshot{}, err
	}
	data := result.(struct {
		questions questionSnapshot
		windows   []db.PanelScheduleWindow
	})
	window, err := activeWindow(data.windows, time.Now())
	return admissionSnapshot{data.questions, window}, err
}

func activeWindow(windows []db.PanelScheduleWindow, now time.Time) (db.PanelScheduleWindow, error) {
	if len(windows) == 0 {
		return db.PanelScheduleWindow{}, nil
	}
	next := int64(0)
	for _, w := range windows {
		if w.ClosedEarly {
			continue
		}
		if w.StartTime <= now.Unix() && w.EndTime > now.Unix() {
			return w, nil
		}
		if w.StartTime > now.Unix() && (next == 0 || w.StartTime < next) {
			next = w.StartTime
		}
	}
	return db.PanelScheduleWindow{}, &errWindowClosed{NextStartTime: next}
}

// BeginPanelUpdate must surround every dashboard mutation of an existing panel
// or its schedules, including mutations which can partially succeed.
func BeginPanelUpdate(serverID int64, panelID int32) func() {
	panelPolicyMu.Lock()
	return func() {
		defer panelPolicyMu.Unlock()
		modalQuestions.invalidate(panelID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, panelErr := queries.GetPanelConfigByID(ctx, serverID, panelID)
		windows, err := queries.ListScheduleWindowsByPanel(ctx, panelID)
		openingQueue.reconcile(panelID, windows, panelErr != nil || err != nil)
	}
}

func validateJobWindow(ctx context.Context, j *ticketJob) error {
	windows, err := queries.ListScheduleWindowsByPanel(ctx, j.panelID)
	if err != nil {
		return err
	}
	current, err := activeWindow(windows, time.Now())
	if err != nil {
		return err
	}
	j.queue.mu.Lock()
	defer j.queue.mu.Unlock()
	if current.ID != j.window.ID {
		return errJobCanceled
	}
	if current.ID != 0 {
		// Read/update job policy only under its queue lock.
		j.window = current
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return errJobCanceled
	}
	return ctx.Err()
}
