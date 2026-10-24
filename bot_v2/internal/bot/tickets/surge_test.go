package tickets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func jobFor(key string) *ticketJob {
	return &ticketJob{key: key, panelID: 1, queuedAt: time.Now()}
}

func stopTestQueue(t *testing.T, q *ticketQueue) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	q.stop(ctx)
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition did not become true")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestBurst500BoundedWorkers(t *testing.T) {
	var active, maximum, completed atomic.Int32
	var waitsMu sync.Mutex
	var waits []time.Duration
	q := newTicketQueue(500, func(ctx context.Context, j *ticketJob) error {
		n := active.Add(1)
		for old := maximum.Load(); n > old && !maximum.CompareAndSwap(old, n); old = maximum.Load() {
		}
		waitsMu.Lock()
		waits = append(waits, time.Since(j.queuedAt))
		waitsMu.Unlock()
		time.Sleep(time.Millisecond)
		active.Add(-1)
		completed.Add(1)
		return nil
	}, func(*ticketJob, string) {})
	t.Cleanup(func() { stopTestQueue(t, q) })
	for n := 0; n < 500; n++ {
		if err := q.enqueue(jobFor(fmt.Sprint(n))); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.enqueue(jobFor("overflow")); !errors.Is(err, errQueueFull) {
		t.Fatalf("overflow: %v", err)
	}
	q.start(2)
	eventually(t, func() bool { return completed.Load() == 500 })
	if maximum.Load() != 2 {
		t.Fatalf("maximum concurrency = %d", maximum.Load())
	}
	waitsMu.Lock()
	sort.Slice(waits, func(a, b int) bool { return waits[a] < waits[b] })
	t.Logf("mock burst: max_workers=%d p95_wait=%s max_wait=%s", maximum.Load(), waits[474], waits[499])
	waitsMu.Unlock()
}

func TestQueueFIFOAndDuplicates(t *testing.T) {
	order := make(chan string, 3)
	answers := make(chan string, 1)
	q := newTicketQueue(3, func(ctx context.Context, j *ticketJob) error {
		if j.key == "server:user" {
			answers <- j.qna[0].Answer
		}
		order <- j.key
		return nil
	}, func(*ticketJob, string) {})
	t.Cleanup(func() { stopTestQueue(t, q) })
	first := jobFor("server:user")
	first.qna = []QnA{{Question: "Original", Answer: "First"}}
	if err := q.enqueue(first); err != nil {
		t.Fatal(err)
	}
	repeated := jobFor(first.key)
	repeated.panelID = 2
	if err := q.enqueue(repeated); !errors.Is(err, errPending) {
		t.Fatal("duplicate across panels admitted")
	}
	for _, key := range []string{"second", "third"} {
		if err := q.enqueue(jobFor(key)); err != nil {
			t.Fatal(err)
		}
	}
	q.start(1)
	for _, want := range []string{first.key, "second", "third"} {
		select {
		case got := <-order:
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("worker stalled")
		}
	}
	if <-answers != "First" {
		t.Fatal("duplicate replaced original answers")
	}
}

func TestDuplicateWhileRunning(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	q := newTicketQueue(2, func(ctx context.Context, j *ticketJob) error {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil
	}, func(*ticketJob, string) {})
	t.Cleanup(func() { stopTestQueue(t, q) })
	q.enqueue(jobFor("user"))
	q.start(1)
	<-entered
	if err := q.enqueue(jobFor("user")); !errors.Is(err, errPending) {
		t.Fatal("running duplicate admitted")
	}
	close(release)
	eventually(t, func() bool { return !q.hasPending("user") })
}

func TestWindowChanges(t *testing.T) {
	for _, action := range []string{"close", "delete", "panel delete", "move", "extend"} {
		t.Run(action, func(t *testing.T) {
			q := newTicketQueue(2, func(context.Context, *ticketJob) error { return nil }, func(*ticketJob, string) {})
			t.Cleanup(func() { stopTestQueue(t, q) })
			j := jobFor("user")
			j.window = db.PanelScheduleWindow{ID: 7, PanelConfigID: 1, StartTime: time.Now().Unix() - 1, EndTime: time.Now().Unix() + 60}
			q.enqueue(j)
			w := j.window
			unavailable := false
			windows := []db.PanelScheduleWindow{w}
			switch action {
			case "close":
				windows[0].ClosedEarly = true
			case "delete":
				windows = nil
			case "panel delete":
				unavailable = true
			case "move":
				windows[0].StartTime = time.Now().Unix() + 10
			case "extend":
				windows[0].EndTime += 60
			}
			q.reconcile(1, windows, unavailable)
			if action == "extend" {
				if !q.hasPending("user") {
					t.Fatal("extension canceled valid request")
				}
				return
			}
			if q.hasPending("user") {
				t.Fatal("closed request still pending")
			}
			q.reconcile(1, []db.PanelScheduleWindow{w}, false)
			if q.hasPending("user") {
				t.Fatal("reopening restored discarded request")
			}
			if notice := <-q.notices; notice.message != windowCanceledMessage {
				t.Fatalf("wrong notice: %q", notice.message)
			}
		})
	}
}

func TestNaturalClosureAndExpiry(t *testing.T) {
	for _, scheduled := range []bool{false, true} {
		t.Run(fmt.Sprint(scheduled), func(t *testing.T) {
			q := newTicketQueue(1, nil, func(*ticketJob, string) {})
			t.Cleanup(func() { stopTestQueue(t, q) })
			j := jobFor("user")
			if scheduled {
				j.window = db.PanelScheduleWindow{ID: 1, StartTime: time.Now().Unix() - 1, EndTime: time.Now().Unix()}
			} else {
				j.queuedAt = time.Now().Add(-queueLifetime)
			}
			q.enqueue(j)
			eventually(t, func() bool { return !q.hasPending(j.key) })
		})
	}
}

func TestClosureStopsPreparingButFinishesSentRequest(t *testing.T) {
	for _, sent := range []bool{false, true} {
		t.Run(fmt.Sprint(sent), func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var q *ticketQueue
			q = newTicketQueue(2, func(ctx context.Context, j *ticketJob) error {
				if sent {
					if err := q.markStarted(j); err != nil {
						return err
					}
				}
				close(entered)
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, func(*ticketJob, string) {})
			t.Cleanup(func() { stopTestQueue(t, q) })
			j := jobFor("user")
			j.window = db.PanelScheduleWindow{ID: 1, StartTime: time.Now().Unix() - 1, EndTime: time.Now().Unix() + 60}
			q.enqueue(j)
			q.start(1)
			<-entered
			q.reconcile(1, nil, false)
			if sent {
				if j.ctx.Err() != nil {
					t.Fatal("sent request canceled")
				}
				close(release)
			} else {
				eventually(t, func() bool { return j.ctx.Err() != nil })
			}
			eventually(t, func() bool { return !q.hasPending("user") })
		})
	}
}

func TestShutdownBoundedAndRejectsNewWork(t *testing.T) {
	entered := make(chan struct{})
	var q *ticketQueue
	q = newTicketQueue(2, func(ctx context.Context, j *ticketJob) error {
		q.markStarted(j)
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}, func(*ticketJob, string) {})
	q.enqueue(jobFor("running"))
	q.start(1)
	<-entered
	q.enqueue(jobFor("waiting"))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	q.stop(ctx)
	if time.Since(started) > 250*time.Millisecond {
		t.Fatal("shutdown exceeded budget")
	}
	if q.hasPending("waiting") {
		t.Fatal("waiting request retained")
	}
	if err := q.enqueue(jobFor("new")); !errors.Is(err, errQueueStopped) {
		t.Fatal("accepted after shutdown")
	}
}

func TestQuestionCacheCoalescingBoundAndInvalidation(t *testing.T) {
	var calls atomic.Int32
	version := "old"
	c := questionCache{entries: make(map[int32]questionSnapshot), load: func(context.Context, int32) (questionSnapshot, error) {
		calls.Add(1)
		time.Sleep(time.Millisecond)
		return questionSnapshot{title: version}, nil
	}}
	var wg sync.WaitGroup
	for n := 0; n < 50; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.get(context.Background(), 1); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("cache loads=%d", calls.Load())
	}
	version = "new"
	c.invalidate(1)
	v, err := c.get(context.Background(), 1)
	if err != nil || v.title != "new" {
		t.Fatalf("stale cache: %+v %v", v, err)
	}
	for n := int32(2); n < 270; n++ {
		c.get(context.Background(), n)
	}
	if len(c.entries) > 256 {
		t.Fatalf("cache grew to %d", len(c.entries))
	}
}

type stubRow func(...any) error

func (r stubRow) Scan(dest ...any) error { return r(dest...) }

type stubRows struct {
	pgx.Rows
	values   [][]any
	position int
}

func (r *stubRows) Next() bool {
	if r.position >= len(r.values) {
		return false
	}
	r.position++
	return true
}
func (r *stubRows) Scan(dest ...any) error {
	for n, v := range r.values[r.position-1] {
		reflect.ValueOf(dest[n]).Elem().Set(reflect.ValueOf(v))
	}
	return nil
}
func (*stubRows) Close()     {}
func (*stubRows) Err() error { return nil }

type surgeDB struct {
	mu             sync.Mutex
	windows        []db.PanelScheduleWindow
	question       string
	maxTickets     int32
	count          int64
	activeChannels []string
	insertErr      error
	readErr        error
	deletes        int
	reads          atomic.Int32
	readDelay      time.Duration
	reading        atomic.Int32
	maxReading     atomic.Int32
}

func (d *surgeDB) read(ctx context.Context) error {
	d.reads.Add(1)
	n := d.reading.Add(1)
	defer d.reading.Add(-1)
	for old := d.maxReading.Load(); n > old && !d.maxReading.CompareAndSwap(old, n); old = d.maxReading.Load() {
	}
	if d.readDelay != 0 {
		select {
		case <-time.After(d.readDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return ctx.Err()
}

func (d *surgeDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return stubRow(func(dest ...any) error {
		if err := d.read(ctx); err != nil {
			return err
		}
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.readErr != nil {
			return d.readErr
		}
		for _, v := range dest {
			reflect.ValueOf(v).Elem().SetZero()
		}
		switch {
		case strings.Contains(sql, "COUNT(*)"):
			*dest[0].(*int64) = d.count
		case strings.Contains(sql, "FROM server_config"):
			*dest[3].(*int32) = d.maxTickets
		case strings.Contains(sql, "FROM questions_config"):
			*dest[0].(*[]string) = []string{d.question}
		}
		return nil
	})
}
func (d *surgeDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if err := d.read(ctx); err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.readErr != nil {
		return nil, d.readErr
	}
	r := &stubRows{}
	if strings.Contains(sql, "panel_schedule_window") {
		for _, w := range d.windows {
			r.values = append(r.values, []any{w.ID, w.PanelConfigID, w.StartTime, w.EndTime, w.ClosedEarly, w.CreatedAt})
		}
	}
	if strings.Contains(sql, "FROM active_ticket") {
		for _, id := range d.activeChannels {
			r.values = append(r.values, []any{id})
		}
	}
	return r, nil
}
func (d *surgeDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if strings.Contains(sql, "INSERT INTO active_ticket") {
		return pgconn.CommandTag{}, d.insertErr
	}
	if strings.Contains(sql, "DELETE FROM active_ticket") {
		d.deletes++
		d.count--
	}
	return pgconn.CommandTag{}, nil
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func installSurgeDB(t *testing.T, d *surgeDB) {
	t.Helper()
	old := queries
	queries = db.New(d)
	modalQuestions.invalidate(1)
	t.Cleanup(func() { queries = old; modalQuestions.invalidate(1) })
}

func interaction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "123", AppID: "app", Token: "secret", GuildID: "1", Member: &discordgo.Member{User: &discordgo.User{ID: "user", Username: "test"}}}}
}

func TestAcknowledgmentBeforeDBAndFailureStopsAdmission(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(fmt.Sprint(failed), func(t *testing.T) {
			d := &surgeDB{readErr: errors.New("DB offline")}
			installSurgeDB(t, d)
			s, _ := discordgo.New("Bot test")
			s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				if strings.Contains(r.URL.Path, "/callback") {
					if d.reads.Load() != 0 {
						t.Error("DB queried before acknowledgment")
					}
					if failed {
						return response(500, `{}`), nil
					}
				}
				return response(200, `{}`), nil
			})}
			i := interaction()
			i.Type = discordgo.InteractionModalSubmit
			i.Data = discordgo.ModalSubmitInteractionData{CustomID: "ticket_modal_1"}
			start := time.Now()
			HandleModalSubmit(s, i)
			t.Logf("mock acknowledgment path=%s failed=%t", time.Since(start), failed)
			if failed && d.reads.Load() != 0 {
				t.Fatal("failed acknowledgment admitted DB work")
			}
		})
	}
}

func TestDashboardMutationInvalidatesEvenPartialFailure(t *testing.T) {
	d := &surgeDB{question: "old"}
	installSurgeDB(t, d)
	panelPolicyMu.RLock()
	old, err := modalQuestions.get(context.Background(), 1)
	panelPolicyMu.RUnlock()
	if err != nil || old.questions[0].Label != "old" {
		t.Fatal("initial read failed")
	}
	finish := BeginPanelUpdate(1, 1)
	d.mu.Lock()
	d.question = "committed before later failure"
	d.mu.Unlock()
	finish()
	panelPolicyMu.RLock()
	updated, err := modalQuestions.get(context.Background(), 1)
	panelPolicyMu.RUnlock()
	if err != nil || updated.questions[0].Label != d.question {
		t.Fatal("partial mutation left stale questions")
	}
}

func TestScheduleReadsAreNotRetained(t *testing.T) {
	d := &surgeDB{question: "question", windows: []db.PanelScheduleWindow{{ID: 1, StartTime: time.Now().Unix() - 1, EndTime: time.Now().Unix() + 60}}}
	installSurgeDB(t, d)
	panelPolicyMu.RLock()
	_, err := loadAdmission(context.Background(), 1, 1)
	panelPolicyMu.RUnlock()
	if err != nil {
		t.Fatal(err)
	}
	d.mu.Lock()
	d.windows[0].ClosedEarly = true
	d.mu.Unlock()
	panelPolicyMu.RLock()
	_, err = loadAdmission(context.Background(), 1, 1)
	panelPolicyMu.RUnlock()
	var closed *errWindowClosed
	if !errors.As(err, &closed) {
		t.Fatal("schedule read stayed stale")
	}
}

func TestAdmissionBurstCoalescesDBReads(t *testing.T) {
	d := &surgeDB{question: "question", readDelay: 10 * time.Millisecond}
	installSurgeDB(t, d)
	start := make(chan struct{})
	var wg sync.WaitGroup
	startedAt := time.Now()
	for n := 0; n < 500; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			panelPolicyMu.RLock()
			_, err := loadAdmission(ctx, 1, 1)
			panelPolicyMu.RUnlock()
			if err != nil {
				t.Error(err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if d.maxReading.Load() != 1 {
		t.Fatalf("same-panel reads did not coalesce: max=%d", d.maxReading.Load())
	}
	if d.reads.Load() >= 500 {
		t.Fatalf("burst produced %d DB reads", d.reads.Load())
	}
	t.Logf("mock DB burst: users=500 reads=%d max_concurrent_reads=%d elapsed=%s", d.reads.Load(), d.maxReading.Load(), time.Since(startedAt))
}

func TestCacheFillCannotOutliveDashboardInvalidation(t *testing.T) {
	loadStarted, finishLoad, readerDone, writerDone := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	c := questionCache{entries: make(map[int32]questionSnapshot), load: func(context.Context, int32) (questionSnapshot, error) {
		if calls.Add(1) == 1 {
			close(loadStarted)
			<-finishLoad
			return questionSnapshot{title: "old"}, nil
		}
		return questionSnapshot{title: "new"}, nil
	}}
	go func() {
		panelPolicyMu.RLock()
		c.get(context.Background(), 1)
		panelPolicyMu.RUnlock()
		close(readerDone)
	}()
	<-loadStarted
	go func() {
		panelPolicyMu.Lock()
		c.invalidate(1)
		panelPolicyMu.Unlock()
		close(writerDone)
	}()
	select {
	case <-writerDone:
		t.Fatal("writer passed an active cache fill")
	case <-time.After(10 * time.Millisecond):
	}
	close(finishLoad)
	<-readerDone
	<-writerDone
	panelPolicyMu.RLock()
	value, err := c.get(context.Background(), 1)
	panelPolicyMu.RUnlock()
	if err != nil || value.title != "new" {
		t.Fatal("old cache fill survived invalidation")
	}
}

func TestOldFormRejectedAfterQuestionEdit(t *testing.T) {
	d := &surgeDB{question: "new question"}
	installSurgeDB(t, d)
	s, _ := discordgo.New("Bot test")
	var message string
	s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodPatch {
			body, _ := io.ReadAll(r.Body)
			message = string(body)
		}
		return response(200, `{}`), nil
	})}
	i := interaction()
	i.Type = discordgo.InteractionModalSubmit
	i.Data = discordgo.ModalSubmitInteractionData{CustomID: "ticket_modal_1:" + questionVersion("old title", []QuestionData{{Label: "old question"}})}
	HandleModalSubmit(s, i)
	if !strings.Contains(message, "Panel questions changed") {
		t.Fatalf("old form not rejected: %s", message)
	}
	if openingQueue.hasPending("1:user") {
		t.Fatal("old form queued")
	}
}

func TestClosureWhileWaitingForDiscordBucket(t *testing.T) {
	d := &surgeDB{}
	installSurgeDB(t, d)
	s, _ := discordgo.New("Bot test")
	var creates atomic.Int32
	s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if err := r.Context().Err(); err != nil {
			return nil, err
		}
		creates.Add(1)
		return response(200, `{"id":"created"}`), nil
	})}
	// Hold the actual DiscordGo bucket: closure happens after live validation but before send.
	bucket := s.Ratelimiter.GetBucket(discordgo.EndpointGuildChannels("1"))
	bucket.Lock()
	q := newTicketQueue(1, nil, func(*ticketJob, string) {})
	j := jobFor("user")
	j.session = s
	j.interaction = interaction()
	q.enqueue(j)
	done := make(chan error, 1)
	go func() { done <- openTicket(context.Background(), j) }()
	eventually(t, func() bool { return d.reads.Load() >= 6 })
	q.reconcile(1, nil, true)
	bucket.Unlock()
	select {
	case err := <-done:
		if !errors.Is(err, errJobCanceled) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("bucket request did not return")
	}
	q.cancel()
	if creates.Load() != 0 {
		t.Fatal("request sent after closure")
	}
}

func TestCreationFailureHandlingAndTicketLimit(t *testing.T) {
	for _, scenario := range []string{"insert failure", "cleanup failure", "Discord failure", "ticket limit", "cache miss exists", "unknown channel"} {
		t.Run(scenario, func(t *testing.T) {
			d := &surgeDB{}
			if strings.Contains(scenario, "failure") && scenario != "Discord failure" {
				d.insertErr = errors.New("insert failed")
			}
			if scenario == "ticket limit" || scenario == "cache miss exists" || scenario == "unknown channel" {
				d.maxTickets = 1
				d.count = 1
				d.activeChannels = []string{"existing"}
			}
			installSurgeDB(t, d)
			var creates, deletes atomic.Int32
			s, _ := discordgo.New("Bot test")
			s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/channels"):
					creates.Add(1)
					if scenario == "Discord failure" {
						return response(502, `{}`), nil
					}
					return response(200, `{"id":"created"}`), nil
				case r.Method == http.MethodDelete:
					deletes.Add(1)
					if scenario == "cleanup failure" {
						return response(403, `{"code":50013}`), nil
					}
					return response(200, `{"id":"created"}`), nil
				case r.Method == http.MethodGet:
					if scenario == "unknown channel" {
						return response(404, `{"code":10003,"message":"Unknown Channel"}`), nil
					}
					return response(200, `{"id":"existing"}`), nil
				}
				return response(200, `{}`), nil
			})}
			q := newTicketQueue(1, nil, func(*ticketJob, string) {})
			j := jobFor("user")
			j.session = s
			j.interaction = interaction()
			q.enqueue(j)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := openTicket(ctx, j)
			q.mu.Lock()
			if j.timer != nil {
				j.timer.Stop()
			}
			j.cancel()
			q.mu.Unlock()
			q.cancel()
			switch scenario {
			case "insert failure", "cleanup failure":
				if err == nil || creates.Load() != 1 || deletes.Load() != 1 {
					t.Fatalf("err=%v create=%d delete=%d", err, creates.Load(), deletes.Load())
				}
			case "Discord failure":
				if err == nil || creates.Load() != 1 || deletes.Load() != 0 {
					t.Fatal("creation retried or missing error")
				}
			case "ticket limit", "cache miss exists":
				if !errors.Is(err, errMaxTickets) || creates.Load() != 0 || d.deletes != 0 {
					t.Fatal("limit/cache miss bypassed tracking")
				}
			case "unknown channel":
				if err != nil || creates.Load() != 1 || d.deletes != 1 {
					t.Fatalf("confirmed stale record not pruned: %v", err)
				}
			}
		})
	}
}
