package tickets

import (
	"fmt"
)

type QnA struct {
	Question string
	Answer   string
}

type QuestionData struct {
	Label       string
	IsRequired  bool
	Style       string
	Placeholder string
}

const (
	panelButtonPrefix = "open_ticket_"
	panelSelectID     = "select_panel"
	panelModalPrefix  = "ticket_modal_"
	closeTicketID     = "close_ticket"
)

var errMaxTickets = fmt.Errorf("max tickets reached")

type transcriptPayload struct {
	TicketID string              `json:"ticketId"`
	Username string              `json:"username"`
	UserID   string              `json:"userId"`
	Messages []transcriptMessage `json:"messages"`
	Metadata transcriptMetadata  `json:"metadata"`
}

type transcriptAuthor struct {
	ID            string  `json:"id"`
	Username      string  `json:"username"`
	Discriminator string  `json:"discriminator"`
	Avatar        *string `json:"avatar"`
	Bot           bool    `json:"bot"`
}

type transcriptEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type transcriptImage struct {
	URL string `json:"url"`
}

type transcriptFooter struct {
	Text    string  `json:"text"`
	IconURL *string `json:"iconUrl"`
}

type transcriptEmbedAuthor struct {
	Name    string  `json:"name"`
	URL     *string `json:"url"`
	IconURL *string `json:"iconUrl"`
}

type transcriptEmbed struct {
	Title       *string                `json:"title"`
	Description *string                `json:"description"`
	URL         *string                `json:"url"`
	Color       *int                   `json:"color"`
	Fields      []transcriptEmbedField `json:"fields"`
	Image       *transcriptImage       `json:"image"`
	Thumbnail   *transcriptImage       `json:"thumbnail"`
	Footer      *transcriptFooter      `json:"footer"`
	Author      *transcriptEmbedAuthor `json:"author"`
}

type transcriptAttachment struct {
	ID          string  `json:"id"`
	Filename    string  `json:"filename"`
	URL         string  `json:"url"`
	ProxyURL    string  `json:"proxyUrl"`
	Size        int     `json:"size"`
	ContentType *string `json:"contentType"`
	Width       *int    `json:"width"`
	Height      *int    `json:"height"`
}

type transcriptReaction struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

type transcriptMessage struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Author          transcriptAuthor       `json:"author"`
	Content         *string                `json:"content"`
	Timestamp       string                 `json:"timestamp"`
	Embeds          []transcriptEmbed      `json:"embeds,omitempty"`
	Attachments     []transcriptAttachment `json:"attachments,omitempty"`
	Edited          bool                   `json:"edited"`
	EditedTimestamp *string                `json:"editedTimestamp"`
	Reactions       []transcriptReaction   `json:"reactions,omitempty"`
}

type transcriptClosedBy struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type transcriptParticipant struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	MessageCount int    `json:"messageCount"`
}

type transcriptMetadata struct {
	TicketOpenedAt   string                  `json:"ticketOpenedAt"`
	TicketClosedAt   string                  `json:"ticketClosedAt"`
	ClosedBy         transcriptClosedBy      `json:"closedBy"`
	TotalMessages    int                     `json:"totalMessages"`
	TotalAttachments int                     `json:"totalAttachments"`
	TotalEmbeds      int                     `json:"totalEmbeds"`
	Participants     []transcriptParticipant `json:"participants"`
}
