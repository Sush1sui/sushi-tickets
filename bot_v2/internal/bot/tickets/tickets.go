package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
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
	panelID, ok := parsePanelID(data.CustomID, panelModalPrefix)
	if !ok {
		return
	}

	_, dbQuestions, err := loadPanelQuestions(int32(panelID))
	if err != nil {
		log.Printf("failed to load questions for modal submit: %v", err)
	}

	answersMap := extractModalAnswersMap(data)
	qna := make([]QnA, 0, len(dbQuestions))
	for idx, q := range dbQuestions {
		ans := answersMap[idx]
		qna = append(qna, QnA{Question: q.Label, Answer: ans})
	}

	// Fallback if DB questions list is empty but modal has answers
	if len(qna) == 0 && len(answersMap) > 0 {
		for idx, ans := range answersMap {
			qna = append(qna, QnA{Question: fmt.Sprintf("Question %d", idx+1), Answer: ans})
		}
	}

	respondDeferred(s, i)
	if err := openTicket(s, i, panelID, qna); err != nil {
		log.Printf("open ticket failed: %v", err)
		if err == errMaxTickets {
			editEphemeral(s, i, "You reached max open tickets. Please close existing tickets.")
			return
		}
		editEphemeral(s, i, "Failed to create ticket. Please try again later.")
		return
	}
}

func handlePanelOpen(s *discordgo.Session, i *discordgo.InteractionCreate, panelID int32) {
	modalTitle, questions, err := loadPanelQuestions(panelID)
	if err != nil {
		log.Printf("load questions failed: %v", err)
		respondEphemeral(s, i, "Failed to load panel. Please try again later.")
		return
	}

	if len(questions) > 0 {
		showQuestionsModal(s, i, panelID, modalTitle, questions)
		return
	}

	respondDeferred(s, i)
	if err := openTicket(s, i, panelID, nil); err != nil {
		log.Printf("open ticket failed: %v", err)
		if err == errMaxTickets {
			editEphemeral(s, i, "You reached max open tickets. Please close existing tickets.")
			return
		}
		editEphemeral(s, i, "Failed to create ticket. Please try again later.")
		return
	}
}

func openTicket(s *discordgo.Session, i *discordgo.InteractionCreate, panelID int32, qna []QnA) error {
	if queries == nil {
		return fmt.Errorf("queries not set")
	}
	if i.GuildID == "" || i.Member == nil || i.Member.User == nil {
		return fmt.Errorf("missing guild/member")
	}

	ctx := context.Background()
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

	serverConfig, _ := queries.GetServerConfig(ctx, serverID)
	if serverConfig.MaxTicketPerUser > 0 {
		count, err := queries.CountActiveTicketsByUser(ctx, serverID, i.Member.User.ID)
		if err != nil {
			return err
		}
		if count >= int64(serverConfig.MaxTicketPerUser) {
			if s != nil && s.State != nil {
				channels, err := queries.GetActiveTicketChannelsByUser(ctx, serverID, i.Member.User.ID)
				if err == nil {
					for _, channelID := range channels {
						if _, err := s.State.Channel(channelID); err == nil {
							continue
						}
						if err := queries.DeleteActiveTicketByChannel(ctx, serverID, channelID); err != nil {
							log.Printf("delete stale active ticket failed: %v", err)
						}
					}
					count, err = queries.CountActiveTicketsByUser(ctx, serverID, i.Member.User.ID)
					if err != nil {
						return err
					}
				}
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

	channel, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             parentID,
		Topic:                channelTopic,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		return err
	}

	if err := queries.CreateActiveTicket(ctx, serverID, i.Member.User.ID, channel.ID, time.Now().Unix()); err != nil {
		log.Printf("create active ticket failed: %v", err)
	}

	SendWelcomeMessage(s, channel.ID, i.Member.User, panel.MentionRolesOnOpen, welcomeMsg, hasWelcome, qna)
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
		members, _ := queries.GetAuthorizedMembers(ctx, serverID)
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

		roles, _ := queries.GetAuthorizedRoles(ctx, serverID)
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

	channelID := i.ChannelID
	if channelID == "" {
		editEphemeral(s, i, "Channel not found.")
		return
	}

	if !canCloseTicket(s, i) {
		editEphemeral(s, i, "Not allowed to close this ticket.")
		return
	}

	saveTranscriptOnClose(s, i)

	if queries != nil {
		serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
		if err == nil {
			if err := queries.DeleteActiveTicketByChannel(context.Background(), serverID, channelID); err != nil {
				log.Printf("delete active ticket failed: %v", err)
			}
		}
	}

	if _, err := s.ChannelDelete(channelID); err != nil {
		editEphemeral(s, i, "Failed to close ticket.")
		return
	}
}

func saveTranscriptOnClose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if queries == nil || storageClient == nil {
		return
	}
	if i == nil || i.GuildID == "" || i.ChannelID == "" {
		return
	}

	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		return
	}

	serverConfig, err := queries.GetServerConfig(context.Background(), serverID)
	if err != nil || !serverConfig.TicketTranscriptCid.Valid || serverConfig.TicketTranscriptCid.String == "" {
		return
	}

	messages, err := fetchAllMessages(s, i.ChannelID)
	if err != nil {
		log.Printf("fetch transcript messages failed: %v", err)
		return
	}

	openedAt := int64(0)
	userID := ""
	username := ""
	if info, err := queries.GetActiveTicketByChannel(context.Background(), serverID, i.ChannelID); err == nil {
		openedAt = info.CreatedAt
		userID = info.UserID
	}
	if openedAt == 0 {
		openedAt = time.Now().Unix()
	}
	if userID == "" && i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
		username = i.Member.User.Username
	}
	if username == "" && userID != "" && s != nil && s.State != nil {
		if m, err := s.State.Member(i.GuildID, userID); err == nil && m.User != nil {
			username = m.User.Username
		}
	}

	closedBy := ""
	if i.Member != nil && i.Member.User != nil {
		closedBy = i.Member.User.ID
	}
	closedAt := time.Now().Unix()

	content, totalAttachments, totalEmbeds, participants := buildTranscriptContent(messages)

	payload := transcriptPayload{
		TicketID: i.ChannelID,
		Username: username,
		UserID:   userID,
		Messages: content,
		Metadata: transcriptMetadata{
			TicketOpenedAt:   time.Unix(openedAt, 0).UTC().Format(time.RFC3339),
			TicketClosedAt:   time.Unix(closedAt, 0).UTC().Format(time.RFC3339),
			ClosedBy:         transcriptClosedBy{ID: closedBy, Username: usernameOrID(s, i.GuildID, closedBy)},
			TotalMessages:    len(content),
			TotalAttachments: totalAttachments,
			TotalEmbeds:      totalEmbeds,
			Participants:     participants,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal transcript failed: %v", err)
		return
	}

	storageKey := fmt.Sprintf("transcripts/%d/%s/%d.json", serverID, i.ChannelID, closedAt)
	if err := storageClient.UploadTranscript(context.Background(), storageKey, data); err != nil {
		log.Printf("upload transcript failed: %v", err)
		return
	}

	row, err := queries.CreateTranscript(context.Background(), db.CreateTranscriptParams{
		ServerConfigID:   serverID,
		TicketID:         pgtype.Text{String: i.ChannelID, Valid: true},
		Username:         pgtype.Text{String: username, Valid: username != ""},
		UserID:           pgtype.Text{String: userID, Valid: userID != ""},
		OpenedAt:         openedAt,
		ClosedAt:         closedAt,
		ClosedBy:         closedBy,
		StorageKey:       storageKey,
		TotalMessages:    pgtype.Int4{Int32: int32(len(content)), Valid: true},
		TotalAttachments: pgtype.Int4{Int32: int32(totalAttachments), Valid: true},
		TotalEmbeds:      pgtype.Int4{Int32: int32(totalEmbeds), Valid: true},
	})
	if err != nil {
		log.Printf("create transcript row failed: %v", err)
		return
	}

	channelName := ""
	if ch := getChannelFromInteraction(s, i); ch != nil {
		channelName = ch.Name
	} else {
		channelName = i.ChannelID
	}

	SendTranscriptLog(
		s,
		serverConfig.TicketTranscriptCid.String,
		serverID,
		row.ID,
		channelName,
		userID,
		closedBy,
		len(content),
		totalAttachments,
	)
}
