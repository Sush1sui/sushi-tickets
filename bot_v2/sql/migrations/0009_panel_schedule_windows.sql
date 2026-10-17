-- Scheduled ticket windows: per-panel start/end windows that gate ticket opening.
-- A panel with no windows is always open (backward compatible).
-- Active window: start_time <= now AND end_time > now AND closed_early = false.
-- Times are unix seconds (UTC), consistent with the rest of the schema.

CREATE TABLE IF NOT EXISTS panel_schedule_window (
    id BIGSERIAL PRIMARY KEY,
    panel_config_id INTEGER NOT NULL REFERENCES panel_config(id) ON DELETE CASCADE,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    closed_early BOOLEAN NOT NULL DEFAULT false,
    created_at BIGINT NOT NULL,
    CHECK (end_time > start_time)
);

-- Hot path: ticket-open click looks up the active/last window for one panel.
CREATE INDEX IF NOT EXISTS idx_psw_panel_start ON panel_schedule_window (panel_config_id, start_time DESC);
-- Dashboard/scheduler scans: upcoming and closing-soon windows.
CREATE INDEX IF NOT EXISTS idx_psw_start ON panel_schedule_window (start_time);
CREATE INDEX IF NOT EXISTS idx_psw_end ON panel_schedule_window (end_time) WHERE closed_early = false;

-- Indexing audit fixes: both cleanup queries below are full table scans today.
-- Expired-session cleanup runs every 4h (session_cleanup.go).
CREATE INDEX IF NOT EXISTS idx_auth_session_expires_at ON auth_session (expires_at);
-- Expired idempotency record cleanup runs on a background interval as well.
CREATE INDEX IF NOT EXISTS idx_idempotency_expires_at ON idempotency_key (expires_at);
