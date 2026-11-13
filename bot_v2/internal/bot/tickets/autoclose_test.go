package tickets

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/storage"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type closeDB struct {
	transcriptCID string
	updates       atomic.Int32
	deletes       atomic.Int32
}

func (d *closeDB) Exec(_ context.Context, sql string, _ ...interface{}) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "DELETE FROM active_ticket") {
		d.deletes.Add(1)
		return pgconn.CommandTag{}, nil
	}
	if strings.Contains(sql, "UPDATE transcript") {
		d.updates.Add(1)
	}
	return pgconn.CommandTag{}, nil
}

func (d *closeDB) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return &stubRows{}, nil
}

func (d *closeDB) QueryRow(_ context.Context, sql string, _ ...interface{}) pgx.Row {
	return stubRow(func(dest ...any) error {
		switch {
		case strings.Contains(sql, "FROM server_config"):
			*dest[0].(*int64) = 42
			*dest[1].(*string) = "number"
			*dest[2].(*pgtype.Text) = pgtype.Text{String: d.transcriptCID, Valid: d.transcriptCID != ""}
			*dest[3].(*int32) = 2
			*dest[4].(*[]byte) = []byte("[]")
			*dest[5].(*pgtype.Int4) = pgtype.Int4{}
			*dest[6].(*pgtype.Int4) = pgtype.Int4{}
			return nil
		case strings.Contains(sql, "FROM transcript"):
			*dest[0].(*int32) = 7
			*dest[1].(*string) = "transcripts/42/tickets/ticket.json"
			return nil
		}
		return pgx.ErrNoRows
	})
}

func resetAutoCloseForTest(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("BOT_TOKEN", "test")
	t.Setenv("DISCORD_CLIENT_ID", "test")
	t.Setenv("DISCORD_CLIENT_SECRET", "test")
	t.Setenv("DISCORD_REDIRECT_URL", "http://localhost")
	t.Setenv("CLIENT_ORIGIN", "http://localhost")
	t.Setenv("ACCESS_TOKEN_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	autoClose.Lock()
	autoClose.trackers = make(map[string]*ticketTracker)
	autoClose.configs = make(map[int64]autoCloseSettings)
	autoClose.captures = make(map[int64]bool)
	autoClose.ready = false
	autoClose.hydrated = false
	autoClose.session = nil
	autoClose.Unlock()
	spools.Lock()
	spools.m = make(map[string]*spoolState)
	spools.Unlock()
	oldQueries := queries
	oldStorage := storageClient
	t.Cleanup(func() {
		queries = oldQueries
		storageClient = oldStorage
		spools.Lock()
		spools.m = make(map[string]*spoolState)
		spools.Unlock()
		autoClose.Lock()
		autoClose.trackers = make(map[string]*ticketTracker)
		autoClose.configs = make(map[int64]autoCloseSettings)
		autoClose.captures = make(map[int64]bool)
		autoClose.ready = false
		autoClose.hydrated = false
		autoClose.session = nil
		autoClose.Unlock()
	})
}

func TestSpoolReplayRebuildsTranscriptState(t *testing.T) {
	resetAutoCloseForTest(t)
	storageClient = storage.NewClientAt(t.TempDir())
	now := time.Unix(1_700_000_000, 0)
	if err := spoolStart(42, "ticket", "opener", now.Unix(), now.Unix(), false, ""); err != nil {
		t.Fatal(err)
	}
	human := &discordgo.Message{ID: "100", Content: "hello", Timestamp: now, Author: &discordgo.User{ID: "opener", Username: "user"}}
	bot := &discordgo.Message{ID: "101", Content: "welcome", Timestamp: now.Add(time.Second), Author: &discordgo.User{ID: "bot", Bot: true}}
	_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(human, "create")})
	_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(bot, "create")})
	_ = spoolAppend(42, "ticket", spoolEvent{T: "ra", MID: "100", Emoji: "👍", User: "u1"})
	_ = spoolAppend(42, "ticket", spoolEvent{T: "ra", MID: "100", Emoji: "👍", User: "u2"})
	_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "update", M: &spoolMessage{ID: "101", Content: "edited", Timestamp: now.Add(time.Second).UnixMilli()}})
	_ = spoolAppend(42, "ticket", spoolEvent{T: "d", ID: "101"})
	if err := spoolFlush("ticket", true); err != nil {
		t.Fatal(err)
	}
	replay, err := spoolReplayTicket(42, "ticket")
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.messages) != 1 || replay.messages[0].ID != "100" {
		t.Fatalf("unexpected messages: %d", len(replay.messages))
	}
	if !replay.responded {
		t.Fatal("human message not treated as response")
	}
	if len(replay.messages[0].Reactions) != 1 || replay.messages[0].Reactions[0].Count != 2 {
		t.Fatalf("reaction counts not rebuilt: %+v", replay.messages[0].Reactions)
	}
	if replay.lastActivity.Before(now) {
		t.Fatalf("last activity = %s", replay.lastActivity)
	}
}

func TestHumanMessageCountsAsResponseBotDoesNot(t *testing.T) {
	resetAutoCloseForTest(t)
	now := time.Unix(1_700_000_000, 0)
	autoClose.Lock()
	autoClose.trackers["ticket"] = &ticketTracker{
		guildID: 42, openerID: "opener", channelID: "ticket",
		createdAt: now, lastActivity: now, responseKnown: true, observedSince: now,
	}
	autoClose.Unlock()

	HandleAutoCloseMessage(nil, &discordgo.MessageCreate{Message: &discordgo.Message{
		ID: "1", ChannelID: "ticket", Timestamp: now.Add(time.Minute),
		Author: &discordgo.User{ID: "staff"},
	}})
	HandleAutoCloseMessage(nil, &discordgo.MessageCreate{Message: &discordgo.Message{
		ID: "2", ChannelID: "ticket", Timestamp: now.Add(2 * time.Minute),
		Author: &discordgo.User{ID: "bot", Bot: true},
	}})
	tracker := getTracker("ticket")
	if !tracker.responded || !tracker.responseKnown {
		t.Fatal("human message did not mark response")
	}
	if !tracker.lastActivity.Equal(now.Add(time.Minute)) {
		t.Fatalf("bot message changed activity: %s", tracker.lastActivity)
	}
}

func TestDueReasonUsesObservedWindowWhenUncertain(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cfg := autoCloseSettings{active: true, inactivity: time.Hour, noResponse: time.Hour}
	tracker := &ticketTracker{createdAt: now.Add(-5 * time.Hour), observedSince: now, uncertain: true, responseKnown: false}
	if reason := dueReason(tracker, cfg, now.Add(59*time.Minute)); reason != "" {
		t.Fatalf("closed before a full observed interval: %s", reason)
	}
	if reason := dueReason(tracker, cfg, now.Add(time.Hour)); reason != "inactivity" {
		t.Fatalf("expected inactivity, got %q", reason)
	}
	tracker.responded = true
	tracker.responseKnown = true
	if reason := dueReason(tracker, cfg, now.Add(time.Hour)); reason != "inactivity" {
		t.Fatalf("responded ticket should still use inactivity, got %q", reason)
	}
	cfg.noResponse = 0
	cfg.inactivity = 0
	if reason := dueReason(tracker, cfg, now.Add(24*time.Hour)); reason != "" {
		t.Fatal("disabled rules closed")
	}
}

func TestNoResponseRequiresKnownHistory(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cfg := autoCloseSettings{active: true, noResponse: time.Hour}
	unknown := &ticketTracker{createdAt: now.Add(-2 * time.Hour), observedSince: now.Add(-2 * time.Hour), responseKnown: false}
	if reason := dueReason(unknown, cfg, now); reason != "" {
		t.Fatalf("unknown history triggered no-response: %q", reason)
	}
	known := &ticketTracker{createdAt: now.Add(-2 * time.Hour), observedSince: now.Add(-2 * time.Hour), responseKnown: true}
	if reason := dueReason(known, cfg, now); reason != "no_response" {
		t.Fatalf("known unanswered ticket did not qualify: %q", reason)
	}
	known.responded = true
	if reason := dueReason(known, cfg, now); reason != "" {
		t.Fatal("responded ticket triggered no-response")
	}
}

func TestTranscriptFailureBlocksChannelDelete(t *testing.T) {
	resetAutoCloseForTest(t)
	database := &closeDB{transcriptCID: "logs"}
	queries = db.New(database)
	storageClient = nil
	session, _ := discordgo.New("Bot test")
	var requests atomic.Int32
	session.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		if r.Method == http.MethodGet {
			t.Fatal("unexpected Discord read")
		}
		return response(200, `{}`), nil
	})}
	autoClose.Lock()
	autoClose.session = session
	autoClose.Unlock()
	tracker := &ticketTracker{
		guildID: 42, openerID: "opener", channelID: "ticket",
		createdAt: nowFn().Add(-time.Hour), capture: true,
	}
	deleted, err := closeTracker(tracker, "manual", "staff")
	if err == nil {
		t.Fatal("expected transcript persistence failure")
	}
	if deleted {
		t.Fatal("channel reported deleted despite persistence failure")
	}
	if requests.Load() != 0 {
		t.Fatalf("Discord was contacted %d times", requests.Load())
	}
}

func TestClosePersistsTranscriptBeforeDeleteAndCleansUp(t *testing.T) {
	resetAutoCloseForTest(t)
	database := &closeDB{transcriptCID: "logs"}
	queries = db.New(database)
	storageClient = storage.NewClientAt(t.TempDir())
	session, _ := discordgo.New("Bot test")
	var deletes, getRequests atomic.Int32
	session.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet {
			getRequests.Add(1)
			t.Fatal("unexpected Discord read")
		}
		if r.Method == http.MethodDelete {
			deletes.Add(1)
		}
		return response(200, `{}`), nil
	})}
	autoClose.Lock()
	autoClose.session = session
	autoClose.ready = true
	autoClose.hydrated = true
	autoClose.configs[42] = autoCloseSettings{active: true, inactivity: time.Minute}
	autoClose.Unlock()

	now := nowFn()
	tracker := &ticketTracker{
		guildID: 42, openerID: "opener", channelID: "ticket",
		createdAt: now.Add(-time.Hour), lastActivity: now.Add(-time.Hour),
		observedSince: now.Add(-time.Hour), responseKnown: true, capture: true,
	}
	autoClose.Lock()
	autoClose.trackers["ticket"] = tracker
	autoClose.Unlock()
	if err := spoolStart(42, "ticket", "opener", tracker.createdAt.Unix(), tracker.createdAt.Unix(), false, ""); err != nil {
		t.Fatal(err)
	}
	_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(&discordgo.Message{
		ID: "100", Content: "hello", Timestamp: now,
		Author: &discordgo.User{ID: "opener", Username: "user"},
	}, "create")})

	deleted, err := closeTracker(tracker, "manual", "staff")
	if err != nil || !deleted {
		t.Fatalf("close failed: deleted=%t err=%v", deleted, err)
	}
	if deletes.Load() != 1 || getRequests.Load() != 0 {
		t.Fatalf("unexpected Discord calls deletes=%d gets=%d", deletes.Load(), getRequests.Load())
	}
	if database.updates.Load() == 0 || database.deletes.Load() == 0 {
		t.Fatal("transcript row was not upserted or active row not deleted")
	}
	if getTracker("ticket") != nil {
		t.Fatal("tracker retained after confirmed close")
	}
	if spoolExists(42, "ticket") {
		t.Fatal("spool retained after confirmed close")
	}
	if markers, _ := listMarkers(); len(markers) != 0 {
		t.Fatalf("marker retained: %+v", markers)
	}
}

func TestRecoveryCancelsPendingManualDeletion(t *testing.T) {
	resetAutoCloseForTest(t)
	database := &closeDB{transcriptCID: "logs"}
	queries = db.New(database)
	storageClient = storage.NewClientAt(t.TempDir())
	session, _ := discordgo.New("Bot test")
	var deletes atomic.Int32
	session.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodDelete {
			deletes.Add(1)
		}
		if r.Method == http.MethodGet {
			t.Fatal("unexpected Discord read")
		}
		return response(200, `{}`), nil
	})}
	if err := writeMarker(cleanupMarker{Server: 42, Channel: "ticket", Mode: "manual", State: "pending", ClosedBy: "staff"}); err != nil {
		t.Fatal(err)
	}
	markers, _ := listMarkers()
	if len(markers) != 1 {
		t.Fatal("marker missing")
	}
	recoverCleanupMarkers(session)
	if deletes.Load() != 0 {
		t.Fatal("pending manual marker deleted a channel on recovery")
	}
	if markers, _ := listMarkers(); len(markers) != 0 {
		t.Fatal("cancelled marker retained")
	}
}

func TestSpoolWriteFailureBlocksClose(t *testing.T) {
	resetAutoCloseForTest(t)
	storageClient = storage.NewClientAt(t.TempDir())
	if err := spoolStart(42, "ticket", "opener", 1, 1, false, ""); err != nil {
		t.Fatal(err)
	}
	spools.Lock()
	spools.m["ticket"].failed = io.ErrClosedPipe
	spools.Unlock()
	tracker := &ticketTracker{guildID: 42, openerID: "opener", channelID: "ticket", capture: true}
	if _, err := closeTracker(tracker, "auto", "auto"); err == nil {
		t.Fatal("close continued with a failed spool")
	}
}

func TestLateEventsDuringDeletionAreFinalized(t *testing.T) {
	resetAutoCloseForTest(t)
	database := &closeDB{transcriptCID: "logs"}
	queries = db.New(database)
	client := storage.NewClientAt(t.TempDir())
	storageClient = client
	session, _ := discordgo.New("Bot test")
	now := nowFn()
	session.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet {
			t.Fatal("unexpected Discord read")
		}
		if r.Method == http.MethodDelete {
			_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(&discordgo.Message{
				ID: "200", Content: "late message", Timestamp: now,
				Author: &discordgo.User{ID: "staff", Username: "staff"},
			}, "create")})
		}
		return response(200, `{}`), nil
	})}
	autoClose.Lock()
	autoClose.session = session
	autoClose.ready = true
	autoClose.hydrated = true
	autoClose.configs[42] = autoCloseSettings{active: true, inactivity: time.Minute}
	autoClose.Unlock()

	tracker := &ticketTracker{
		guildID: 42, openerID: "opener", channelID: "ticket",
		createdAt: now.Add(-time.Hour), lastActivity: now.Add(-time.Hour),
		observedSince: now.Add(-time.Hour), responseKnown: true, capture: true,
	}
	autoClose.Lock()
	autoClose.trackers["ticket"] = tracker
	autoClose.Unlock()
	if err := spoolStart(42, "ticket", "opener", tracker.createdAt.Unix(), tracker.createdAt.Unix(), false, ""); err != nil {
		t.Fatal(err)
	}
	_ = spoolAppend(42, "ticket", spoolEvent{T: "m", Kind: "create", M: spoolMessageFromDiscord(&discordgo.Message{
		ID: "100", Content: "first", Timestamp: now,
		Author: &discordgo.User{ID: "opener", Username: "user"},
	}, "create")})

	if _, err := closeTracker(tracker, "manual", "staff"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(client.BaseDir(), "transcripts", "42", "tickets", "ticket.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "late message") {
		t.Fatal("late event was not finalized into the transcript")
	}
	if strings.Contains(string(data), `"historyComplete":true`) {
		t.Fatal("uncertain capture labeled complete")
	}
}

func TestPendingAutoMarkerRechecksEligibility(t *testing.T) {
	resetAutoCloseForTest(t)
	database := &closeDB{transcriptCID: "logs"}
	queries = db.New(database)
	storageClient = storage.NewClientAt(t.TempDir())
	session, _ := discordgo.New("Bot test")
	var deletes atomic.Int32
	session.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet {
			t.Fatal("unexpected Discord read")
		}
		if r.Method == http.MethodDelete {
			deletes.Add(1)
		}
		return response(200, `{}`), nil
	})}

	now := nowFn()
	recent := &ticketTracker{
		guildID: 42, openerID: "opener", channelID: "recent",
		createdAt: now.Add(-time.Hour), lastActivity: now, observedSince: now.Add(-time.Hour),
		capture: true,
	}
	autoClose.Lock()
	autoClose.session = session
	autoClose.ready = true
	autoClose.hydrated = true
	autoClose.configs[42] = autoCloseSettings{active: true, inactivity: time.Minute}
	autoClose.trackers["recent"] = recent
	autoClose.Unlock()
	if err := writeMarker(cleanupMarker{Server: 42, Channel: "recent", Mode: "auto", State: "pending", ClosedBy: "auto"}); err != nil {
		t.Fatal(err)
	}
	recoverCleanupMarkers(session)
	if deletes.Load() != 0 {
		t.Fatal("ineligible pending auto marker deleted a channel")
	}
	if markers, _ := listMarkers(); len(markers) != 0 {
		t.Fatal("ineligible marker not cancelled")
	}
}
