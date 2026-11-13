package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// TranscriptRef identifies the newest transcript row for a ticket.
type TranscriptRef struct {
	ID         int32
	StorageKey string
}

const getLatestTranscriptByTicket = `
SELECT id, storage_key
FROM transcript
WHERE server_config_id = $1 AND ticket_id = $2
ORDER BY id DESC
LIMIT 1
`

// GetLatestTranscriptByTicket reuses an existing row instead of deleting or
// inserting duplicates. It reports pgx.ErrNoRows when no row exists yet.
func (q *Queries) GetLatestTranscriptByTicket(ctx context.Context, serverConfigID int64, ticketID string) (TranscriptRef, error) {
	row := q.db.QueryRow(ctx, getLatestTranscriptByTicket, serverConfigID, ticketID)
	var ref TranscriptRef
	if err := row.Scan(&ref.ID, &ref.StorageKey); err != nil {
		return TranscriptRef{}, err
	}
	return ref, nil
}

const updateTranscriptByID = `
UPDATE transcript SET
    username = $2,
    user_id = $3,
    opened_at = $4,
    closed_at = $5,
    closed_by = $6,
    storage_key = $7,
    total_messages = $8,
    total_attachments = $9,
    total_embeds = $10
WHERE id = $1 AND server_config_id = $11
`

type UpdateTranscriptByIDParams struct {
	ID               int32
	Username         pgtype.Text
	UserID           pgtype.Text
	OpenedAt         int64
	ClosedAt         int64
	ClosedBy         string
	StorageKey       string
	TotalMessages    pgtype.Int4
	TotalAttachments pgtype.Int4
	TotalEmbeds      pgtype.Int4
	ServerConfigID   int64
}

func (q *Queries) UpdateTranscriptByID(ctx context.Context, arg UpdateTranscriptByIDParams) error {
	_, err := q.db.Exec(ctx, updateTranscriptByID,
		arg.ID,
		arg.Username,
		arg.UserID,
		arg.OpenedAt,
		arg.ClosedAt,
		arg.ClosedBy,
		arg.StorageKey,
		arg.TotalMessages,
		arg.TotalAttachments,
		arg.TotalEmbeds,
		arg.ServerConfigID,
	)
	return err
}
