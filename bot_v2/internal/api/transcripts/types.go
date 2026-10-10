package transcripts

import (
	"context"

	"github.com/Sush1sui/FNS_BOT/internal/db"
)

type Handler struct {
	DB      *db.Queries
	Storage interface {
		DownloadTranscript(ctx context.Context, key string) ([]byte, error)
	}
}