package tickets

import (
	"context"
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
	if tracker := getTracker(event.ID); tracker != nil {
		autoClose.Lock()
		if tracker.closing {
			// The close path owns this channel and will persist, finalize, and clean up.
			autoClose.Unlock()
			return
		}
		tracker.closing = true
		autoClose.Unlock()

		// Externally deleted channels cannot guarantee transcript-before-delete.
		// Write a best-effort transcript from the local spool, then clean up.
		marker := cleanupMarker{
			Server: serverID, Channel: event.ID, Mode: "external", State: "pending",
			Deleted: true, Opener: tracker.openerID, OpenedAt: tracker.createdAt.Unix(),
		}
		autoClose.Lock()
		tracker.cleanup = &marker
		autoClose.Unlock()
		err := writeMarker(marker)
		if err == nil {
			err = finalizeDeletedTicket(tracker, marker, ^uint64(0))
		}
		if err != nil {
			log.Printf("external_delete cleanup channel=%s failed: %v", event.ID, err)
			autoClose.Lock()
			tracker.closing = false
			autoClose.Unlock()
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := queries.DeleteActiveTicketByChannel(ctx, serverID, event.ID); err != nil {
		log.Printf("active ticket cleanup guild=%s channel=%s: %v", event.GuildID, event.ID, err)
	}
}

// Reconcile on guild availability using cache only. A cache miss is unknown,
// not proof of deletion, and must never trigger a Discord REST read.
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
		if s == nil || s.State == nil {
			return
		}
		if _, err := s.State.Channel(channelID); err == nil {
			continue
		}
		log.Printf("active ticket channel missing from cache; retaining row guild=%s channel=%s", event.ID, channelID)
	}
}
