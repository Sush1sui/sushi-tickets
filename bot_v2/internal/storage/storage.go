// internal/storage/storage.go
package storage

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound is returned when no transcript file exists for a key.
var ErrNotFound = errors.New("transcript not found")

// Client stores transcript JSON as files under a base directory.
// Storage keys (transcripts/<server>/<channel>/<ts>.json) map 1:1 to paths.
type Client struct {
	dir string
}

// NewFileClient resolves the transcript dir from TRANSCRIPT_DIR
// (/var/lib/sushi-tickets-transcripts fallback) and ensures it exists.
func NewFileClient() *Client {
	dir := os.Getenv("TRANSCRIPT_DIR")
	if dir == "" {
		dir = "/var/lib/sushi-tickets-transcripts"
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Fatalf("Failed to create transcript dir %s: %v", dir, err)
	}
	return &Client{dir: dir}
}

// NewClientAt returns a client rooted at dir. Used by tests and tooling.
func NewClientAt(dir string) *Client {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Fatalf("Failed to create transcript dir %s: %v", dir, err)
	}
	return &Client{dir: dir}
}

// BaseDir exposes the storage root for adjacent local state such as the
// gateway message spool.
func (c *Client) BaseDir() string {
	return c.dir
}

// keyPath maps a storage key to a file under the base dir.
// Clean resolves ".." lexically and Join keeps cleaned absolute paths
// under dir, so a key can never escape the base directory.
func (c *Client) keyPath(key string) (string, error) {
	clean := filepath.Clean("/" + key)
	if clean == "/" || strings.Contains(clean, "..") {
		return "", errors.New("invalid storage key")
	}
	return filepath.Join(c.dir, clean), nil
}

// UploadTranscript writes transcript JSON atomically (tmp + rename).
func (c *Client) UploadTranscript(_ context.Context, key string, data []byte) error {
	p, err := c.keyPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	// A transcript must be durable before the bot deletes its Discord channel.
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// DownloadTranscript reads transcript JSON for a key.
func (c *Client) DownloadTranscript(_ context.Context, key string) ([]byte, error) {
	p, err := c.keyPath(key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}
