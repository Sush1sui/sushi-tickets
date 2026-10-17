package utils

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseServerID(r *http.Request) (int64, error) {
	serverIDStr := r.PathValue("server_id")
	return strconv.ParseInt(serverIDStr, 10, 64)
}

func ParseServerPanelIDs(r *http.Request) (int64, int32, error) {
	serverIDStr := r.PathValue("server_id")
	panelIDStr := r.PathValue("panel_id")
	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	panelID64, err := strconv.ParseInt(panelIDStr, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return serverID, int32(panelID64), nil
}

func ParseServerMultiPanelIDs(r *http.Request) (int64, int32, error) {
	serverIDStr := r.PathValue("server_id")
	panelIDStr := r.PathValue("multi_panel_id")
	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	panelID64, err := strconv.ParseInt(panelIDStr, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return serverID, int32(panelID64), nil
}

func ParseServerPanelWindowIDs(r *http.Request) (int64, int32, int64, error) {
	serverIDStr := r.PathValue("server_id")
	panelIDStr := r.PathValue("panel_id")
	windowIDStr := r.PathValue("window_id")
	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	panelID64, err := strconv.ParseInt(panelIDStr, 10, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	windowID, err := strconv.ParseInt(windowIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return serverID, int32(panelID64), windowID, nil
}

func ToText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func IsBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func MissingPanelFields(title, content, channelID, btnColor, btnTxt string) []string {
	missing := make([]string, 0, 5)
	if IsBlank(title) {
		missing = append(missing, "title")
	}
	if IsBlank(content) {
		missing = append(missing, "content")
	}
	if IsBlank(channelID) {
		missing = append(missing, "channelId")
	}
	if IsBlank(btnColor) {
		missing = append(missing, "btnColor")
	}
	if IsBlank(btnTxt) {
		missing = append(missing, "btnTxt")
	}
	return missing
}

func MissingMultiPanelFields(title, content, channelID string) []string {
	missing := make([]string, 0, 3)
	if IsBlank(title) {
		missing = append(missing, "title")
	}
	if IsBlank(content) {
		missing = append(missing, "content")
	}
	if IsBlank(channelID) {
		missing = append(missing, "channelId")
	}
	return missing
}

func HasWelcomeMessage(title, description, titleURL, largeImgURL, smallImgURL, footerText, footerIconURL string) bool {
	return title != "" || description != "" || titleURL != "" ||
		largeImgURL != "" || smallImgURL != "" || footerText != "" ||
		footerIconURL != ""
}

func ToWelcomeParams(embedColor int32, title, description, titleURL, largeImgURL, smallImgURL, footerText, footerIconURL string) db.WelcomeMessageParams {
	return db.WelcomeMessageParams{
		EmbedColor:    embedColor,
		Title:         title,
		Description:   description,
		TitleURL:      ToText(titleURL),
		LargeImgUrl:   ToText(largeImgURL),
		SmallImgUrl:   ToText(smallImgURL),
		Footer:        ToText(footerText),
		FooterIconUrl: ToText(footerIconURL),
	}
}
