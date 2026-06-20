package tickets

import (
	"fmt"
	"strings"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/utils"
	"github.com/bwmarrin/discordgo"
)

func SendWelcomeMessage(
	s *discordgo.Session,
	channelID string,
	user *discordgo.User,
	mentionRoles []string,
	welcome db.WelcomeMsgConfig,
	hasWelcome bool,
	qna []QnA,
) {
	if s == nil || user == nil {
		return
	}

	embed := buildWelcomeEmbed(user, welcome, hasWelcome)

	mentions := make([]string, 0, 1+len(mentionRoles))
	mentions = append(mentions, user.Mention())
	for _, roleID := range mentionRoles {
		if roleID == "" {
			continue
		}
		mentions = append(mentions, fmt.Sprintf("<@&%s>", roleID))
	}

	closeButton := discordgo.Button{
		Label:    "Close Ticket",
		Style:    discordgo.DangerButton,
		CustomID: closeTicketID,
		Emoji: &discordgo.ComponentEmoji{
			Name: "🔒",
		},
	}

	message := &discordgo.MessageSend{
		Content: strings.Join(mentions, " "),
		Embeds:  []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{closeButton}},
		},
	}

	_, _ = s.ChannelMessageSendComplex(channelID, message)
	sendQnAMessage(s, channelID, user, qna, embed.Color)
}

func buildWelcomeEmbed(user *discordgo.User, welcome db.WelcomeMsgConfig, hasWelcome bool) *discordgo.MessageEmbed {
	if hasWelcome {
		embed := &discordgo.MessageEmbed{
			Title:       welcome.Title,
			Description: welcome.Description,
			Color:       int(welcome.EmbedColor),
		}

		if url := utils.TextOrEmpty(welcome.TitleUrl); url != "" {
			embed.URL = url
		}
		if url := utils.TextOrEmpty(welcome.LargeImgUrl); url != "" {
			embed.Image = &discordgo.MessageEmbedImage{URL: url}
		}
		if url := utils.TextOrEmpty(welcome.SmallImgUrl); url != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: url}
		}
		if footerText := utils.TextOrEmpty(welcome.Footer); footerText != "" {
			embed.Footer = &discordgo.MessageEmbedFooter{Text: footerText}
			if iconURL := utils.TextOrEmpty(welcome.FooterIconUrl); iconURL != "" {
				embed.Footer.IconURL = iconURL
			}
		}

		if embed.Description == "" {
			embed.Description = fmt.Sprintf("Welcome %s! A staff member will assist you shortly.", user.Mention())
		}
		if embed.Title == "" {
			embed.Title = "Sushi Tickets"
		}

		return embed
	}

	return &discordgo.MessageEmbed{
		Title:       "Sushi Tickets",
		Description: fmt.Sprintf("Hello %s! Thank you for creating a ticket. A staff member will assist you shortly.", user.Mention()),
		Color:       0xFF5A36,
	}
}

func sendQnAMessage(s *discordgo.Session, channelID string, user *discordgo.User, qna []QnA, color int) {
	if len(qna) == 0 {
		return
	}

	var sb strings.Builder
	for i, qa := range qna {
		ans := strings.TrimSpace(qa.Answer)
		if ans == "" {
			ans = "N/A"
		}
		const maxRunes = 500
		if len([]rune(ans)) > maxRunes {
			r := []rune(ans)
			ans = string(r[:maxRunes-1]) + "…"
		}
		sb.WriteString(fmt.Sprintf("**%s**: %s", qa.Question, ans))
		if i < len(qna)-1 {
			sb.WriteString("\n")
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Ticket Responses",
		Color:       color,
		Description: sb.String(),
	}

	message := &discordgo.MessageSend{
		Content: fmt.Sprintf("> Ticket responses from %s", user.Mention()),
		Embeds:  []*discordgo.MessageEmbed{embed},
	}

	_, _ = s.ChannelMessageSendComplex(channelID, message)
}
