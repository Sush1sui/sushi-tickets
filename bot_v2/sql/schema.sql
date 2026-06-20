-- sql/schema.sql

CREATE TABLE server_config (
    id BIGINT PRIMARY KEY,
    ticket_name_style TEXT NOT NULL,
    ticket_transcript_cid TEXT,
    max_ticket_per_user INTEGER NOT NULL DEFAULT 2,
    ticket_permissions JSONB,
    max_panel INTEGER DEFAULT 3,
    max_multi_panel INTEGER DEFAULT 3
);

CREATE TABLE auto_close_config (
    server_config_id BIGINT NOT NULL UNIQUE REFERENCES server_config(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT false,
    close_on_user_leave BOOLEAN NOT NULL DEFAULT false,
    close_since_open_with_no_response_mins INTEGER,
    close_since_last_message_mins INTEGER,
    PRIMARY KEY (server_config_id)
);

CREATE TABLE panel_config (
    id SERIAL PRIMARY KEY,
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    mention_roles_on_open TEXT[],
    category_id TEXT,
    title TEXT NOT NULL,
    content TEXT,
    embed_color INTEGER NOT NULL,
    channel_id TEXT NOT NULL,
    btn_color TEXT NOT NULL,
    btn_txt TEXT NOT NULL,
    btn_emoji TEXT,
    large_img_url TEXT,
    small_img_url TEXT
);

CREATE TABLE questions_config (
    id SERIAL PRIMARY KEY,
    panel_config_id INTEGER NOT NULL REFERENCES panel_config(id) ON DELETE CASCADE,
    questions TEXT[],
    required_flags BOOLEAN[] NOT NULL DEFAULT '{}'
);

CREATE TABLE welcome_msg_config (
    id SERIAL PRIMARY KEY,
    panel_config_id INTEGER NOT NULL REFERENCES panel_config(id) ON DELETE CASCADE,
    embed_color INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    title_url TEXT,
    large_img_url TEXT,
    small_img_url TEXT,
    footer TEXT,
    footer_icon_url TEXT
);

CREATE TABLE multi_panel_config (
    id SERIAL PRIMARY KEY,
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT,
    embed_color INTEGER NOT NULL,
    channel_id TEXT NOT NULL,
    large_img_url TEXT,
    small_img_url TEXT,
    use_dropdown BOOLEAN NOT NULL,
    panel_config_ids INTEGER[],
    footer TEXT,
    foot_icon_url TEXT
);

CREATE TABLE transcript (
    id SERIAL PRIMARY KEY,
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    ticket_id TEXT,
    username TEXT,
    user_id TEXT,
    opened_at BIGINT NOT NULL,
    closed_at BIGINT NOT NULL,
    closed_by TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    total_messages INTEGER DEFAULT 0,
    total_attachments INTEGER DEFAULT 0,
    total_embeds INTEGER DEFAULT 0
);

CREATE TABLE active_ticket (
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    PRIMARY KEY (server_config_id, channel_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_config_server_id ON panel_config (server_config_id, id);
CREATE INDEX IF NOT EXISTS idx_questions_config_panel_id ON questions_config (panel_config_id);
CREATE INDEX IF NOT EXISTS idx_welcome_msg_panel_id ON welcome_msg_config (panel_config_id);
CREATE INDEX IF NOT EXISTS idx_multi_panel_config_server_id ON multi_panel_config (server_config_id, id);
CREATE INDEX IF NOT EXISTS idx_multi_panel_panel_ids_gin ON multi_panel_config USING GIN (panel_config_ids);
CREATE INDEX IF NOT EXISTS idx_transcript_server_closed_at ON transcript (server_config_id, closed_at DESC);
CREATE INDEX IF NOT EXISTS idx_active_ticket_user ON active_ticket (server_config_id, user_id);

CREATE TABLE authorized_members (
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL,
    PRIMARY KEY (server_config_id, member_id)
);

CREATE TABLE authorized_roles (
    server_config_id BIGINT NOT NULL REFERENCES server_config(id) ON DELETE CASCADE,
    role_id TEXT NOT NULL,
    PRIMARY KEY (server_config_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_am_member ON authorized_members (member_id);
CREATE INDEX IF NOT EXISTS idx_ar_role ON authorized_roles (role_id);

CREATE TABLE auth_session (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    discord_id TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT,
    image TEXT,
    access_token TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    last_seen BIGINT NOT NULL
);

CREATE TABLE idempotency_key (
    key TEXT NOT NULL,
    user_id TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_code INTEGER NOT NULL,
    response_body BYTEA NOT NULL,
    response_content_type TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    PRIMARY KEY (key, user_id)
);
