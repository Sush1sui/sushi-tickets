package transcripts

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)



func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) HandleListTranscripts(w http.ResponseWriter, r *http.Request) {
	serverID, err := parseServerID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid server id"})
		return
	}

	page, limit := parsePagination(r)

	items, err := h.DB.GetTranscriptsByServer(context.Background(), db.GetTranscriptsByServerParams{
		ServerConfigID: serverID,
		Limit:          int32(limit),
		Offset:         int32((page - 1) * limit),
	})
	if err != nil {
		log.Printf("load transcripts failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load transcripts"})
		return
	}

	total, err := h.DB.CountTranscriptsByServer(context.Background(), serverID)
	if err != nil {
		log.Printf("count transcripts failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to count transcripts"})
		return
	}

	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"transcripts": formatTranscriptList(items),
		"pagination": map[string]int{
			"page":  page,
			"limit": limit,
			"total": int(total),
			"pages": pages,
		},
	})
}

func (h *Handler) HandleGetTranscript(w http.ResponseWriter, r *http.Request) {
	serverID, transcriptID, err := parseIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	item, err := h.DB.GetTranscriptByID(context.Background(), db.GetTranscriptByIDParams{
		ID:             transcriptID,
		ServerConfigID: serverID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "transcript not found"})
			return
		}
		log.Printf("load transcript failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load transcript"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"transcript": formatTranscript(item),
	})
}

func (h *Handler) HandleGetTranscriptContent(w http.ResponseWriter, r *http.Request) {
	serverID, transcriptID, err := parseIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}

	item, err := h.DB.GetTranscriptByID(context.Background(), db.GetTranscriptByIDParams{
		ID:             transcriptID,
		ServerConfigID: serverID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "transcript not found"})
			return
		}
		log.Printf("load transcript failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load transcript"})
		return
	}

	data, err := h.Storage.DownloadTranscript(context.Background(), item.StorageKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "transcript file not found"})
			return
		}
		log.Printf("download transcript failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to download transcript"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

type TranscriptListItem struct {
	ID             int32  `json:"id"`
	TicketID       string `json:"ticketId"`
	Username       string `json:"username"`
	UserID         string `json:"userId"`
	OpenedAt       int64  `json:"openedAt"`
	ClosedAt       int64  `json:"closedAt"`
	ClosedBy       string `json:"closedBy"`
	TotalMessages  int    `json:"totalMessages"`
	TotalAttachments int  `json:"totalAttachments"`
	TotalEmbeds    int    `json:"totalEmbeds"`
}

type TranscriptDetail struct {
	ID             int32  `json:"id"`
	TicketID       string `json:"ticketId"`
	Username       string `json:"username"`
	UserID         string `json:"userId"`
	OpenedAt       int64  `json:"openedAt"`
	ClosedAt       int64  `json:"closedAt"`
	ClosedBy       string `json:"closedBy"`
	StorageKey     string `json:"storageKey"`
	TotalMessages  int    `json:"totalMessages"`
	TotalAttachments int  `json:"totalAttachments"`
	TotalEmbeds    int    `json:"totalEmbeds"`
}

func formatTranscriptList(items []db.Transcript) []TranscriptListItem {
	result := make([]TranscriptListItem, len(items))
	for i, t := range items {
		result[i] = TranscriptListItem{
			ID:             t.ID,
			TicketID:       pgTextOrEmpty(t.TicketID),
			Username:       pgTextOrEmpty(t.Username),
			UserID:         pgTextOrEmpty(t.UserID),
			OpenedAt:       t.OpenedAt,
			ClosedAt:       t.ClosedAt,
			ClosedBy:       t.ClosedBy,
			TotalMessages:  pgInt4OrZero(t.TotalMessages),
			TotalAttachments: pgInt4OrZero(t.TotalAttachments),
			TotalEmbeds:    pgInt4OrZero(t.TotalEmbeds),
		}
	}
	return result
}

func formatTranscript(t db.Transcript) TranscriptDetail {
	return TranscriptDetail{
		ID:             t.ID,
		TicketID:       pgTextOrEmpty(t.TicketID),
		Username:       pgTextOrEmpty(t.Username),
		UserID:         pgTextOrEmpty(t.UserID),
		OpenedAt:       t.OpenedAt,
		ClosedAt:       t.ClosedAt,
		ClosedBy:       t.ClosedBy,
		StorageKey:     t.StorageKey,
		TotalMessages:  pgInt4OrZero(t.TotalMessages),
		TotalAttachments: pgInt4OrZero(t.TotalAttachments),
		TotalEmbeds:    pgInt4OrZero(t.TotalEmbeds),
	}
}

func pgTextOrEmpty(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func pgInt4OrZero(t pgtype.Int4) int {
	if t.Valid {
		return int(t.Int32)
	}
	return 0
}

func parseServerID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("server_id"), 10, 64)
}

func parseIDs(r *http.Request) (int64, int32, error) {
	serverID, err := strconv.ParseInt(r.PathValue("server_id"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	transcriptID64, err := strconv.ParseInt(r.PathValue("transcript_id"), 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return serverID, int32(transcriptID64), nil
}

func parsePagination(r *http.Request) (page, limit int) {
	page = 1
	limit = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	return
}
