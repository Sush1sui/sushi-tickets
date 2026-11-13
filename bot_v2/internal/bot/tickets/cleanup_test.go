package tickets

import (
	"context"
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

func TestGuildReconciliationUsesCacheOnly(t *testing.T) {
	d := &cleanupDB{surgeDB: surgeDB{activeChannels: []string{"live", "deleted", "forbidden", "unavailable", "network", "other404"}}}
	installCleanupDB(t, d)
	s, _ := discordgo.New("Bot test")
	s.State.GuildAdd(&discordgo.Guild{ID: "42"})
	s.State.ChannelAdd(&discordgo.Channel{ID: "live", GuildID: "42"})
	var requests int
	s.Client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		return response(200, `{}`), nil
	})}
	HandleGuildCreate(s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "42", Unavailable: true}})
	if d.selected != nil {
		t.Fatal("queried an unavailable guild")
	}
	HandleGuildCreate(s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "42"}})
	if requests != 0 {
		t.Fatalf("cache-only reconciliation issued %d Discord requests", requests)
	}
	if !reflect.DeepEqual(d.selected, []any{int64(42)}) {
		t.Fatalf("incorrect reconciliation scope: args=%v", d.selected)
	}
	if len(d.deleted) != 0 {
		t.Fatalf("deleted rows on cache miss without proof: %v", d.deleted)
	}
}
