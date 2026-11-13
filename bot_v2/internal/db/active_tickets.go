package db

import "context"

// ActiveTicketRow is the tracking metadata needed to restore auto-close and
// transcript capture after a restart.
type ActiveTicketRow struct {
	UserID    string
	ChannelID string
	CreatedAt int64
}

const listActiveTicketsByServer = `
SELECT user_id, channel_id, created_at
FROM active_ticket
WHERE server_config_id = $1
`

func (q *Queries) ListActiveTicketsByServer(ctx context.Context, serverConfigID int64) ([]ActiveTicketRow, error) {
	rows, err := q.db.Query(ctx, listActiveTicketsByServer, serverConfigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ActiveTicketRow, 0)
	for rows.Next() {
		var row ActiveTicketRow
		if err := rows.Scan(&row.UserID, &row.ChannelID, &row.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

const listAutoCloseServers = `
SELECT server_config_id FROM auto_close_config WHERE is_active = true
`

func (q *Queries) ListAutoCloseServers(ctx context.Context) ([]int64, error) {
	rows, err := q.db.Query(ctx, listAutoCloseServers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

const listTranscriptEnabledServers = `
SELECT id FROM server_config
WHERE ticket_transcript_cid IS NOT NULL AND ticket_transcript_cid <> ''
`

func (q *Queries) ListTranscriptEnabledServers(ctx context.Context) ([]int64, error) {
	rows, err := q.db.Query(ctx, listTranscriptEnabledServers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

const countActiveTicketsByUser = `
SELECT COUNT(*)
FROM active_ticket
WHERE server_config_id = $1 AND user_id = $2
`

func (q *Queries) CountActiveTicketsByUser(ctx context.Context, serverConfigID int64, userID string) (int64, error) {
	row := q.db.QueryRow(ctx, countActiveTicketsByUser, serverConfigID, userID)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const createActiveTicket = `
INSERT INTO active_ticket (server_config_id, user_id, channel_id, created_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (server_config_id, channel_id) DO NOTHING
`

func (q *Queries) CreateActiveTicket(ctx context.Context, serverConfigID int64, userID, channelID string, createdAt int64) error {
	_, err := q.db.Exec(ctx, createActiveTicket, serverConfigID, userID, channelID, createdAt)
	return err
}

const deleteActiveTicketByChannel = `
DELETE FROM active_ticket
WHERE server_config_id = $1 AND channel_id = $2
`

func (q *Queries) DeleteActiveTicketByChannel(ctx context.Context, serverConfigID int64, channelID string) error {
	_, err := q.db.Exec(ctx, deleteActiveTicketByChannel, serverConfigID, channelID)
	return err
}

const getActiveTicketChannelsByUser = `
SELECT channel_id
FROM active_ticket
WHERE server_config_id = $1 AND user_id = $2
`

func (q *Queries) GetActiveTicketChannelsByUser(ctx context.Context, serverConfigID int64, userID string) ([]string, error) {
	rows, err := q.db.Query(ctx, getActiveTicketChannelsByUser, serverConfigID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		items = append(items, channelID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) GetActiveTicketChannelsByServer(ctx context.Context, serverConfigID int64) ([]string, error) {
	rows, err := q.db.Query(ctx, `SELECT channel_id FROM active_ticket WHERE server_config_id = $1`, serverConfigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	channels := make([]string, 0)
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		channels = append(channels, channelID)
	}
	return channels, rows.Err()
}

const getActiveTicketByChannel = `
SELECT user_id, created_at
FROM active_ticket
WHERE server_config_id = $1 AND channel_id = $2
LIMIT 1
`

type ActiveTicketInfo struct {
	UserID    string
	CreatedAt int64
}

func (q *Queries) GetActiveTicketByChannel(ctx context.Context, serverConfigID int64, channelID string) (ActiveTicketInfo, error) {
	row := q.db.QueryRow(ctx, getActiveTicketByChannel, serverConfigID, channelID)
	var info ActiveTicketInfo
	if err := row.Scan(&info.UserID, &info.CreatedAt); err != nil {
		return ActiveTicketInfo{}, err
	}
	return info, nil
}
