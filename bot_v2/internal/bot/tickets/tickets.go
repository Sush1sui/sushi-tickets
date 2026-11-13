package tickets

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i == nil || i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	switch {
	case strings.HasPrefix(data.CustomID, panelButtonPrefix):
		panelID, ok := parsePanelID(data.CustomID, panelButtonPrefix)
		if !ok {
			return
		}
		handlePanelOpen(s, i, panelID)
	case data.CustomID == panelSelectID:
		if len(data.Values) == 0 {
			return
		}
		panelID, err := strconv.Atoi(data.Values[0])
		if err != nil {
			return
		}
		handlePanelOpen(s, i, int32(panelID))
	case data.CustomID == closeTicketID:
		handleCloseTicket(s, i)
	}
}

func HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i == nil || i.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := i.ModalSubmitData()
	panelCustomID, _, _ := strings.Cut(data.CustomID, ":")
	panelID, ok := parsePanelID(panelCustomID, panelModalPrefix)
	if !ok {
		return
	}

	if err := respondDeferred(s, i); err != nil {
		return
	}
	if !panelPolicyMu.TryRLock() {
		editEphemeral(s, i, "Panel settings are being updated. Please try again.")
		return
	}
	defer panelPolicyMu.RUnlock()
	submitOpening(s, i, panelID, &data)
}

func submitOpening(s *discordgo.Session, i *discordgo.InteractionCreate, panelID int32, data *discordgo.ModalSubmitInteractionData) {
	if i.Member == nil || i.Member.User == nil {
		return
	}
	key := i.GuildID + ":" + i.Member.User.ID
	if openingQueue.hasPending(key) {
		openingQueue.noteDuplicate()
		editEphemeral(s, i, ticketFailureMessage(errPending))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		editEphemeral(s, i, ticketFailureMessage(err))
		return
	}
	panel, err := loadAdmission(ctx, serverID, panelID)
	if err != nil {
		editEphemeral(s, i, ticketFailureMessage(err))
		return
	}
	var qna []QnA
	if data != nil {
		_, version, _ := strings.Cut(data.CustomID, ":")
		if version != panel.version {
			editEphemeral(s, i, "Panel questions changed. Please open the form again.")
			return
		}
		answersMap := extractModalAnswersMap(*data)
		for idx, q := range panel.questions {
			qna = append(qna, QnA{Question: q.Label, Answer: answersMap[idx]})
		}
		// Fallback if DB questions list is empty but modal has answers
		if len(qna) == 0 && len(answersMap) > 0 {
			for idx, ans := range answersMap {
				qna = append(qna, QnA{Question: fmt.Sprintf("Question %d", idx+1), Answer: ans})
			}
		}
	}
	j := &ticketJob{session: s, interaction: i, panelID: panelID, qna: qna, window: panel.window, queuedAt: time.Now(), key: key}
	if err := openingQueue.enqueue(j); err != nil {
		editEphemeral(s, i, ticketFailureMessage(err))
	}
}

func windowClosedMessage(nextStartTime int64) string {
	if nextStartTime > 0 {
		return fmt.Sprintf("Tickets are currently closed for this panel. Reopens <t:%d:R>.", nextStartTime)
	}
	return "Tickets are currently closed for this panel."
}

func handlePanelOpen(s *discordgo.Session, i *discordgo.InteractionCreate, panelID int32) {
	responseCtx, responseCancel := context.WithTimeout(context.Background(), 2750*time.Millisecond)
	defer responseCancel()
	if i.Member == nil || i.Member.User == nil {
		return
	}
	if !panelPolicyMu.TryRLock() {
		respondEphemeral(s, i, "Panel settings are being updated. Please try again.", discordgo.WithContext(responseCtx))
		return
	}
	defer panelPolicyMu.RUnlock()
	if openingQueue.hasPending(i.GuildID + ":" + i.Member.User.ID) {
		respondEphemeral(s, i, ticketFailureMessage(errPending), discordgo.WithContext(responseCtx))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		return
	}
	panel, err := loadAdmission(ctx, serverID, panelID)
	if err != nil {
		respondEphemeral(s, i, ticketFailureMessage(err), discordgo.WithContext(responseCtx))
		return
	}
	if len(panel.questions) > 0 {
		showQuestionsModal(s, i, panelID, panel.title, panel.questions, panel.version, discordgo.WithContext(responseCtx))
		return
	}
	if err := respondDeferred(s, i, discordgo.WithContext(responseCtx)); err != nil {
		return
	}
	j := &ticketJob{session: s, interaction: i, panelID: panelID, window: panel.window, queuedAt: time.Now(), key: i.GuildID + ":" + i.Member.User.ID}
	if err := openingQueue.enqueue(j); err != nil {
		editEphemeral(s, i, ticketFailureMessage(err))
	}
}

func openTicket(ctx context.Context, j *ticketJob) error {
	s, i, panelID, qna := j.session, j.interaction, j.panelID, j.qna
	if queries == nil {
		return fmt.Errorf("queries not set")
	}
	if i.GuildID == "" || i.Member == nil || i.Member.User == nil {
		return fmt.Errorf("missing guild/member")
	}

	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		return err
	}

	panel, err := queries.GetPanelConfigByID(ctx, serverID, panelID)
	if err != nil {
		return err
	}

	welcomeMsg, hasWelcome, err := queries.GetPanelWelcome(ctx, panelID)
	if err != nil {
		return err
	}

	serverConfig, err := queries.GetServerConfig(ctx, serverID)
	if err != nil {
		return err
	}
	if serverConfig.MaxTicketPerUser > 0 {
		count, err := queries.CountActiveTicketsByUser(ctx, serverID, i.Member.User.ID)
		if err != nil {
			return err
		}
		if count >= int64(serverConfig.MaxTicketPerUser) {
			channels, err := queries.GetActiveTicketChannelsByUser(ctx, serverID, i.Member.User.ID)
			if err != nil {
				return err
			}
			for _, channelID := range channels {
				// Cache misses are unknown, not proof of deletion. Never fall back
				// to a Discord REST read; stale rows are cleared by gateway events.
				if s.State != nil {
					if _, err := s.State.Channel(channelID); err == nil {
						continue
					}
				}
			}
			count, err = queries.CountActiveTicketsByUser(ctx, serverID, i.Member.User.ID)
			if err != nil {
				return err
			}

			if count >= int64(serverConfig.MaxTicketPerUser) {
				return errMaxTickets
			}
		}
	}
	allowedPerms := parseTicketPermissions(serverConfig.TicketPermissions)
	userPerms := baseUserPerms(allowedPerms)
	staffPerms := staffPerms()

	parentID := ""
	if panel.CategoryID.Valid {
		parentID = panel.CategoryID.String
	}

	channelName := buildTicketChannelName(serverConfig.TicketNameStyle, i.Member.User.Username, i.ID)
	channelTopic := fmt.Sprintf("ticket_opener:%s panel:%d", i.Member.User.ID, panelID)

	overwrites, err := buildPermissionOverwrites(ctx, s, i, serverID, panel.MentionRolesOnOpen, userPerms, staffPerms)
	if err != nil {
		return err
	}

	// Recheck live policy after preparation; never hold a DB connection while waiting on Discord.
	panelPolicyMu.RLock()
	err = validateJobWindow(ctx, j)
	panelPolicyMu.RUnlock()
	if err != nil {
		return err
	}
	requestCtx, requestCancel := context.WithCancel(ctx)
	defer requestCancel()
	var startErr error
	beforeSend := func(cfg *discordgo.RequestConfig) {
		panelPolicyMu.RLock()
		defer panelPolicyMu.RUnlock()
		if ctx.Err() != nil {
			startErr = ctx.Err()
			requestCancel()
			return
		}
		startErr = j.queue.markStarted(j)
		if startErr != nil {
			requestCancel()
		}
	}
	channel, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             parentID,
		Topic:                channelTopic,
		PermissionOverwrites: overwrites,
	}, discordgo.WithContext(requestCtx), discordgo.WithRestRetries(0), discordgo.WithRetryOnRatelimit(false), beforeSend)
	if startErr != nil {
		return startErr
	}
	if err != nil {
		return err
	}

	if err := queries.CreateActiveTicket(ctx, serverID, i.Member.User.ID, channel.ID, time.Now().Unix()); err != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, cleanupErr := s.ChannelDelete(channel.ID, discordgo.WithContext(cleanupCtx), discordgo.WithRestRetries(0), discordgo.WithRetryOnRatelimit(false))
		log.Printf("ticket_queue persistence_failed channel=%s user=%s cleanup_failed=%t", channel.ID, i.Member.User.ID, cleanupErr != nil)
		return err
	}
	TrackNewTicket(serverID, i.Member.User.ID, channel.ID, time.Now(), serverConfig.TicketTranscriptCid.Valid && serverConfig.TicketTranscriptCid.String != "")

	SendWelcomeMessage(ctx, s, channel.ID, i.Member.User, panel.MentionRolesOnOpen, welcomeMsg, hasWelcome, qna)
	editEphemeral(s, i, fmt.Sprintf("Ticket created! Please check <#%s>", channel.ID))
	return nil
}

func buildPermissionOverwrites(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, serverID int64, mentionRoles []string, userPerms, staffPerms int64) ([]*discordgo.PermissionOverwrite, error) {
	overwrites := make([]*discordgo.PermissionOverwrite, 0, 6)
	added := map[string]struct{}{}

	overwrites = append(overwrites, &discordgo.PermissionOverwrite{
		ID:   i.GuildID,
		Type: discordgo.PermissionOverwriteTypeRole,
		Deny: discordgo.PermissionViewChannel,
	})
	added[i.GuildID] = struct{}{}

	userID := i.Member.User.ID
	overwrites = append(overwrites, &discordgo.PermissionOverwrite{
		ID:    userID,
		Type:  discordgo.PermissionOverwriteTypeMember,
		Allow: userPerms,
	})
	added[userID] = struct{}{}

	if s != nil && s.State != nil && s.State.User != nil {
		botID := s.State.User.ID
		if _, ok := added[botID]; !ok {
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{
				ID:    botID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageChannels,
			})
			added[botID] = struct{}{}
		}
	}

	if queries != nil {
		members, err := queries.GetAuthorizedMembers(ctx, serverID)
		if err != nil {
			return nil, err
		}
		for _, mid := range members {
			if mid == "" {
				continue
			}
			if _, ok := added[mid]; ok {
				continue
			}
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{
				ID:    mid,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: staffPerms,
			})
			added[mid] = struct{}{}
		}

		roles, err := queries.GetAuthorizedRoles(ctx, serverID)
		if err != nil {
			return nil, err
		}
		for _, rid := range roles {
			if rid == "" {
				continue
			}
			if _, ok := added[rid]; ok {
				continue
			}
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{
				ID:    rid,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: staffPerms,
			})
			added[rid] = struct{}{}
		}
	}

	for _, roleID := range mentionRoles {
		if roleID == "" {
			continue
		}
		if _, ok := added[roleID]; ok {
			continue
		}
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID:    roleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: staffPerms,
		})
		added[roleID] = struct{}{}
	}

	return overwrites, nil
}

func handleCloseTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondDeferred(s, i)

	if i.ChannelID == "" {
		editEphemeral(s, i, "Channel not found.")
		return
	}
	if !canCloseTicket(s, i) {
		editEphemeral(s, i, "Not allowed to close this ticket.")
		return
	}
	if message := CloseTicketManually(s, i); message != "" {
		editEphemeral(s, i, message)
	}
}

// CloseTicketManually keeps the existing authorization boundary and persists
// the configured transcript before deleting the channel. Capture stays active
// until the channel is gone.
func CloseTicketManually(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		return "Failed to close ticket."
	}
	channelID := i.ChannelID
	tracker := getTracker(channelID)
	if tracker == nil {
		createdAt := nowFn()
		openerID := ""
		if queries != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			info, infoErr := queries.GetActiveTicketByChannel(ctx, serverID, channelID)
			cancel()
			if infoErr == nil {
				createdAt = time.Unix(info.CreatedAt, 0)
				openerID = info.UserID
			}
		}
		if openerID == "" && i.Member != nil && i.Member.User != nil {
			openerID = i.Member.User.ID
		}
		tracker = &ticketTracker{
			guildID: serverID, openerID: openerID, channelID: channelID,
			createdAt: createdAt, observedSince: nowFn(), uncertain: true,
			capture: loadCapture(serverID),
		}
		autoClose.Lock()
		autoClose.trackers[channelID] = tracker
		autoClose.Unlock()
	}
	if tracker.capture && !spoolExists(serverID, channelID) {
		if err := spoolStart(serverID, channelID, tracker.openerID, tracker.createdAt.Unix(), nowFn().Unix(), !tracker.createdHere, "capture_started_at_close"); err != nil {
			return "Failed to start transcript capture. Ticket left open."
		}
	}

	autoClose.Lock()
	if tracker.closing {
		autoClose.Unlock()
		return "Ticket is already closing."
	}
	tracker.closing = true
	autoClose.Unlock()

	closedBy := ""
	if i.Member != nil && i.Member.User != nil {
		closedBy = i.Member.User.ID
	}
	deleted, closeErr := closeTracker(tracker, "manual", closedBy)
	if !deleted || closeErr != nil {
		autoClose.Lock()
		tracker.closing = false
		autoClose.Unlock()
	}
	if closeErr != nil {
		if deleted {
			return "Ticket closed, but transcript finalization will retry."
		}
		return "Failed to save transcript. Ticket left open; try again."
	}
	return ""
}
