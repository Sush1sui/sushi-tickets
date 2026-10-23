package serverconfig

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Sush1sui/FNS_BOT/internal/bot"
	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func validateOptionalIntRange(value *int, field string, min, max int, errs utils.ValidationErrors) {
	if value == nil {
		return
	}
	utils.ValidateIntRange(*value, field, min, max, errs)
}

func validateServerConfigForm(f FormData) utils.ValidationErrors {
	errs := make(utils.ValidationErrors)
	utils.ValidateMaxLength(f.TicketNameStyle, "TicketNameStyle", 32, errs)
	utils.ValidateIntRange(f.MaxTicketsPerUser, "MaxTicketsPerUser", 1, 100, errs)
	validateOptionalIntRange(f.AutoCloseNoResponseDays, "AutoCloseNoResponseDays", 0, 365, errs)
	validateOptionalIntRange(f.AutoCloseNoResponseHours, "AutoCloseNoResponseHours", 0, 23, errs)
	validateOptionalIntRange(f.AutoCloseNoResponseMins, "AutoCloseNoResponseMins", 0, 59, errs)
	validateOptionalIntRange(f.AutoCloseSinceLastMessageDays, "AutoCloseSinceLastMessageDays", 0, 365, errs)
	validateOptionalIntRange(f.AutoCloseSinceLastMessageHours, "AutoCloseSinceLastMessageHours", 0, 23, errs)
	validateOptionalIntRange(f.AutoCloseSinceLastMessageMins, "AutoCloseSinceLastMessageMins", 0, 59, errs)
	return errs
}

func totalMinutes(days, hours, mins *int) int {
	total := 0
	if days != nil {
		total += *days * 24 * 60
	}
	if hours != nil {
		total += *hours * 60
	}
	if mins != nil {
		total += *mins
	}
	return total
}

func optionalMinutes(total int) pgtype.Int4 {
	if total <= 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(total), Valid: true}
}

func splitMinutes(total pgtype.Int4) (days, hours, mins *int) {
	if !total.Valid || total.Int32 <= 0 {
		return nil, nil, nil
	}

	remaining := int(total.Int32)
	dayValue := remaining / (24 * 60)
	remaining %= 24 * 60
	hourValue := remaining / 60
	minuteValue := remaining % 60

	return &dayValue, &hourValue, &minuteValue
}

func (h *Handler) HandleGetServerConfig(w http.ResponseWriter, r *http.Request) {
	serverIDStr := r.PathValue("server_id")
	showChannels := r.URL.Query().Get("show_channels") == "true"

	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid Server ID"})
		return
	}

	config, _ := h.DB.GetServerConfig(context.Background(), serverID)
	responseConfig := ServerConfigResponse{ServerConfig: config}

	if autoClose, err := h.DB.GetAutoCloseConfig(context.Background(), serverID); err == nil {
		responseConfig.AutoClose = autoClose.IsActive
		responseConfig.AutoCloseOnUserLeave = autoClose.CloseOnUserLeave
		responseConfig.AutoCloseNoResponseDays,
			responseConfig.AutoCloseNoResponseHours,
			responseConfig.AutoCloseNoResponseMins = splitMinutes(autoClose.CloseSinceOpenWithNoResponseMins)
		responseConfig.AutoCloseSinceLastMessageDays,
			responseConfig.AutoCloseSinceLastMessageHours,
			responseConfig.AutoCloseSinceLastMessageMins = splitMinutes(autoClose.CloseSinceLastMessageMins)
	}

	if showChannels {
		channels, _ := utils.GetGuildChannelsCache(bot.Session, serverIDStr)
		response := map[string]any{
			"config":   responseConfig,
			"channels": channels,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	writeJSON(w, http.StatusOK, responseConfig)
}

func (h *Handler) HandleUpdateServerConfig(w http.ResponseWriter, r *http.Request) {
	serverIDStr := r.PathValue("server_id")
	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid Server ID"})
		return
	}

	var form FormData
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	if errs := validateServerConfigForm(form); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": errs})
		return
	}

	// Encode ticket permissions as JSON array of names (only true ones)
	var perms TicketPermissions
	if form.TicketPermissionsAttachFiles {
		perms = append(perms, "attachFiles")
	}
	if form.TicketPermissionsEmbedLinks {
		perms = append(perms, "embedLinks")
	}
	if form.TicketPermissionsAddReactions {
		perms = append(perms, "addReactions")
	}
	permsJSON, _ := json.Marshal(perms)

	// Upsert server config
	cfg, err := h.DB.UpsertServerConfig(context.Background(), db.UpsertServerConfigParams{
		ID:                  serverID,
		TicketNameStyle:     form.TicketNameStyle,
		TicketTranscriptCid: pgtype.Text{String: form.TicketTranscripts, Valid: form.TicketTranscripts != ""},
		MaxTicketPerUser:    int32(form.MaxTicketsPerUser),
		TicketPermissions:   permsJSON,
	})
	if err != nil {
		log.Printf("failed to save config: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save config"})
		return
	}
	_ = cfg // use cfg to ensure it's read

	// Convert days/hours/mins to total minutes. Empty, zero, and negative
	// durations are intentionally stored as NULL.
	noResponseMins := optionalMinutes(totalMinutes(
		form.AutoCloseNoResponseDays,
		form.AutoCloseNoResponseHours,
		form.AutoCloseNoResponseMins,
	))
	sinceLastMessageMins := optionalMinutes(totalMinutes(
		form.AutoCloseSinceLastMessageDays,
		form.AutoCloseSinceLastMessageHours,
		form.AutoCloseSinceLastMessageMins,
	))

	// Upsert auto close config
	_, err = h.DB.UpsertAutoCloseConfig(context.Background(), db.UpsertAutoCloseConfigParams{
		ServerConfigID:                   serverID,
		IsActive:                         form.AutoClose,
		CloseOnUserLeave:                 form.AutoCloseOnUserLeave,
		CloseSinceOpenWithNoResponseMins: noResponseMins,
		CloseSinceLastMessageMins:        sinceLastMessageMins,
	})
	if err != nil {
		log.Printf("failed to save auto close config: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save auto close config"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
