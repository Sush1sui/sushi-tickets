package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/config"
	"github.com/bwmarrin/discordgo"
)

func usernameOrID(s *discordgo.Session, guildID, userID string) string {
	if userID == "" {
		return ""
	}
	if s != nil && s.State != nil {
		if m, err := s.State.Member(guildID, userID); err == nil && m.User != nil {
			return m.User.Username
		}
	}
	return userID
}

func buildTranscriptContent(messages []*discordgo.Message) ([]transcriptMessage, int, int, []transcriptParticipant) {
	items := make([]transcriptMessage, 0, len(messages))
	participants := make(map[string]*transcriptParticipant)
	attachmentsTotal := 0
	embedsTotal := 0

	for _, m := range messages {
		if m == nil || m.Author == nil {
			continue
		}

		attachmentsTotal += len(m.Attachments)
		embedsTotal += len(m.Embeds)

		msgType := "message"
		content := m.Content

		timestamp := m.Timestamp.Format(time.RFC3339)
		item := transcriptMessage{
			ID:   m.ID,
			Type: msgType,
			Author: transcriptAuthor{
				ID:            m.Author.ID,
				Username:      m.Author.Username,
				Discriminator: m.Author.Discriminator,
				Avatar:        strPtr(m.Author.Avatar),
				Bot:           m.Author.Bot,
			},
			Content:   &content,
			Timestamp: timestamp,
			Edited:    m.EditedTimestamp != nil,
		}
		if m.EditedTimestamp != nil {
			edited := m.EditedTimestamp.Format(time.RFC3339)
			item.EditedTimestamp = &edited
		}

		if len(m.Embeds) > 0 {
			item.Embeds = make([]transcriptEmbed, 0, len(m.Embeds))
			for _, e := range m.Embeds {
				embed := transcriptEmbed{
					Title:       strPtr(e.Title),
					Description: strPtr(e.Description),
					URL:         strPtr(e.URL),
					Color:       intPtr(e.Color),
					Fields:      make([]transcriptEmbedField, 0, len(e.Fields)),
				}
				for _, f := range e.Fields {
					embed.Fields = append(embed.Fields, transcriptEmbedField{Name: f.Name, Value: f.Value, Inline: f.Inline})
				}
				if e.Image != nil {
					embed.Image = &transcriptImage{URL: e.Image.URL}
				}
				if e.Thumbnail != nil {
					embed.Thumbnail = &transcriptImage{URL: e.Thumbnail.URL}
				}
				if e.Footer != nil {
					embed.Footer = &transcriptFooter{Text: e.Footer.Text, IconURL: strPtr(e.Footer.IconURL)}
				}
				if e.Author != nil {
					embed.Author = &transcriptEmbedAuthor{Name: e.Author.Name, URL: strPtr(e.Author.URL), IconURL: strPtr(e.Author.IconURL)}
				}
				item.Embeds = append(item.Embeds, embed)
			}
		}

		if len(m.Attachments) > 0 {
			item.Attachments = make([]transcriptAttachment, 0, len(m.Attachments))
			for _, a := range m.Attachments {
				item.Attachments = append(item.Attachments, transcriptAttachment{
					ID:          a.ID,
					Filename:    a.Filename,
					URL:         a.URL,
					ProxyURL:    a.ProxyURL,
					Size:        a.Size,
					ContentType: strPtr(a.ContentType),
					Width:       intPtr(a.Width),
					Height:      intPtr(a.Height),
				})
			}
		}

		if len(m.Reactions) > 0 {
			item.Reactions = make([]transcriptReaction, 0, len(m.Reactions))
			for _, r := range m.Reactions {
				emoji := r.Emoji.Name
				item.Reactions = append(item.Reactions, transcriptReaction{Emoji: emoji, Count: r.Count})
			}
		}

		items = append(items, item)
		if p, ok := participants[m.Author.ID]; ok {
			p.MessageCount++
		} else {
			participants[m.Author.ID] = &transcriptParticipant{ID: m.Author.ID, Username: m.Author.Username, MessageCount: 1}
		}
	}

	list := make([]transcriptParticipant, 0, len(participants))
	for _, p := range participants {
		list = append(list, *p)
	}

	return items, attachmentsTotal, embedsTotal, list
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func intPtr(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func fetchAllMessages(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
	if s == nil {
		return nil, fmt.Errorf("missing session")
	}

	all := make([]*discordgo.Message, 0)
	before := ""
	for {
		msgs, err := s.ChannelMessages(channelID, 100, before, "", "")
		if err != nil {
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}
		all = append(all, msgs...)
		before = msgs[len(msgs)-1].ID
		if len(msgs) < 100 {
			break
		}
	}

	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	return all, nil
}

func canCloseTicket(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i == nil || i.Member == nil || i.Member.User == nil {
		return false
	}

	channel := getChannelFromInteraction(s, i)
	if channel != nil && strings.Contains(channel.Topic, "ticket_opener:") {
		parts := strings.Split(channel.Topic, "ticket_opener:")
		if len(parts) > 1 {
			opener := strings.Fields(parts[1])
			if len(opener) > 0 && opener[0] == i.Member.User.ID {
				return true
			}
		}
	}

	serverID, err := strconv.ParseInt(i.GuildID, 10, 64)
	if err != nil {
		return false
	}

	if queries != nil {
		members, _ := queries.GetAuthorizedMembers(context.Background(), serverID)
		for _, mid := range members {
			if mid == i.Member.User.ID {
				return true
			}
		}

		roles, _ := queries.GetAuthorizedRoles(context.Background(), serverID)
		roleSet := make(map[string]struct{}, len(i.Member.Roles))
		for _, rid := range i.Member.Roles {
			roleSet[rid] = struct{}{}
		}
		for _, rid := range roles {
			if _, ok := roleSet[rid]; ok {
				return true
			}
		}
	}

	return false
}

func showQuestionsModal(s *discordgo.Session, i *discordgo.InteractionCreate, panelID int32, questions []string, required []bool) {
	inputs := make([]discordgo.MessageComponent, 0, 5)
	limit := len(questions)
	if limit > 5 {
		limit = 5
	}

	for idx := 0; idx < limit; idx++ {
		q := questions[idx]
		req := false
		if idx < len(required) {
			req = required[idx]
		}
		inputs = append(inputs, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID: fmt.Sprintf("q_%d", idx),
					Label:    q,
					Style:    discordgo.TextInputParagraph,
					Required: req,
				},
			},
		})
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   fmt.Sprintf("%s%d", panelModalPrefix, panelID),
			Title:      "Ticket Questions",
			Components: inputs,
		},
	})
}

func loadPanelQuestions(panelID int32) ([]string, []bool, error) {
	if queries == nil {
		return nil, nil, fmt.Errorf("queries not set")
	}
	row, err := queries.GetPanelQuestions(context.Background(), panelID)
	if err != nil {
		return nil, nil, err
	}
	return row.Questions, row.RequiredFlags, nil
}

func parsePanelID(value, prefix string) (int32, bool) {
	idStr := strings.TrimPrefix(value, prefix)
	if idStr == "" {
		return 0, false
	}
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(id64), true
}

func parseTicketPermissions(raw []byte) map[string]struct{} {
	perms := make(map[string]struct{})
	if len(raw) == 0 {
		return perms
	}

	var list []string
	if err := json.Unmarshal(raw, &list); err != nil {
		return perms
	}

	for _, item := range list {
		perms[item] = struct{}{}
	}
	return perms
}

func baseUserPerms(perms map[string]struct{}) int64 {
	var allow int64 = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory
	if _, ok := perms["attachFiles"]; ok {
		allow |= discordgo.PermissionAttachFiles
	}
	if _, ok := perms["embedLinks"]; ok {
		allow |= discordgo.PermissionEmbedLinks
	}
	if _, ok := perms["addReactions"]; ok {
		allow |= discordgo.PermissionAddReactions
	}
	return allow
}

func getChannelFromInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.Channel {
	if s == nil || i == nil || i.ChannelID == "" {
		return nil
	}

	if s.State != nil {
		if ch, err := s.State.Channel(i.ChannelID); err == nil {
			return ch
		}
	}

	ch, err := s.Channel(i.ChannelID)
	if err != nil {
		return nil
	}
	return ch
}

func staffPerms() int64 {
	return discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionAddReactions | discordgo.PermissionAttachFiles | discordgo.PermissionEmbedLinks
}

func buildTicketChannelName(style, username, interactionID string) string {
	name := "ticket"
	if style == "name" && username != "" {
		name = "ticket-" + username
	} else {
		suffix := interactionID
		if len(suffix) > 4 {
			suffix = suffix[len(suffix)-4:]
		}
		name = fmt.Sprintf("ticket-%s", suffix)
	}

	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "--", "-")
	return name
}

func respondDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func editEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &message})
}

func extractModalAnswersMap(data discordgo.ModalSubmitInteractionData) map[int]string {
	answers := make(map[int]string)

	for _, row := range data.Components {
		var rowComponents []discordgo.MessageComponent
		switch r := row.(type) {
		case discordgo.ActionsRow:
			rowComponents = r.Components
		case *discordgo.ActionsRow:
			rowComponents = r.Components
		default:
			continue
		}
		for _, comp := range rowComponents {
			var input discordgo.TextInput
			switch c := comp.(type) {
			case discordgo.TextInput:
				input = c
			case *discordgo.TextInput:
				input = *c
			default:
				continue
			}
			var idx int
			if _, err := fmt.Sscanf(input.CustomID, "q_%d", &idx); err == nil {
				answers[idx] = input.Value
			}
		}
	}

	return answers
}

func SendTranscriptLog(
	s *discordgo.Session,
	logChannelID string,
	serverID int64,
	rowID int32,
	ticketChannelName string,
	openerID string,
	closedByID string,
	totalMessages int,
	totalAttachments int,
) {
	if s == nil || logChannelID == "" {
		return
	}

	dashboardURL := fmt.Sprintf("%s/servers/%d/transcripts/%d", config.Load().ClientOrigin, serverID, rowID)

	embed := &discordgo.MessageEmbed{
		Title:       "📄 Ticket Transcript Log",
		Description: fmt.Sprintf("A ticket transcript has been successfully saved for ticket **#%s**.", ticketChannelName),
		Color:       0xffffff, // White
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Ticket Info",
				Value:  fmt.Sprintf("**ID:** `%d`\n**Channel:** `#%s`", rowID, ticketChannelName),
				Inline: true,
			},
			{
				Name:   "Opener",
				Value:  fmt.Sprintf("<@%s>\n`%s`", openerID, openerID),
				Inline: true,
			},
			{
				Name:   "Closed By",
				Value:  fmt.Sprintf("<@%s>\n`%s`", closedByID, closedByID),
				Inline: true,
			},
			{
				Name:   "Stats",
				Value:  fmt.Sprintf("💬 **Messages:** `%d`\n📎 **Attachments:** `%d`", totalMessages, totalAttachments),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Sushi Tickets • Transcript System",
		},
	}

	button := discordgo.Button{
		Label: "View Transcript Dashboard",
		Style: discordgo.LinkButton,
		URL:   dashboardURL,
		Emoji: &discordgo.ComponentEmoji{Name: "🌐"},
	}

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{button}},
		},
	}

	_, _ = s.ChannelMessageSendComplex(logChannelID, msg)
}
