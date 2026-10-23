package serverconfig

import "github.com/Sush1sui/FNS_BOT/internal/db"

type StaffMember struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"globalName"`
	AvatarURL  string `json:"avatarUrl"`
}

type StaffRole struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color int    `json:"color"`
}

type StaffResponse struct {
	Members             []StaffMember `json:"members"`
	Roles               []StaffRole   `json:"roles"`
	AuthorizedMemberIds []string      `json:"authorizedMemberIds"`
	AuthorizedRoleIds   []string      `json:"authorizedRoleIds"`
}

type StaffUpdatePayload struct {
	AuthorizedMemberIds []string `json:"authorizedMemberIds"`
	AuthorizedRoleIds   []string `json:"authorizedRoleIds"`
}

// Handler handles HTTP endpoints for server configuration
type Handler struct {
	DB *db.Queries
}
type TicketPermissions []string

type FormData struct {
	TicketNameStyle                string `json:"TicketNameStyle"`
	TicketTranscripts              string `json:"TicketTranscripts"`
	MaxTicketsPerUser              int    `json:"MaxTicketsPerUser"`
	TicketPermissionsAttachFiles   bool   `json:"TicketPermissionsAttachFiles"`
	TicketPermissionsEmbedLinks    bool   `json:"TicketPermissionsEmbedLinks"`
	TicketPermissionsAddReactions  bool   `json:"TicketPermissionsAddReactions"`
	AutoClose                      bool   `json:"AutoClose"`
	AutoCloseOnUserLeave           bool   `json:"AutoCloseOnUserLeave"`
	AutoCloseNoResponseDays        *int   `json:"AutoCloseNoResponseDays"`
	AutoCloseNoResponseHours       *int   `json:"AutoCloseNoResponseHours"`
	AutoCloseNoResponseMins        *int   `json:"AutoCloseNoResponseMins"`
	AutoCloseSinceLastMessageDays  *int   `json:"AutoCloseSinceLastMessageDays"`
	AutoCloseSinceLastMessageHours *int   `json:"AutoCloseSinceLastMessageHours"`
	AutoCloseSinceLastMessageMins  *int   `json:"AutoCloseSinceLastMessageMins"`
}

type ServerConfigResponse struct {
	db.ServerConfig
	AutoClose                      bool `json:"AutoClose"`
	AutoCloseOnUserLeave           bool `json:"AutoCloseOnUserLeave"`
	AutoCloseNoResponseDays        *int `json:"AutoCloseNoResponseDays"`
	AutoCloseNoResponseHours       *int `json:"AutoCloseNoResponseHours"`
	AutoCloseNoResponseMins        *int `json:"AutoCloseNoResponseMins"`
	AutoCloseSinceLastMessageDays  *int `json:"AutoCloseSinceLastMessageDays"`
	AutoCloseSinceLastMessageHours *int `json:"AutoCloseSinceLastMessageHours"`
	AutoCloseSinceLastMessageMins  *int `json:"AutoCloseSinceLastMessageMins"`
}
