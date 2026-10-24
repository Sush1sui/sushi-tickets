package bot

import (
	"fmt"
	"log"
	"sync"

	"github.com/Sush1sui/FNS_BOT/internal/bot/deploy"
	"github.com/Sush1sui/FNS_BOT/internal/bot/tickets"
	"github.com/Sush1sui/FNS_BOT/internal/config"
	"github.com/bwmarrin/discordgo"
)

var Session *discordgo.Session

func StartBot() {
	s, e := discordgo.New("Bot " + config.Load().BotToken)
	if e != nil {
		log.Fatalf("Error creating Discord session: %v", e)
	}

	s.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildPresences | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages

	deployed := map[string]struct{}{}
	var deployedMu sync.Mutex

	deployEvents := func(sess *discordgo.Session) {
		// Safe to call multiple times; DeployEvents is idempotent.
		if err := deploy.DeployEvents(sess, ""); err != nil {
			log.Printf("Event deploy failed: %v", err)
		}
	}

	deployForGuild := func(sess *discordgo.Session, guildID string) {
		if guildID == "" {
			return
		}

		deployedMu.Lock()
		if _, ok := deployed[guildID]; ok {
			deployedMu.Unlock()
			return
		}
		deployed[guildID] = struct{}{}
		deployedMu.Unlock()

		// Bulk overwrite per guild to keep commands in sync.
		if err := deploy.DeployCommands(sess, guildID); err != nil {
			log.Printf("Command deploy failed for guild %s: %v", guildID, err)
		}
	}

	s.AddHandler(func(sess *discordgo.Session, r *discordgo.Ready) {
		// Initial deploy pass on startup.
		sess.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: "idle",
			Activities: []*discordgo.Activity{
				{
					Name: "Sushi Tickets!",
				},
			},
		})

		deployEvents(sess)

		for _, g := range r.Guilds {
			gid := g.ID
			go deployForGuild(sess, gid)
		}
	})

	s.AddHandler(tickets.HandleChannelDelete)
	s.AddHandler(tickets.HandleGuildCreate)
	s.AddHandler(func(sess *discordgo.Session, g *discordgo.GuildCreate) {
		if g == nil {
			return
		}
		// Deploy for newly joined guild.
		deployEvents(sess)
		go deployForGuild(sess, g.ID)
	})

	s.AddHandler(func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		if i == nil {
			return
		}

		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			name := i.ApplicationCommandData().Name
			// Fast map lookup for handlers.
			if handler, ok := deploy.CommandHandlers[name]; ok {
				handler(sess, i)
			}
		case discordgo.InteractionMessageComponent:
			tickets.HandleComponentInteraction(sess, i)
		case discordgo.InteractionModalSubmit:
			tickets.HandleModalSubmit(sess, i)
		}
	})

	e = s.Open()
	if e != nil {
		log.Fatalf("Error opening Discord session: %v", e)
	}

	Session = s
	fmt.Println("Bot is up!")
}

func StopBot() {
	if Session == nil {
		return
	}

	_ = Session.Close()
	Session = nil
}
