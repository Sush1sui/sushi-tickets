package tickets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	autoCloseTick         = 2 * time.Second
	autoCloseEvalEvery    = 30
	autoCloseDeleteBudget = 20 * time.Second
)

var nowFn = time.Now

type autoCloseSettings struct {
	active       bool
	closeOnLeave bool
	noResponse   time.Duration
	inactivity   time.Duration
}

type ticketTracker struct {
	guildID       int64
	openerID      string
	channelID     string
	createdAt     time.Time
	lastActivity  time.Time
	responded     bool
	responseKnown bool
	observedSince time.Time
	uncertain     bool
	left          bool
	closing       bool
	capture       bool
	createdHere   bool
	transcriptID  int32
	transcriptKey string
	totalMessages int
	totalAttach   int
	totalEmbeds   int
	cleanup       *cleanupMarker
}

var autoClose = struct {
	sync.Mutex
	session  *discordgo.Session
	trackers map[string]*ticketTracker
	configs  map[int64]autoCloseSettings
	captures map[int64]bool
	ready    bool
	hydrated bool
	readyAt  time.Time
	wake     chan struct{}
	started  bool
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}{
	trackers: make(map[string]*ticketTracker),
	configs:  make(map[int64]autoCloseSettings),
	captures: make(map[int64]bool),
	wake:     make(chan struct{}, 1),
}

func loadSettings(serverID int64) autoCloseSettings {
	if queries == nil {
		return autoCloseSettings{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cfg, err := queries.GetAutoCloseConfig(ctx, serverID)
	if err != nil {
		return autoCloseSettings{}
	}
	var settings autoCloseSettings
	settings.active = cfg.IsActive
	settings.closeOnLeave = cfg.CloseOnUserLeave
	if cfg.CloseSinceOpenWithNoResponseMins.Valid && cfg.CloseSinceOpenWithNoResponseMins.Int32 > 0 {
		settings.noResponse = time.Duration(cfg.CloseSinceOpenWithNoResponseMins.Int32) * time.Minute
	}
	if cfg.CloseSinceLastMessageMins.Valid && cfg.CloseSinceLastMessageMins.Int32 > 0 {
		settings.inactivity = time.Duration(cfg.CloseSinceLastMessageMins.Int32) * time.Minute
	}
	return settings
}

func loadCapture(serverID int64) bool {
	if queries == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cfg, err := queries.GetServerConfig(ctx, serverID)
	if err != nil {
		return false
	}
	return cfg.TicketTranscriptCid.Valid && cfg.TicketTranscriptCid.String != ""
}

func StartAutoClose(s *discordgo.Session) {
	autoClose.Lock()
	if autoClose.started {
		autoClose.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	autoClose.session = s
	autoClose.cancel = cancel
	autoClose.started = true
	autoClose.wg.Add(1)
	autoClose.Unlock()

	go func() {
		defer autoClose.wg.Done()
		ticker := time.NewTicker(autoCloseTick)
		defer ticker.Stop()
		ticks := 0
		for {
			select {
			case <-ctx.Done():
				flushDirtySpools(true)
				return
			case <-ticker.C:
				flushDirtySpools(false)
				ticks++
				if ticks%autoCloseEvalEvery == 0 {
					evaluateAutoClose()
				}
			case <-autoClose.wake:
				flushDirtySpools(false)
				evaluateAutoClose()
			}
		}
	}()
}

func StopAutoClose() {
	autoClose.Lock()
	cancel := autoClose.cancel
	autoClose.cancel = nil
	autoClose.started = false
	autoClose.ready = false
	autoClose.hydrated = false
	autoClose.Unlock()
	if cancel != nil {
		cancel()
		autoClose.wg.Wait()
	}
	flushDirtySpools(true)
}

func wakeAutoClose() {
	select {
	case autoClose.wake <- struct{}{}:
	default:
	}
}

func flushDirtySpools(syncWrite bool) {
	for _, channel := range dirtySpoolChannels() {
		if err := spoolFlush(channel, syncWrite); err != nil {
			log.Printf("auto_close spool_flush channel=%s failed=%t", channel, true)
		}
	}
}

// TrackNewTicket starts in-memory tracking and, when configured, local capture
// immediately after the active_ticket row is committed.
func TrackNewTicket(serverID int64, openerID, channelID string, createdAt time.Time, capture bool) {
	autoClose.Lock()
	autoClose.trackers[channelID] = &ticketTracker{
		guildID:       serverID,
		openerID:      openerID,
		channelID:     channelID,
		createdAt:     createdAt,
		lastActivity:  createdAt,
		responseKnown: true,
		observedSince: createdAt,
		capture:       capture,
		createdHere:   true,
	}
	autoClose.Unlock()
	if capture {
		if err := spoolStart(serverID, channelID, openerID, createdAt.Unix(), createdAt.Unix(), false, ""); err != nil {
			log.Printf("auto_close spool_start channel=%s failed: %v", channelID, err)
		}
	}
}

func HydrateAutoClose(s *discordgo.Session, _ []*discordgo.Guild) {
	if s == nil || queries == nil {
		return
	}
	autoIDs, autoErr := queries.ListAutoCloseServers(context.Background())
	capIDs, capErr := queries.ListTranscriptEnabledServers(context.Background())

	settings := make(map[int64]autoCloseSettings)
	captures := make(map[int64]bool)
	if autoErr == nil {
		for _, id := range autoIDs {
			settings[id] = loadSettings(id)
		}
	}
	if capErr == nil {
		for _, id := range capIDs {
			captures[id] = true
		}
	}

	autoClose.Lock()
	autoClose.session = s
	autoClose.readyAt = nowFn()
	autoClose.ready = true
	autoClose.hydrated = false
	for id, cfg := range settings {
		autoClose.configs[id] = cfg
	}
	for id := range captures {
		autoClose.captures[id] = true
	}
	autoClose.Unlock()

	relevant := make(map[int64]struct{}, len(settings)+len(captures))
	for id := range settings {
		relevant[id] = struct{}{}
	}
	for id := range captures {
		relevant[id] = struct{}{}
	}
	for id := range relevant {
		HydrateAutoCloseGuild(s, strconv.FormatInt(id, 10))
	}

	autoClose.Lock()
	autoClose.hydrated = true
	autoClose.Unlock()
	recoverCleanupMarkers(s)
}

func HydrateAutoCloseGuild(s *discordgo.Session, guildID string) {
	if s == nil || queries == nil || guildID == "" {
		return
	}
	serverID, err := strconv.ParseInt(guildID, 10, 64)
	if err != nil {
		return
	}
	settings := loadSettings(serverID)
	capture := loadCapture(serverID)
	if !settings.active && !capture {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := queries.ListActiveTicketsByServer(ctx, serverID)
	if err != nil {
		log.Printf("auto_close hydrate guild=%d failed: %v", serverID, err)
		return
	}

	autoClose.Lock()
	autoClose.configs[serverID] = settings
	if capture {
		autoClose.captures[serverID] = true
	}
	autoClose.session = s
	autoClose.Unlock()

	now := nowFn()
	for _, row := range rows {
		hydrateTicket(s, serverID, row, capture, now)
	}
}

func hydrateTicket(s *discordgo.Session, serverID int64, row db.ActiveTicketRow, capture bool, now time.Time) {
	channelID := row.ChannelID
	created := time.Unix(row.CreatedAt, 0)

	meta, metaErr := spoolScanMetadata(serverID, channelID)
	if metaErr != nil {
		log.Printf("auto_close spool_scan channel=%s failed: %v", channelID, metaErr)
	}

	responded := meta.responded
	if !responded {
		responded = cachedHumanMessage(s, channelID)
	}

	autoClose.Lock()
	tracker, exists := autoClose.trackers[channelID]
	if exists {
		if tracker.closing {
			autoClose.Unlock()
			return
		}
		tracker.capture = tracker.capture || capture || meta.exists
		tracker.observedSince = now
		tracker.uncertain = true
		if !tracker.responded {
			tracker.responseKnown = false
			tracker.responded = responded
			if responded {
				tracker.responseKnown = true
			}
		}
		if meta.lastActivity.After(tracker.lastActivity) {
			tracker.lastActivity = meta.lastActivity
		}
	} else {
		tracker = &ticketTracker{
			guildID:       serverID,
			openerID:      row.UserID,
			channelID:     channelID,
			createdAt:     created,
			lastActivity:  meta.lastActivity,
			responded:     responded,
			responseKnown: responded,
			observedSince: now,
			uncertain:     true,
			capture:       capture || meta.exists,
		}
		autoClose.trackers[channelID] = tracker
	}
	autoClose.Unlock()

	if tracker.capture {
		if !meta.exists {
			if err := spoolStart(serverID, channelID, tracker.openerID, row.CreatedAt, now.Unix(), true, "capture_resumed_after_restart"); err != nil {
				log.Printf("auto_close spool_start channel=%s failed: %v", channelID, err)
				return
			}
		} else {
			_ = spoolAppend(serverID, channelID, spoolEvent{T: "g", TS: now.Unix(), Reason: "gateway_restart"})
		}
	}
}

func cachedHumanMessage(s *discordgo.Session, channelID string) bool {
	if s == nil || s.State == nil {
		return false
	}
	channel, err := s.State.Channel(channelID)
	if err != nil || channel == nil {
		return false
	}
	for _, msg := range channel.Messages {
		if msg == nil || msg.Author == nil {
			continue
		}
		if !msg.Author.Bot && msg.WebhookID == "" {
			return true
		}
	}
	return false
}

// RefreshGuildConfig re-reads settings after a dashboard save. Capture is
// never torn down for open tickets; only the auto-close eligibility changes.
func RefreshGuildConfig(serverID int64) {
	settings := loadSettings(serverID)
	capture := loadCapture(serverID)
	autoClose.Lock()
	autoClose.configs[serverID] = settings
	wasCapture := autoClose.captures[serverID]
	if capture {
		autoClose.captures[serverID] = true
	}
	session := autoClose.session
	autoClose.Unlock()

	if session == nil || queries == nil {
		return
	}
	if !settings.active && !capture {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := queries.ListActiveTicketsByServer(ctx, serverID)
	if err != nil {
		return
	}
	now := nowFn()
	for _, row := range rows {
		if capture && !wasCapture {
			if err := spoolStart(serverID, row.ChannelID, row.UserID, row.CreatedAt, now.Unix(), true, "capture_started_after_ticket_creation"); err != nil {
				log.Printf("auto_close spool_start channel=%s failed: %v", row.ChannelID, err)
			}
		}
		hydrateTicket(session, serverID, row, capture, now)
	}
}

func getTracker(channelID string) *ticketTracker {
	autoClose.Lock()
	defer autoClose.Unlock()
	return autoClose.trackers[channelID]
}

func HandleAutoCloseMessage(_ *discordgo.Session, event *discordgo.MessageCreate) {
	if event == nil || event.Message == nil || event.Author == nil {
		return
	}
	msg := event.Message
	human := !msg.Author.Bot && msg.WebhookID == ""

	var capture bool
	autoClose.Lock()
	tracker := autoClose.trackers[msg.ChannelID]
	if tracker != nil {
		ts := msg.Timestamp
		if ts.IsZero() {
			ts = nowFn()
		}
		if human {
			tracker.responded = true
			tracker.responseKnown = true
			if ts.After(tracker.lastActivity) {
				tracker.lastActivity = ts
			}
		}
		capture = tracker.capture
	}
	autoClose.Unlock()

	if tracker == nil || !capture {
		return
	}
	if err := spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(msg, "create")}); err != nil {
		log.Printf("auto_close spool_append channel=%s failed: %v", tracker.channelID, err)
	}
}

func HandleAutoCloseMessageUpdate(_ *discordgo.Session, event *discordgo.MessageUpdate) {
	if event == nil || event.Message == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "m", Kind: "update", M: spoolMessageFromDiscord(event.Message, "update")})
}

func HandleAutoCloseMessageDelete(_ *discordgo.Session, event *discordgo.MessageDelete) {
	if event == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture || event.ID == "" {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "d", ID: event.ID})
}

func HandleAutoCloseMessageDeleteBulk(_ *discordgo.Session, event *discordgo.MessageDeleteBulk) {
	if event == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture || len(event.Messages) == 0 {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "x", IDs: event.Messages})
}

func HandleAutoCloseReactionAdd(_ *discordgo.Session, event *discordgo.MessageReactionAdd) {
	if event == nil || event.MessageReaction == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "ra", MID: event.MessageID, Emoji: event.Emoji.Name, User: event.UserID})
}

func HandleAutoCloseReactionRemove(_ *discordgo.Session, event *discordgo.MessageReactionRemove) {
	if event == nil || event.MessageReaction == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "rr", MID: event.MessageID, Emoji: event.Emoji.Name, User: event.UserID})
}

func HandleAutoCloseReactionRemoveAll(_ *discordgo.Session, event *discordgo.MessageReactionRemoveAll) {
	if event == nil || event.MessageReaction == nil {
		return
	}
	tracker := getTracker(event.ChannelID)
	if tracker == nil || !tracker.capture {
		return
	}
	_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "rc", MID: event.MessageID})
}

func HandleAutoCloseMemberRemove(_ *discordgo.Session, event *discordgo.GuildMemberRemove) {
	if event == nil || event.Member == nil || event.User == nil || event.GuildID == "" {
		return
	}
	guildID, err := strconv.ParseInt(event.GuildID, 10, 64)
	if err != nil {
		return
	}
	matched := false
	autoClose.Lock()
	for _, tracker := range autoClose.trackers {
		if tracker.guildID == guildID && tracker.openerID == event.User.ID {
			tracker.left = true
			matched = true
		}
	}
	autoClose.Unlock()
	if matched {
		wakeAutoClose()
	}
}

func HandleGatewayDisconnect(_ *discordgo.Session, _ *discordgo.Disconnect) {
	autoClose.Lock()
	autoClose.ready = false
	now := nowFn()
	captures := make([]*ticketTracker, 0, len(autoClose.trackers))
	for _, tracker := range autoClose.trackers {
		if !tracker.closing {
			tracker.uncertain = true
			if !tracker.responded {
				tracker.responseKnown = false
			}
		}
		if tracker.capture {
			captures = append(captures, tracker)
		}
	}
	autoClose.Unlock()
	for _, tracker := range captures {
		_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "g", TS: now.Unix(), Reason: "gateway_disconnect"})
	}
}

func HandleGatewayResumed(s *discordgo.Session, _ *discordgo.Resumed) {
	autoClose.Lock()
	autoClose.ready = true
	autoClose.readyAt = nowFn()
	now := autoClose.readyAt
	trackers := make([]*ticketTracker, 0, len(autoClose.trackers))
	for _, tracker := range autoClose.trackers {
		if tracker.closing {
			continue
		}
		tracker.observedSince = now
		tracker.uncertain = true
		if !tracker.responded {
			tracker.responseKnown = false
		}
		trackers = append(trackers, tracker)
	}
	autoClose.Unlock()

	// A cached human message proves a response even when the cache is bounded.
	for _, tracker := range trackers {
		if tracker.capture {
			_ = spoolAppend(tracker.guildID, tracker.channelID, spoolEvent{T: "g", TS: now.Unix(), Reason: "gateway_resumed"})
		}
		if !tracker.responded && cachedHumanMessage(s, tracker.channelID) {
			autoClose.Lock()
			tracker.responded = true
			tracker.responseKnown = true
			autoClose.Unlock()
		}
	}
}

func dueReason(t *ticketTracker, cfg autoCloseSettings, now time.Time) string {
	if !cfg.active {
		return ""
	}
	if cfg.closeOnLeave && t.left {
		return "creator_left"
	}
	if cfg.noResponse > 0 && t.responseKnown && !t.responded {
		base := t.createdAt
		if t.observedSince.After(base) {
			base = t.observedSince
		}
		if !now.Before(base.Add(cfg.noResponse)) {
			return "no_response"
		}
	}
	if cfg.inactivity > 0 {
		base := t.lastActivity
		if t.observedSince.After(base) {
			base = t.observedSince
		}
		if !now.Before(base.Add(cfg.inactivity)) {
			return "inactivity"
		}
	}
	return ""
}

func evaluateAutoClose() {
	retryDeletedTickets()
	autoClose.Lock()
	if !autoClose.ready || !autoClose.hydrated {
		autoClose.Unlock()
		return
	}
	now := nowFn()
	candidates := make([]*ticketTracker, 0)
	for _, tracker := range autoClose.trackers {
		if tracker.closing || tracker.cleanup != nil {
			continue
		}
		cfg, ok := autoClose.configs[tracker.guildID]
		if !ok || !cfg.active {
			continue
		}
		if dueReason(tracker, cfg, now) != "" {
			candidates = append(candidates, tracker)
		}
	}
	autoClose.Unlock()

	for _, tracker := range candidates {
		attemptAutoClose(tracker)
	}
}

func attemptAutoClose(tracker *ticketTracker) {
	autoClose.Lock()
	if tracker.closing || tracker.cleanup != nil {
		autoClose.Unlock()
		return
	}
	cfg, ok := autoClose.configs[tracker.guildID]
	ready := autoClose.ready && autoClose.hydrated
	if !ok || !ready || !cfg.active || dueReason(tracker, cfg, nowFn()) == "" {
		autoClose.Unlock()
		return
	}
	tracker.closing = true
	autoClose.Unlock()

	deleted, err := closeTracker(tracker, "auto", autoCloseID())
	if err != nil {
		log.Printf("auto_close channel=%s deleted=%t failed: %v", tracker.channelID, deleted, err)
	}
	if !deleted || err != nil {
		autoClose.Lock()
		tracker.closing = false
		autoClose.Unlock()
	}
}

func autoCloseID() string {
	autoClose.Lock()
	defer autoClose.Unlock()
	if autoClose.session != nil && autoClose.session.State != nil && autoClose.session.State.User != nil {
		return autoClose.session.State.User.ID
	}
	return "auto-close"
}

// closeTracker persists the configured transcript, writes recovery state,
// deletes the channel, then finalizes the transcript with late events.
func closeTracker(tracker *ticketTracker, mode, closedBy string) (bool, error) {
	seq := spoolSeq(tracker.channelID)
	if tracker.capture {
		if err := spoolFlush(tracker.channelID, true); err != nil {
			return false, err
		}
	}
	rowID, err := persistTrackerTranscript(tracker, closedBy)
	if err != nil {
		return false, err
	}
	marker := cleanupMarker{
		Server: tracker.guildID, Channel: tracker.channelID, Mode: mode,
		State: "pending", RowID: rowID, ClosedBy: closedBy,
		Opener: tracker.openerID, OpenedAt: tracker.createdAt.Unix(),
	}
	if err := writeMarker(marker); err != nil {
		return false, err
	}

	if mode == "auto" {
		autoClose.Lock()
		cfg, ok := autoClose.configs[tracker.guildID]
		eligible := ok && autoClose.ready && autoClose.hydrated && dueReason(tracker, cfg, nowFn()) != ""
		autoClose.Unlock()
		if !eligible {
			deleteMarker(tracker.guildID, tracker.channelID)
			return false, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), autoCloseDeleteBudget)
	defer cancel()
	autoClose.Lock()
	session := autoClose.session
	autoClose.Unlock()
	if session == nil {
		deleteMarker(tracker.guildID, tracker.channelID)
		return false, errors.New("session unavailable")
	}
	_, err = session.ChannelDelete(tracker.channelID, discordgo.WithContext(ctx), discordgo.WithRestRetries(0), discordgo.WithRetryOnRatelimit(false))
	if err != nil && !isUnknownChannel(err) {
		deleteMarker(tracker.guildID, tracker.channelID)
		return false, err
	}
	marker.Deleted = true
	autoClose.Lock()
	tracker.cleanup = &marker
	autoClose.Unlock()
	if err := writeMarker(marker); err != nil {
		return true, err
	}
	return true, finalizeDeletedTicket(tracker, marker, seq)
}

func finalizeDeletedTicket(tracker *ticketTracker, marker cleanupMarker, seq uint64) error {
	// Finalize with any events observed during deletion. On failure the marker
	// and spool stay for recovery. Capture remains active until the channel is gone.
	for attempt := 0; attempt < 2; attempt++ {
		current := spoolSeq(tracker.channelID)
		if current == seq {
			break
		}
		if err := spoolFlush(tracker.channelID, true); err != nil {
			return err
		}
		rowID, err := persistTrackerTranscript(tracker, marker.ClosedBy)
		if err != nil {
			return err
		}
		seq = current
		marker.RowID = rowID
	}
	if spoolSeq(tracker.channelID) != seq {
		return errors.New("transcript still receiving events; finalization pending")
	}
	marker.State = "confirmed"
	if err := writeMarker(marker); err != nil {
		return err
	}
	if !marker.Logged {
		sendTranscriptLog(tracker, marker.ClosedBy, marker.RowID)
		marker.Logged = true
		if err := writeMarker(marker); err != nil {
			return err
		}
	}
	autoClose.Lock()
	tracker.cleanup = &marker
	autoClose.Unlock()
	return cleanupConfirmed(marker)
}

func cleanupConfirmed(marker cleanupMarker) error {
	if queries == nil {
		return errors.New("database unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := queries.DeleteActiveTicketByChannel(ctx, marker.Server, marker.Channel); err != nil {
		log.Printf("auto_close row_cleanup channel=%s failed: %v", marker.Channel, err)
		return err
	}
	spoolDelete(marker.Channel)
	deleteMarker(marker.Server, marker.Channel)
	autoClose.Lock()
	delete(autoClose.trackers, marker.Channel)
	autoClose.Unlock()
	return nil
}

// Retry only known deletions; live tickets still use the normal eligibility checks.
// The in-memory scan avoids reading marker files on idle ticks.
func retryDeletedTickets() {
	autoClose.Lock()
	var pending []*ticketTracker
	for _, tracker := range autoClose.trackers {
		if tracker.cleanup != nil && !tracker.closing {
			tracker.closing = true
			pending = append(pending, tracker)
		}
	}
	autoClose.Unlock()
	for _, tracker := range pending {
		marker := *tracker.cleanup
		var err error
		if marker.State == "confirmed" {
			err = cleanupConfirmed(marker)
		} else {
			err = finalizeDeletedTicket(tracker, marker, ^uint64(0))
		}
		if err != nil {
			log.Printf("ticket cleanup channel=%s failed: %v", tracker.channelID, err)
			autoClose.Lock()
			tracker.closing = false
			autoClose.Unlock()
		}
	}
}

func persistTrackerTranscript(tracker *ticketTracker, closedBy string) (int32, error) {
	if queries == nil {
		return 0, errors.New("database unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverConfig, err := queries.GetServerConfig(ctx, tracker.guildID)
	if err != nil {
		return 0, err
	}
	if !serverConfig.TicketTranscriptCid.Valid || serverConfig.TicketTranscriptCid.String == "" {
		return 0, nil
	}

	replay, err := spoolReplayTicket(tracker.guildID, tracker.channelID)
	if err != nil {
		return 0, err
	}
	if !replay.exists {
		now := nowFn()
		if err := spoolStart(tracker.guildID, tracker.channelID, tracker.openerID, tracker.createdAt.Unix(), now.Unix(), true, "capture_missing_at_close"); err != nil {
			return 0, err
		}
		replay, err = spoolReplayTicket(tracker.guildID, tracker.channelID)
		if err != nil {
			return 0, err
		}
	}

	content, totalAttachments, totalEmbeds, participants := buildTranscriptContent(replay.messages)
	openedAt := tracker.createdAt.Unix()
	closedAt := nowFn().Unix()
	captureStart := replay.captureStart
	if captureStart.IsZero() {
		captureStart = tracker.createdAt
	}
	complete := tracker.createdHere && !replay.gap && !replay.uncertain && !spoolFailed(tracker.channelID)
	note := ""
	if !complete {
		note = "Recording began after this ticket opened or was interrupted. Messages before recording started, or sent while the bot was offline, were never observed and cannot be recovered. Buffered events not yet written can also be lost if the process crashes."
	}
	username := usernameOrID(trackerWeSession(), strconv.FormatInt(tracker.guildID, 10), tracker.openerID)
	payload := transcriptPayload{
		TicketID: tracker.channelID,
		Username: username,
		UserID:   tracker.openerID,
		Messages: content,
		Metadata: transcriptMetadata{
			TicketOpenedAt:   time.Unix(openedAt, 0).UTC().Format(time.RFC3339),
			TicketClosedAt:   time.Unix(closedAt, 0).UTC().Format(time.RFC3339),
			ClosedBy:         transcriptClosedBy{ID: closedBy, Username: usernameOrID(trackerWeSession(), strconv.FormatInt(tracker.guildID, 10), closedBy)},
			TotalMessages:    len(content),
			TotalAttachments: totalAttachments,
			TotalEmbeds:      totalEmbeds,
			Participants:     participants,
			HistoryComplete:  complete,
			CaptureStart:     captureStart.UTC().Format(time.RFC3339),
			GapDetected:      replay.gap,
			HistoryNote:      note,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	key := fmt.Sprintf("transcripts/%d/tickets/%s.json", tracker.guildID, tracker.channelID)
	if storageClient == nil {
		return 0, errors.New("storage unavailable")
	}
	if err := storageClient.UploadTranscript(ctx, key, data); err != nil {
		return 0, err
	}

	params := db.UpdateTranscriptByIDParams{
		Username:         pgtype.Text{String: username, Valid: username != ""},
		UserID:           pgtype.Text{String: tracker.openerID, Valid: tracker.openerID != ""},
		OpenedAt:         openedAt,
		ClosedAt:         closedAt,
		ClosedBy:         closedBy,
		StorageKey:       key,
		TotalMessages:    pgtype.Int4{Int32: int32(len(content)), Valid: true},
		TotalAttachments: pgtype.Int4{Int32: int32(totalAttachments), Valid: true},
		TotalEmbeds:      pgtype.Int4{Int32: int32(totalEmbeds), Valid: true},
		ServerConfigID:   tracker.guildID,
	}

	rowID := tracker.transcriptID
	if rowID == 0 {
		ref, err := queries.GetLatestTranscriptByTicket(ctx, tracker.guildID, tracker.channelID)
		if err == nil {
			rowID = ref.ID
		}
	}
	if rowID != 0 {
		params.ID = rowID
		if err := queries.UpdateTranscriptByID(ctx, params); err != nil {
			return 0, err
		}
	} else {
		row, err := queries.CreateTranscript(ctx, db.CreateTranscriptParams{
			ServerConfigID:   tracker.guildID,
			TicketID:         pgtype.Text{String: tracker.channelID, Valid: true},
			Username:         params.Username,
			UserID:           params.UserID,
			OpenedAt:         openedAt,
			ClosedAt:         closedAt,
			ClosedBy:         closedBy,
			StorageKey:       key,
			TotalMessages:    params.TotalMessages,
			TotalAttachments: params.TotalAttachments,
			TotalEmbeds:      params.TotalEmbeds,
		})
		if err != nil {
			return 0, err
		}
		rowID = row.ID
	}
	autoClose.Lock()
	tracker.transcriptID = rowID
	tracker.transcriptKey = key
	tracker.totalMessages = len(content)
	tracker.totalAttach = totalAttachments
	tracker.totalEmbeds = totalEmbeds
	autoClose.Unlock()
	return rowID, nil
}

func trackerWeSession() *discordgo.Session {
	autoClose.Lock()
	defer autoClose.Unlock()
	return autoClose.session
}

func sendTranscriptLog(tracker *ticketTracker, closedBy string, rowID int32) {
	session := trackerWeSession()
	if session == nil || queries == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverConfig, err := queries.GetServerConfig(ctx, tracker.guildID)
	if err != nil || !serverConfig.TicketTranscriptCid.Valid {
		return
	}
	SendTranscriptLog(session, serverConfig.TicketTranscriptCid.String, tracker.guildID, rowID, tracker.channelID, tracker.openerID, closedBy, tracker.totalMessages, tracker.totalAttach)
}

func isUnknownChannel(err error) bool {
	var restErr *discordgo.RESTError
	return errors.As(err, &restErr) && restErr.Message != nil && restErr.Message.Code == discordgo.ErrCodeUnknownChannel
}

func recoverCleanupMarkers(s *discordgo.Session) {
	markers, err := listMarkers()
	if err != nil {
		return
	}
	for _, marker := range markers {
		switch {
		case marker.State == "confirmed" || marker.Deleted:
			tracker := trackerForMarker(marker)
			autoClose.Lock()
			if !tracker.closing {
				tracker.cleanup = &marker
			}
			autoClose.Unlock()
		case marker.Mode == "auto":
			tracker := getTracker(marker.Channel)
			if tracker == nil {
				deleteMarker(marker.Server, marker.Channel)
				continue
			}
			autoClose.Lock()
			cfg, ok := autoClose.configs[marker.Server]
			eligible := ok && autoClose.ready && autoClose.hydrated && !tracker.closing && dueReason(tracker, cfg, nowFn()) != ""
			autoClose.Unlock()
			if !eligible {
				deleteMarker(marker.Server, marker.Channel)
				continue
			}
			autoClose.Lock()
			if tracker.closing || tracker.cleanup != nil {
				autoClose.Unlock()
				continue
			}
			tracker.closing = true
			autoClose.Unlock()
			if deleted, err := closeTracker(tracker, "auto", marker.ClosedBy); !deleted || err != nil {
				autoClose.Lock()
				tracker.closing = false
				autoClose.Unlock()
			}
		default:
			// Manual and external pending markers are never executed on recovery.
			deleteMarker(marker.Server, marker.Channel)
		}
	}
	retryDeletedTickets()
}

func trackerForMarker(marker cleanupMarker) *ticketTracker {
	autoClose.Lock()
	defer autoClose.Unlock()
	if tracker, ok := autoClose.trackers[marker.Channel]; ok {
		return tracker
	}
	tracker := &ticketTracker{
		guildID:       marker.Server,
		openerID:      marker.Opener,
		channelID:     marker.Channel,
		createdAt:     time.Unix(marker.OpenedAt, 0),
		capture:       true,
		observedSince: nowFn(),
		uncertain:     true,
	}
	autoClose.trackers[marker.Channel] = tracker
	return tracker
}
