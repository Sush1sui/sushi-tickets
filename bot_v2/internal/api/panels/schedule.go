package panels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxScheduleWindowsPerPanel = 50

type errScheduleValidation struct {
	Field string
	Msg   string
}

func (e *errScheduleValidation) Error() string {
	return fmt.Sprintf("schedule validation failed: %s: %s", e.Field, e.Msg)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return false
	}
	return true
}

func toScheduleWindowView(w db.PanelScheduleWindow) ScheduleWindowView {
	return ScheduleWindowView{
		ID:            w.ID,
		PanelConfigID: w.PanelConfigID,
		StartTime:     w.StartTime,
		EndTime:       w.EndTime,
		ClosedEarly:   w.ClosedEarly,
		CreatedAt:     w.CreatedAt,
	}
}

func (h *Handler) panelExists(r *http.Request, serverID int64, panelID int32) (bool, error) {
	_, err := h.DB.GetPanelConfigByID(r.Context(), serverID, panelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateScheduleTimes(startTime, endTime int64) (string, string) {
	if startTime <= 0 || endTime <= 0 {
		return "times", "startTime and endTime are required unix timestamps"
	}
	now := time.Now().Unix()
	if endTime <= now {
		return "endTime", "endTime must be in the future"
	}
	if endTime <= startTime {
		return "endTime", "endTime must be after startTime"
	}
	return "", ""
}

func hasOverlap(windows []db.PanelScheduleWindow, startTime, endTime int64) bool {
	for _, w := range windows {
		if startTime < w.EndTime && w.StartTime < endTime {
			return true
		}
	}
	return false
}

// syncScheduleWindows persists the full desired window list for a panel in one
// transaction. nil schedule means "don't touch"; present (even empty) means
// make the DB match exactly. Rows keep their IDs; closedEarly is preserved or
// updated from the payload.
func syncScheduleWindows(ctx context.Context, pool *pgxpool.Pool, panelID int32, desired []ScheduleWindowSync) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	tq := db.New(tx)
	commit := func() error {
		if err := tx.Commit(ctx); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		return nil
	}
	rollback := func() {
		_ = tx.Rollback(ctx)
	}

	existingRows, err := tq.ListScheduleWindowsByPanel(ctx, panelID)
	if err != nil {
		rollback()
		return err
	}

	byID := make(map[int64]db.PanelScheduleWindow, len(existingRows))
	for _, row := range existingRows {
		byID[row.ID] = row
	}
	desiredByID := make(map[int64]ScheduleWindowSync, len(desired))
	for _, w := range desired {
		if w.ID != 0 {
			if _, ok := byID[w.ID]; !ok {
				rollback()
				return &errScheduleValidation{Field: "id", Msg: fmt.Sprintf("window %d does not exist", w.ID)}
			}
			desiredByID[w.ID] = w
		}
	}

	now := time.Now().Unix()
	for _, w := range desired {
		isNew := w.ID == 0
		if w.StartTime <= 0 || w.EndTime <= 0 || w.EndTime <= w.StartTime {
			rollback()
			return &errScheduleValidation{Field: "times", Msg: "startTime/endTime must be unix seconds with endTime after startTime"}
		}
		if isNew {
			if w.EndTime <= now {
				rollback()
				return &errScheduleValidation{Field: "endTime", Msg: "endTime must be in the future"}
			}
			continue
		}
		existing := byID[w.ID]
		if existing.StartTime == w.StartTime && existing.EndTime == w.EndTime {
			continue
		}
		if w.EndTime <= now {
			rollback()
			return &errScheduleValidation{Field: "endTime", Msg: "endTime must be in the future"}
		}
	}

	if len(desired) > maxScheduleWindowsPerPanel {
		rollback()
		return &errScheduleValidation{Field: "count", Msg: fmt.Sprintf("max %d schedule windows per panel", maxScheduleWindowsPerPanel)}
	}

	for _, row := range existingRows {
		if _, keep := desiredByID[row.ID]; !keep {
			if err := tq.DeleteScheduleWindow(ctx, row.ID); err != nil {
				rollback()
				return err
			}
		}
	}

	for _, w := range desired {
		if w.ID == 0 {
			if _, err := tq.CreateScheduleWindow(ctx, db.CreateScheduleWindowParams{
				PanelConfigID: panelID,
				StartTime:     w.StartTime,
				EndTime:       w.EndTime,
				ClosedEarly:   w.ClosedEarly,
				CreatedAt:     now,
			}); err != nil {
				rollback()
				return err
			}
			continue
		}
		existing := byID[w.ID]
		if existing.StartTime != w.StartTime || existing.EndTime != w.EndTime {
			if _, err := tq.UpdateScheduleWindow(ctx, db.UpdateScheduleWindowParams{
				ID:        w.ID,
				StartTime: w.StartTime,
				EndTime:   w.EndTime,
			}); err != nil {
				rollback()
				return err
			}
		}
		if existing.ClosedEarly != w.ClosedEarly {
			if err := tq.SetScheduleWindowClosedEarly(ctx, db.SetScheduleWindowClosedEarlyParams{
				ID:          w.ID,
				ClosedEarly: w.ClosedEarly,
			}); err != nil {
				rollback()
				return err
			}
		}
	}

	finalRows, err := tq.ListScheduleWindowsByPanel(ctx, panelID)
	if err != nil {
		rollback()
		return err
	}
	for i, a := range finalRows {
		if hasOverlap(finalRows[i+1:], a.StartTime, a.EndTime) {
			rollback()
			return &errScheduleValidation{Field: "overlap", Msg: "windows must not overlap each other"}
		}
	}

	return commit()
}

func (h *Handler) fetchOwnedWindow(w http.ResponseWriter, r *http.Request, serverID int64, panelID int32, windowID int64) (db.PanelScheduleWindow, bool) {
	win, err := h.DB.GetScheduleWindowByID(r.Context(), windowID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule window not found"})
			return db.PanelScheduleWindow{}, false
		}
		log.Printf("load schedule window failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load schedule window"})
		return db.PanelScheduleWindow{}, false
	}
	if win.PanelConfigID != panelID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule window not found"})
		return db.PanelScheduleWindow{}, false
	}
	if _, err := h.DB.GetPanelConfigByID(r.Context(), serverID, panelID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "panel not found"})
		} else {
			log.Printf("load panel failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel"})
		}
		return db.PanelScheduleWindow{}, false
	}
	return win, true
}

func (h *Handler) HandleListScheduleWindows(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}
	if ok, err := h.panelExists(r, serverID, panelID); err != nil {
		log.Printf("load panel failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel"})
		return
	} else if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "panel not found"})
		return
	}

	windows, err := h.DB.ListScheduleWindowsByPanel(r.Context(), panelID)
	if err != nil {
		log.Printf("list schedule windows failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list schedule windows"})
		return
	}
	items := make([]ScheduleWindowView, 0, len(windows))
	for _, win := range windows {
		items = append(items, toScheduleWindowView(win))
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleCreateScheduleWindow(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, err := utils.ParseServerPanelIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}
	if ok, err := h.panelExists(r, serverID, panelID); err != nil {
		log.Printf("load panel failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load panel"})
		return
	} else if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "panel not found"})
		return
	}

	var payload ScheduleWindowPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	if field, msg := validateScheduleTimes(payload.StartTime, payload.EndTime); field != "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": map[string]string{field: msg}})
		return
	}

	existing, err := h.DB.ListScheduleWindowsByPanel(r.Context(), panelID)
	if err != nil {
		log.Printf("list schedule windows failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list schedule windows"})
		return
	}
	if len(existing) >= maxScheduleWindowsPerPanel {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "max schedule windows reached"})
		return
	}
	if hasOverlap(existing, payload.StartTime, payload.EndTime) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "window overlaps an existing window"})
		return
	}

	win, err := h.DB.CreateScheduleWindow(r.Context(), db.CreateScheduleWindowParams{
		PanelConfigID: panelID,
		StartTime:     payload.StartTime,
		EndTime:       payload.EndTime,
		ClosedEarly:   false,
		CreatedAt:     time.Now().Unix(),
	})
	if err != nil {
		log.Printf("create schedule window failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create schedule window"})
		return
	}
	writeJSON(w, http.StatusCreated, toScheduleWindowView(win))
}

func (h *Handler) HandleUpdateScheduleWindow(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, windowID, err := utils.ParseServerPanelWindowIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}
	if _, ok := h.fetchOwnedWindow(w, r, serverID, panelID, windowID); !ok {
		return
	}

	var payload ScheduleWindowPayload
	if !decodeJSON(w, r, &payload) {
		return
	}
	if field, msg := validateScheduleTimes(payload.StartTime, payload.EndTime); field != "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "fields": map[string]string{field: msg}})
		return
	}

	existing, err := h.DB.ListScheduleWindowsByPanel(r.Context(), panelID)
	if err != nil {
		log.Printf("list schedule windows failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list schedule windows"})
		return
	}
	others := make([]db.PanelScheduleWindow, 0, len(existing))
	for _, ow := range existing {
		if ow.ID != windowID {
			others = append(others, ow)
		}
	}
	if hasOverlap(others, payload.StartTime, payload.EndTime) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "window overlaps an existing window"})
		return
	}

	win, err := h.DB.UpdateScheduleWindow(r.Context(), db.UpdateScheduleWindowParams{
		ID:        windowID,
		StartTime: payload.StartTime,
		EndTime:   payload.EndTime,
	})
	if err != nil {
		log.Printf("update schedule window failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update schedule window"})
		return
	}
	writeJSON(w, http.StatusOK, toScheduleWindowView(win))
}

func (h *Handler) HandleDeleteScheduleWindow(w http.ResponseWriter, r *http.Request) {
	serverID, panelID, windowID, err := utils.ParseServerPanelWindowIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}
	if _, ok := h.fetchOwnedWindow(w, r, serverID, panelID, windowID); !ok {
		return
	}
	if err := h.DB.DeleteScheduleWindow(r.Context(), windowID); err != nil {
		log.Printf("delete schedule window failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete schedule window"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) setScheduleWindowClosedEarly(w http.ResponseWriter, r *http.Request, closed bool) {
	serverID, panelID, windowID, err := utils.ParseServerPanelWindowIDs(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ids"})
		return
	}
	if _, ok := h.fetchOwnedWindow(w, r, serverID, panelID, windowID); !ok {
		return
	}
	if err := h.DB.SetScheduleWindowClosedEarly(r.Context(), db.SetScheduleWindowClosedEarlyParams{
		ID:          windowID,
		ClosedEarly: closed,
	}); err != nil {
		log.Printf("set schedule window closed failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update schedule window"})
		return
	}
	win, err := h.DB.GetScheduleWindowByID(r.Context(), windowID)
	if err != nil {
		log.Printf("load schedule window failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load schedule window"})
		return
	}
	writeJSON(w, http.StatusOK, toScheduleWindowView(win))
}

func (h *Handler) HandleCloseScheduleWindow(w http.ResponseWriter, r *http.Request) {
	h.setScheduleWindowClosedEarly(w, r, true)
}

func (h *Handler) HandleReopenScheduleWindow(w http.ResponseWriter, r *http.Request) {
	h.setScheduleWindowClosedEarly(w, r, false)
}
