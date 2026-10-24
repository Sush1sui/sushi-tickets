package tickets

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleChannelDelete(_ *discordgo.Session, event *discordgo.ChannelDelete) {
	if queries == nil || event == nil || event.Channel == nil || event.ID == "" {
		return
	}
	serverID, err := strconv.ParseInt(event.GuildID, 10, 64)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := queries.DeleteActiveTicketByChannel(ctx, serverID, event.ID); err != nil {
		log.Printf("active ticket cleanup guild=%s channel=%s: %v", event.GuildID, event.ID, err)
	}
}

// Reconcile on guild availability to catch channels deleted while the bot was offline.
func HandleGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if queries == nil || event == nil || event.Guild == nil || event.Unavailable {
		return
	}
	serverID, err := strconv.ParseInt(event.ID, 10, 64)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	channels, err := queries.GetActiveTicketChannelsByServer(ctx, serverID)
	if err != nil {
		log.Printf("active ticket reconciliation guild=%s: %v", event.ID, err)
		return
	}
	for _, channelID := range channels {
		if ctx.Err() != nil {
			return
		}
		// Cache misses and permission failures do not prove a channel was deleted.
		_, err := s.Channel(channelID, discordgo.WithContext(ctx), discordgo.WithRestRetries(0), discordgo.WithRetryOnRatelimit(false))
		var restErr *discordgo.RESTError
		if err == nil {
			continue
		}
		if !errors.As(err, &restErr) || restErr.Message == nil || restErr.Message.Code != discordgo.ErrCodeUnknownChannel {
			log.Printf("active ticket reconciliation guild=%s channel=%s: %v", event.ID, channelID, err)
			continue
		}
		if err := queries.DeleteActiveTicketByChannel(ctx, serverID, channelID); err != nil {
			log.Printf("active ticket cleanup guild=%s channel=%s: %v", event.ID, channelID, err)
		}
	}
}
