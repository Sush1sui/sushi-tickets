package tickets

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type cleanupDB struct {
	surgeDB
	selected []any
	deleted  [][]any
}

func (d *cleanupDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	d.selected = args
	return d.surgeDB.Query(ctx, sql, args...)
}

func (d *cleanupDB) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	d.deleted = append(d.deleted, args)
	return pgconn.CommandTag{}, nil
}

func installCleanupDB(t *testing.T, d *cleanupDB) {
	t.Helper()
	old := queries
	queries = db.New(d)
	t.Cleanup(func() { queries = old })
}

func TestChannelDeleteRemovesOnlyMatchingGuildChannel(t *testing.T) {
	d := &cleanupDB{}
	installCleanupDB(t, d)
	HandleChannelDelete(nil, nil)
	HandleChannelDelete(nil, &discordgo.ChannelDelete{})
	HandleChannelDelete(nil, &discordgo.ChannelDelete{Channel: &discordgo.Channel{ID: "dm"}})
	if len(d.deleted) != 0 {
		t.Fatal("invalid or DM event deleted a ticket")
	}
	HandleChannelDelete(nil, &discordgo.ChannelDelete{Channel: &discordgo.Channel{GuildID: "42", ID: "deleted"}})
	if !reflect.DeepEqual(d.deleted, [][]any{{int64(42), "deleted"}}) {
		t.Fatalf("unexpected delete scope: %v", d.deleted)
	}
}

func TestGuildReconciliationOnlyPurgesConfirmedUnknownChannels(t *testing.T) {
	d := &cleanupDB{surgeDB: surgeDB{activeChannels: []string{"live", "deleted", "forbidden", "unavailable", "network", "other404"}}}
	installCleanupDB(t, d)
	s, _ := discordgo.New("Bot test")
	// A stale cached channel must not hide an offline deletion.
	s.State.GuildAdd(&discordgo.Guild{ID: "42"})
	s.State.ChannelAdd(&discordgo.Channel{ID: "deleted", GuildID: "42"})
	var requests int
	s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.Method != http.MethodGet {
			t.Fatalf("cleanup tried to mutate Discord: %s", r.Method)
		}
		switch r.URL.Path {
		case "/api/v9/channels/deleted":
			return response(404, `{"code":10003,"message":"Unknown Channel"}`), nil
		case "/api/v9/channels/forbidden":
			return response(403, `{"code":50001,"message":"Missing Access"}`), nil
		case "/api/v9/channels/unavailable":
			return response(503, `{}`), nil
		case "/api/v9/channels/network":
			return nil, errors.New("network down")
		case "/api/v9/channels/other404":
			return response(404, `{"code":0}`), nil
		default:
			return response(200, `{"id":"live","guild_id":"42"}`), nil
		}
	})}
	HandleGuildCreate(s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "42", Unavailable: true}})
	if d.selected != nil {
		t.Fatal("queried an unavailable guild")
	}
	HandleGuildCreate(s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "42"}})
	if requests != len(d.activeChannels) || !reflect.DeepEqual(d.selected, []any{int64(42)}) {
		t.Fatalf("incomplete or incorrectly scoped reconciliation: requests=%d args=%v", requests, d.selected)
	}
	if !reflect.DeepEqual(d.deleted, [][]any{{int64(42), "deleted"}}) {
		t.Fatalf("deleted unconfirmed channel rows: %v", d.deleted)
	}
}
