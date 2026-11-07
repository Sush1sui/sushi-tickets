package tickets

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestCanCloseTicketPermissions(t *testing.T) {
	old := queries
	queries = nil
	t.Cleanup(func() { queries = old })

	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	s.State.GuildAdd(&discordgo.Guild{ID: "42"})
	s.State.ChannelAdd(&discordgo.Channel{ID: "ticket", GuildID: "42", Topic: "ticket_opener:123"})

	for _, tc := range []struct {
		name   string
		userID string
		perms  int64
		allow  bool
	}{
		{name: "ticket opener without staff or server permissions", userID: "123", allow: true},
		{name: "other member without staff or server permissions"},
		{name: "manage channels alone", perms: discordgo.PermissionManageChannels},
		{name: "administrator", perms: discordgo.PermissionAdministrator, allow: true},
		{name: "manage server", perms: discordgo.PermissionManageServer, allow: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			userID := tc.userID
			if userID == "" {
				userID = "456"
			}
			i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
				GuildID: "42", ChannelID: "ticket",
				Member: &discordgo.Member{User: &discordgo.User{ID: userID}, Permissions: tc.perms},
			}}
			if got := canCloseTicket(s, i); got != tc.allow {
				t.Fatalf("canCloseTicket = %t, want %t", got, tc.allow)
			}
		})
	}
}
