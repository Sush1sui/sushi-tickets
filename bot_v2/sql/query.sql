-- name: GetServerConfig :one
SELECT * FROM server_config 
WHERE id = $1 LIMIT 1;

-- name: UpsertServerConfig :one
INSERT INTO server_config (
    id, ticket_name_style, ticket_transcript_cid, max_ticket_per_user, 
    ticket_permissions, max_panel, max_multi_panel
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id) DO UPDATE SET
    ticket_name_style = EXCLUDED.ticket_name_style,
    ticket_transcript_cid = EXCLUDED.ticket_transcript_cid,
    max_ticket_per_user = EXCLUDED.max_ticket_per_user,
    ticket_permissions = EXCLUDED.ticket_permissions,
    max_panel = EXCLUDED.max_panel,
    max_multi_panel = EXCLUDED.max_multi_panel
RETURNING *;

-- name: GetAuthorizedMembers :many
SELECT member_id FROM authorized_members
WHERE server_config_id = $1;

-- name: GetAuthorizedRoles :many
SELECT role_id FROM authorized_roles
WHERE server_config_id = $1;

-- name: IsMemberAuthorized :one
SELECT EXISTS(SELECT 1 FROM authorized_members WHERE server_config_id = $1 AND member_id = $2);

-- name: UpsertAuthorizedMember :exec
INSERT INTO authorized_members (server_config_id, member_id)
VALUES ($1, $2)
ON CONFLICT (server_config_id, member_id) DO NOTHING;

-- name: DeleteAuthorizedMember :exec
DELETE FROM authorized_members WHERE server_config_id = $1 AND member_id = $2;

-- name: ClearAuthorizedMembers :exec
DELETE FROM authorized_members WHERE server_config_id = $1;

-- name: UpsertAuthorizedRole :exec
INSERT INTO authorized_roles (server_config_id, role_id)
VALUES ($1, $2)
ON CONFLICT (server_config_id, role_id) DO NOTHING;

-- name: DeleteAuthorizedRole :exec
DELETE FROM authorized_roles WHERE server_config_id = $1 AND role_id = $2;

-- name: ClearAuthorizedRoles :exec
DELETE FROM authorized_roles WHERE server_config_id = $1;

-- name: GetServersByMember :many
SELECT server_config_id FROM authorized_members WHERE member_id = $1;

-- name: GetServersByRole :many
SELECT server_config_id FROM authorized_roles WHERE role_id = $1;

-- name: GetAutoCloseConfig :one
SELECT * FROM auto_close_config
WHERE server_config_id = $1 LIMIT 1;

-- name: UpsertAutoCloseConfig :one
INSERT INTO auto_close_config (
    server_config_id, is_active, close_on_user_leave,
    close_since_open_with_no_response_mins, close_since_last_message_mins
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (server_config_id) DO UPDATE SET
    is_active = EXCLUDED.is_active,
    close_on_user_leave = EXCLUDED.close_on_user_leave,
    close_since_open_with_no_response_mins = EXCLUDED.close_since_open_with_no_response_mins,
    close_since_last_message_mins = EXCLUDED.close_since_last_message_mins
RETURNING *;

-- name: GetTranscriptsByServer :many
SELECT id, server_config_id, ticket_id, username, user_id,
       opened_at, closed_at, closed_by, storage_key,
       total_messages, total_attachments, total_embeds
FROM transcript
WHERE server_config_id = $1
ORDER BY closed_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTranscriptsByServer :one
SELECT COUNT(*) FROM transcript WHERE server_config_id = $1;

-- name: GetTranscriptByID :one
SELECT * FROM transcript WHERE id = $1 AND server_config_id = $2;

-- name: CreateTranscript :one
INSERT INTO transcript (
    server_config_id, ticket_id, username, user_id,
    opened_at, closed_at, closed_by, storage_key,
    total_messages, total_attachments, total_embeds
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetActiveScheduleWindow :one
SELECT id, panel_config_id, start_time, end_time, closed_early, created_at
FROM panel_schedule_window
WHERE panel_config_id = $1
  AND start_time <= $2
  AND end_time > $2
  AND closed_early = false
ORDER BY start_time DESC
LIMIT 1;

-- name: GetNextScheduleWindow :one
SELECT id, panel_config_id, start_time, end_time, closed_early, created_at
FROM panel_schedule_window
WHERE panel_config_id = $1
  AND start_time > $2
ORDER BY start_time ASC
LIMIT 1;

-- name: HasAnyScheduleWindow :one
SELECT EXISTS(SELECT 1 FROM panel_schedule_window WHERE panel_config_id = $1);

-- name: ListScheduleWindowsByPanel :many
SELECT id, panel_config_id, start_time, end_time, closed_early, created_at
FROM panel_schedule_window
WHERE panel_config_id = $1
ORDER BY start_time ASC;

-- name: GetScheduleWindowByID :one
SELECT id, panel_config_id, start_time, end_time, closed_early, created_at
FROM panel_schedule_window
WHERE id = $1;

-- name: CreateScheduleWindow :one
INSERT INTO panel_schedule_window (panel_config_id, start_time, end_time, closed_early, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateScheduleWindow :one
UPDATE panel_schedule_window
SET start_time = $2, end_time = $3
WHERE id = $1
RETURNING *;

-- name: DeleteScheduleWindow :exec
DELETE FROM panel_schedule_window WHERE id = $1;

-- name: DeleteScheduleWindowsByPanel :exec
DELETE FROM panel_schedule_window WHERE panel_config_id = $1;

-- name: SetScheduleWindowClosedEarly :exec
UPDATE panel_schedule_window SET closed_early = $2 WHERE id = $1;
