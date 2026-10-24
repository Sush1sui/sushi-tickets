package tickets

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/bwmarrin/discordgo"
)

const (
	queueCapacity         = 500
	queueLifetime         = 10 * time.Minute
	windowCanceledMessage = "The ticket window closed before your request could be processed."
)

var (
	errQueueFull    = errors.New("ticket queue full")
	errPending      = errors.New("ticket request already pending")
	errQueueStopped = errors.New("ticket queue stopped")
	errJobCanceled  = errors.New("ticket request canceled")
)

type ticketJob struct {
	queue       *ticketQueue
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	panelID     int32
	qna         []QnA
	window      db.PanelScheduleWindow
	queuedAt    time.Time
	key         string
	ctx         context.Context
	cancel      context.CancelFunc
	timer       *time.Timer
	started     bool // Channel creation has been sent to Discord; closure may no longer cancel it.
	canceled    bool
}

type ticketNotice struct {
	job     *ticketJob
	message string
}

type ticketQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	waiting    []*ticketJob
	pending    map[string]*ticketJob
	capacity   int
	stopped    bool
	started    bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	drained    chan struct{}
	drainOnce  sync.Once
	duplicates uint64
	rejected   uint64
	notices    chan ticketNotice
	run        func(context.Context, *ticketJob) error
	notify     func(*ticketJob, string)
}

func newTicketQueue(capacity int, run func(context.Context, *ticketJob) error, notify func(*ticketJob, string)) *ticketQueue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &ticketQueue{capacity: capacity, pending: make(map[string]*ticketJob), ctx: ctx, cancel: cancel, run: run, notify: notify, notices: make(chan ticketNotice, 2*capacity+4), drained: make(chan struct{})}
	q.cond = sync.NewCond(&q.mu)
	return q
}

var openingQueue = newTicketQueue(queueCapacity, openTicket, func(j *ticketJob, message string) {
	editEphemeral(j.session, j.interaction, message)
})

func StartOpeningQueue() { openingQueue.start(2) }

func StopOpeningQueue(ctx context.Context) { openingQueue.stop(ctx) }

func (q *ticketQueue) start(workers int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.started || q.stopped {
		return
	}
	q.started = true
	for n := 0; n < workers; n++ {
		q.wg.Add(1)
		go q.worker()
	}
	// Notifications never occupy creation workers or block configuration writes.
	for n := 0; n < 2; n++ {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			for {
				select {
				case notice := <-q.notices:
					q.notify(notice.job, notice.message)
				case <-q.ctx.Done():
					return
				}
			}
		}()
	}
}

func (q *ticketQueue) hasPending(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pending[key] != nil
}

func (q *ticketQueue) enqueue(j *ticketJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.stopped {
		return errQueueStopped
	}
	if q.pending[j.key] != nil {
		q.noteDuplicateLocked()
		return errPending
	}
	if len(q.waiting) >= q.capacity {
		q.rejected++
		if q.rejected == 1 || q.rejected%100 == 0 {
			log.Printf("ticket_queue rejected count=%d depth=%d", q.rejected, len(q.waiting))
		}
		return errQueueFull
	}
	j.interaction = openingInteraction(j.interaction)
	j.queue = q
	j.ctx, j.cancel = context.WithCancel(q.ctx)
	q.pending[j.key] = j
	q.waiting = append(q.waiting, j)
	q.armLocked(j)
	q.cond.Signal()
	log.Printf("ticket_queue admitted depth=%d", len(q.waiting))
	return nil
}

func (q *ticketQueue) armLocked(j *ticketJob) {
	deadline := j.queuedAt.Add(queueLifetime)
	if j.window.ID != 0 && time.Unix(j.window.EndTime, 0).Before(deadline) {
		deadline = time.Unix(j.window.EndTime, 0)
	}
	if j.timer != nil {
		j.timer.Reset(time.Until(deadline))
		return
	}
	j.timer = time.AfterFunc(time.Until(deadline), func() {
		panelPolicyMu.RLock()
		defer panelPolicyMu.RUnlock()
		q.mu.Lock()
		defer q.mu.Unlock()
		if q.pending[j.key] != j {
			return
		}
		// A stopped timer may already be executing after a schedule extension.
		if !j.started && !j.canceled {
			if message := expiredJob(j, time.Now()); message != "" {
				q.cancelLocked(j, message)
			}
		}
	})
}

func expiredJob(j *ticketJob, now time.Time) string {
	if j.window.ID != 0 && (now.Unix() >= j.window.EndTime || now.Unix() < j.window.StartTime || j.window.ClosedEarly) {
		return windowCanceledMessage
	}
	if !now.Before(j.queuedAt.Add(queueLifetime)) {
		return "Your ticket request expired. Please submit again while the window is open."
	}
	return ""
}

func (q *ticketQueue) noticeLocked(j *ticketJob, message string) {
	select {
	case q.notices <- ticketNotice{j, message}:
	default:
		// shortcut: notification backlog is bounded; persistent Discord outages may lose status updates.
		log.Printf("ticket_queue notification_backlog panel=%d", j.panelID)
	}
}

func (q *ticketQueue) cancelLocked(j *ticketJob, message string) {
	if j.started || j.canceled {
		return
	}
	j.canceled = true
	j.cancel()
	if j.timer != nil {
		j.timer.Stop()
	}
	// Preparing jobs retain their pending key until their worker exits.
	for n, queued := range q.waiting {
		if queued == j {
			copy(q.waiting[n:], q.waiting[n+1:])
			q.waiting[len(q.waiting)-1] = nil
			q.waiting = q.waiting[:len(q.waiting)-1]
			delete(q.pending, j.key)
			j.qna = nil
			break
		}
	}
	q.noticeLocked(j, message)
	log.Printf("ticket_queue canceled panel=%d depth=%d", j.panelID, len(q.waiting))
}

// reconcile is called under panelPolicyMu after a configuration write commits.
func (q *ticketQueue) reconcile(panelID int32, windows []db.PanelScheduleWindow, unavailable bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, j := range q.pending {
		if j.panelID != panelID || j.started || j.canceled {
			continue
		}
		valid := !unavailable && j.window.ID == 0 && len(windows) == 0
		for _, w := range windows {
			if w.ID == j.window.ID && !w.ClosedEarly && w.StartTime <= time.Now().Unix() && w.EndTime > time.Now().Unix() {
				j.window = w
				valid = !unavailable
				break
			}
		}
		if !valid {
			q.cancelLocked(j, windowCanceledMessage)
		} else {
			q.armLocked(j)
		}
	}
}

// markStarted runs after DiscordGo acquires its REST bucket, immediately before sending.
func (q *ticketQueue) markStarted(j *ticketJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.stopped || j.canceled || j.ctx.Err() != nil {
		return errJobCanceled
	}
	if message := expiredJob(j, time.Now()); message != "" {
		q.cancelLocked(j, message)
		return errJobCanceled
	}
	j.started = true
	j.timer.Stop()
	return nil
}

func (q *ticketQueue) worker() {
	defer q.wg.Done()
	for {
		q.mu.Lock()
		for len(q.waiting) == 0 && !q.stopped {
			q.cond.Wait()
		}
		if q.stopped {
			q.mu.Unlock()
			return
		}
		j := q.waiting[0]
		q.waiting[0] = nil
		q.waiting = q.waiting[1:]
		q.mu.Unlock()
		startedAt := time.Now()
		ctx, cancel := context.WithTimeout(j.ctx, 45*time.Second)
		err := q.run(ctx, j)
		cancel()
		q.mu.Lock()
		j.timer.Stop()
		j.cancel()
		delete(q.pending, j.key)
		j.qna = nil
		q.signalDrainedLocked()
		if !j.canceled {
			if err != nil {
				q.noticeLocked(j, ticketFailureMessage(err))
			}
		}
		log.Printf("ticket_queue finished failed=%t wait_ms=%d work_ms=%d depth=%d", err != nil, startedAt.Sub(j.queuedAt).Milliseconds(), time.Since(startedAt).Milliseconds(), len(q.waiting))
		q.mu.Unlock()
	}
}

func (q *ticketQueue) stop(ctx context.Context) {
	q.mu.Lock()
	q.stopped = true
	for _, j := range q.pending {
		q.cancelLocked(j, "The bot is shutting down. Please try again shortly.")
	}
	q.cond.Broadcast()
	q.signalDrainedLocked()
	q.mu.Unlock()
	// Running requests finish within the caller's shutdown budget.
	select {
	case <-q.drained:
	case <-ctx.Done():
	}
	q.cancel()
	workersDone := make(chan struct{})
	go func() { q.wg.Wait(); close(workersDone) }()
	select {
	case <-workersDone:
	case <-ctx.Done():
	}
}

func (q *ticketQueue) signalDrainedLocked() {
	if q.stopped && len(q.pending) == 0 {
		q.drainOnce.Do(func() { close(q.drained) })
	}
}

func (q *ticketQueue) noteDuplicateLocked() {
	q.duplicates++
	// Keep spam from turning diagnostic logs into a disk/CPU bottleneck.
	if q.duplicates == 1 || q.duplicates%100 == 0 {
		log.Printf("ticket_queue duplicate count=%d", q.duplicates)
	}
}

func (q *ticketQueue) noteDuplicate() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.noteDuplicateLocked()
}

// Ticket creation and response edits only need identity, naming, and response IDs.
// Do not retain the source message, form components, roles, or resolved guild data.
func openingInteraction(source *discordgo.InteractionCreate) *discordgo.InteractionCreate {
	if source == nil || source.Interaction == nil {
		return source
	}
	compact := &discordgo.Interaction{ID: source.ID, AppID: source.AppID, Token: source.Token, GuildID: source.GuildID, ChannelID: source.ChannelID, Type: source.Type, Version: source.Version}
	if source.Member != nil && source.Member.User != nil {
		compact.Member = &discordgo.Member{User: &discordgo.User{ID: source.Member.User.ID, Username: source.Member.User.Username}}
	}
	return &discordgo.InteractionCreate{Interaction: compact}
}

func ticketFailureMessage(err error) string {
	switch {
	case errors.Is(err, errPending):
		return "You already have a ticket request pending."
	case errors.Is(err, errQueueFull):
		return "Ticket queue is full. Please try again."
	case errors.Is(err, errQueueStopped):
		return "The bot is shutting down. Please try again shortly."
	case errors.Is(err, errMaxTickets):
		return "You reached max open tickets. Please close existing tickets."
	case errors.Is(err, errJobCanceled):
		return windowCanceledMessage
	}
	var closed *errWindowClosed
	if errors.As(err, &closed) {
		return windowClosedMessage(closed.NextStartTime)
	}
	return "Failed to create ticket. Please try again later."
}
