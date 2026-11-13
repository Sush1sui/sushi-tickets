package tickets

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// The spool is an append-only local event log per ticket. It is the only
// source used to rebuild transcripts and observed activity. It never claims
// to know about events the gateway did not deliver.
type spoolHeader struct {
	T            string `json:"t"`
	Server       int64  `json:"server"`
	Channel      string `json:"channel"`
	Opener       string `json:"opener"`
	Created      int64  `json:"created"`
	CaptureStart int64  `json:"capture_start"`
	Uncertain    bool   `json:"uncertain"`
	Reason       string `json:"reason,omitempty"`
}

type spoolAuthor struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot"`
}

type spoolMessage struct {
	ID          string                         `json:"id"`
	Author      spoolAuthor                    `json:"author"`
	Content     string                         `json:"content"`
	Timestamp   int64                          `json:"ts"`
	Edited      *int64                         `json:"edited,omitempty"`
	Embeds      []*discordgo.MessageEmbed      `json:"embeds,omitempty"`
	Attachments []*discordgo.MessageAttachment `json:"attachments,omitempty"`
}

type spoolEvent struct {
	T      string        `json:"t"`
	M      *spoolMessage `json:"m,omitempty"`
	Kind   string        `json:"kind,omitempty"`
	ID     string        `json:"id,omitempty"`
	IDs    []string      `json:"ids,omitempty"`
	MID    string        `json:"mid,omitempty"`
	Emoji  string        `json:"emoji,omitempty"`
	User   string        `json:"user,omitempty"`
	TS     int64         `json:"ts,omitempty"`
	Reason string        `json:"reason,omitempty"`
}

type spoolState struct {
	flushMu sync.Mutex
	path    string
	server  int64
	buffer  [][]byte
	seq     uint64
	failed  error
	dirty   bool
}

var spools = struct {
	sync.Mutex
	m map[string]*spoolState
}{m: map[string]*spoolState{}}

func spoolDir() (string, error) {
	if storageClient == nil {
		return "", errors.New("storage not configured")
	}
	base := storageClient.BaseDir()
	if base == "" {
		return "", errors.New("storage directory unavailable")
	}
	return filepath.Join(base, "spool"), nil
}

func spoolPath(server int64, channel string) (string, error) {
	dir, err := spoolDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, strconv.FormatInt(server, 10), channel+".jsonl"), nil
}

func spoolStart(server int64, channel, opener string, created, captureStart int64, uncertain bool, reason string) error {
	spools.Lock()
	defer spools.Unlock()
	if st, ok := spools.m[channel]; ok && st != nil {
		return nil
	}
	path, err := spoolPath(server, channel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	header := spoolHeader{
		T: "h", Server: server, Channel: channel, Opener: opener,
		Created: created, CaptureStart: captureStart, Uncertain: uncertain, Reason: reason,
	}
	line, err := json.Marshal(header)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	} else {
		if _, err := f.Write(append(line, '\n')); err != nil {
			f.Close()
			return err
		}
		if err := f.Sync(); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	spools.m[channel] = &spoolState{path: path, server: server}
	return nil
}

func spoolAppend(server int64, channel string, event spoolEvent) error {
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	spools.Lock()
	defer spools.Unlock()
	st, ok := spools.m[channel]
	if !ok {
		path, perr := spoolPath(server, channel)
		if perr != nil {
			return perr
		}
		st = &spoolState{path: path, server: server}
		spools.m[channel] = st
	}
	st.buffer = append(st.buffer, append(line, '\n'))
	st.seq++
	st.dirty = true
	return nil
}

// spoolFlush writes buffered events. Sync is used only at close time so the
// transcript reads durable bytes before the channel is deleted.
func spoolFlush(channel string, syncWrite bool) (flushErr error) {
	spools.Lock()
	st, ok := spools.m[channel]
	spools.Unlock()
	if !ok {
		return nil
	}
	st.flushMu.Lock()
	defer st.flushMu.Unlock()
	spools.Lock()
	if spools.m[channel] != st || (!syncWrite && len(st.buffer) == 0 && st.failed == nil) {
		spools.Unlock()
		return nil
	}
	lines := st.buffer
	st.buffer = nil
	st.dirty = false
	path := st.path
	spools.Unlock()
	defer func() {
		spools.Lock()
		defer spools.Unlock()
		// Keep the unwritten suffix ahead of events received during this flush.
		st.buffer = append(lines, st.buffer...)
		st.failed = flushErr
		st.dirty = len(st.buffer) > 0 || flushErr != nil
	}()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	for len(lines) > 0 {
		n, err := f.Write(lines[0])
		if n == len(lines[0]) {
			lines = lines[1:]
		} else {
			lines[0] = lines[0][n:]
			if err == nil {
				err = io.ErrShortWrite
			}
		}
		if err != nil {
			f.Close()
			return err
		}
	}
	if syncWrite {
		if err := f.Sync(); err != nil {
			f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func spoolFailed(channel string) bool {
	spools.Lock()
	defer spools.Unlock()
	st, ok := spools.m[channel]
	return ok && st.failed != nil
}

func spoolSeq(channel string) uint64 {
	spools.Lock()
	defer spools.Unlock()
	if st, ok := spools.m[channel]; ok {
		return st.seq
	}
	return 0
}

func dirtySpoolChannels() []string {
	spools.Lock()
	defer spools.Unlock()
	out := make([]string, 0)
	for id, st := range spools.m {
		if st.dirty || st.failed != nil {
			out = append(out, id)
		}
	}
	return out
}

func spoolDelete(channel string) {
	spools.Lock()
	st := spools.m[channel]
	spools.Unlock()
	if st == nil {
		return
	}
	st.flushMu.Lock()
	defer st.flushMu.Unlock()
	spools.Lock()
	delete(spools.m, channel)
	spools.Unlock()
	_ = os.Remove(st.path)
}

func spoolExists(server int64, channel string) bool {
	path, err := spoolPath(server, channel)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// spoolMetadata is a streaming summary for restart hydration. It deliberately
// keeps no message bodies: full history is only materialized at close time.
type spoolMetadata struct {
	exists       bool
	responded    bool
	lastActivity time.Time
	gap          bool
	uncertain    bool
	captureStart time.Time
}

func spoolScanMetadata(server int64, channel string) (spoolMetadata, error) {
	path, err := spoolPath(server, channel)
	if err != nil {
		return spoolMetadata{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return spoolMetadata{}, nil
		}
		return spoolMetadata{}, err
	}
	defer f.Close()

	meta := spoolMetadata{exists: true}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var header spoolHeader
		if err := json.Unmarshal(line, &header); err == nil && header.T == "h" {
			meta.uncertain = header.Uncertain
			meta.gap = header.Uncertain
			if header.CaptureStart > 0 {
				meta.captureStart = time.Unix(header.CaptureStart, 0)
			}
			continue
		}
		var ev spoolEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		switch ev.T {
		case "m":
			if ev.M == nil {
				continue
			}
			if ev.M.Author.Bot == false && ev.M.Author.ID != "" {
				meta.responded = true
			}
			if ev.M.Timestamp > 0 {
				ts := time.UnixMilli(ev.M.Timestamp)
				if ts.After(meta.lastActivity) {
					meta.lastActivity = ts
				}
			}
		case "g":
			meta.gap = true
			meta.uncertain = true
		}
	}
	return meta, scanner.Err()
}

type spoolReplay struct {
	exists       bool
	header       spoolHeader
	messages     []*discordgo.Message
	responded    bool
	lastActivity time.Time
	gap          bool
	uncertain    bool
	captureStart time.Time
}

func spoolReplayTicket(server int64, channel string) (spoolReplay, error) {
	spools.Lock()
	st := spools.m[channel]
	spools.Unlock()
	if st != nil {
		// Replay must not read a partially written event from a concurrent flush.
		st.flushMu.Lock()
		defer st.flushMu.Unlock()
	}
	path, err := spoolPath(server, channel)
	if err != nil {
		return spoolReplay{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return spoolReplay{}, nil
		}
		return spoolReplay{}, err
	}
	defer f.Close()

	replay := spoolReplay{exists: true}
	messages := make(map[string]*spoolMessage)
	reactions := make(map[string]map[string]map[string]struct{})
	order := make([]string, 0)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var header spoolHeader
		if err := json.Unmarshal(line, &header); err == nil && header.T == "h" {
			replay.header = header
			replay.uncertain = header.Uncertain
			replay.gap = header.Uncertain
			if header.CaptureStart > 0 {
				replay.captureStart = time.Unix(header.CaptureStart, 0)
			}
			continue
		}
		var ev spoolEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		switch ev.T {
		case "m":
			if ev.M == nil || ev.M.ID == "" {
				continue
			}
			msg := ev.M
			if ev.Kind == "update" {
				existing, ok := messages[msg.ID]
				if !ok {
					messages[msg.ID] = msg
					order = append(order, msg.ID)
				} else {
					if msg.Content != "" {
						existing.Content = msg.Content
					}
					if msg.Embeds != nil {
						existing.Embeds = msg.Embeds
					}
					if msg.Attachments != nil {
						existing.Attachments = msg.Attachments
					}
					if msg.Edited != nil {
						existing.Edited = msg.Edited
					}
				}
			} else {
				if _, ok := messages[msg.ID]; !ok {
					order = append(order, msg.ID)
				}
				messages[msg.ID] = msg
			}
			if msg.Author.Bot == false && msg.Author.ID != "" {
				replay.responded = true
			}
			if msg.Timestamp > 0 {
				ts := time.UnixMilli(msg.Timestamp)
				if ts.After(replay.lastActivity) {
					replay.lastActivity = ts
				}
			}
		case "d":
			delete(messages, ev.ID)
			delete(reactions, ev.ID)
		case "x":
			for _, id := range ev.IDs {
				delete(messages, id)
				delete(reactions, id)
			}
		case "ra":
			if ev.MID == "" || ev.Emoji == "" || ev.User == "" {
				continue
			}
			byEmoji := reactions[ev.MID]
			if byEmoji == nil {
				byEmoji = make(map[string]map[string]struct{})
				reactions[ev.MID] = byEmoji
			}
			users := byEmoji[ev.Emoji]
			if users == nil {
				users = make(map[string]struct{})
				byEmoji[ev.Emoji] = users
			}
			users[ev.User] = struct{}{}
		case "rr":
			if byEmoji := reactions[ev.MID]; byEmoji != nil {
				if users := byEmoji[ev.Emoji]; users != nil {
					delete(users, ev.User)
					if len(users) == 0 {
						delete(byEmoji, ev.Emoji)
					}
				}
			}
		case "rc":
			delete(reactions, ev.MID)
		case "g":
			replay.gap = true
			replay.uncertain = true
		}
	}
	if err := scanner.Err(); err != nil {
		return replay, err
	}

	sort.Slice(order, func(i, j int) bool {
		a, errA := strconv.ParseUint(order[i], 10, 64)
		b, errB := strconv.ParseUint(order[j], 10, 64)
		if errA == nil && errB == nil {
			return a < b
		}
		return order[i] < order[j]
	})
	for _, id := range order {
		msg, ok := messages[id]
		if !ok {
			continue
		}
		out := msg.toDiscordMessage()
		if byEmoji := reactions[id]; len(byEmoji) > 0 {
			names := make([]string, 0, len(byEmoji))
			for emoji := range byEmoji {
				names = append(names, emoji)
			}
			sort.Strings(names)
			for _, emoji := range names {
				out.Reactions = append(out.Reactions, &discordgo.MessageReactions{
					Emoji: &discordgo.Emoji{Name: emoji},
					Count: len(byEmoji[emoji]),
				})
			}
		}
		replay.messages = append(replay.messages, out)
	}
	return replay, nil
}

func (m *spoolMessage) toDiscordMessage() *discordgo.Message {
	msg := &discordgo.Message{
		ID:          m.ID,
		Content:     m.Content,
		Timestamp:   time.UnixMilli(m.Timestamp),
		Embeds:      m.Embeds,
		Attachments: m.Attachments,
		Author: &discordgo.User{
			ID:            m.Author.ID,
			Username:      m.Author.Username,
			Discriminator: m.Author.Discriminator,
			Avatar:        m.Author.Avatar,
			Bot:           m.Author.Bot,
		},
	}
	if m.Edited != nil {
		ts := time.UnixMilli(*m.Edited)
		msg.EditedTimestamp = &ts
	}
	return msg
}

func spoolMessageFromDiscord(m *discordgo.Message, kind string) *spoolMessage {
	if m == nil || m.ID == "" {
		return nil
	}
	out := &spoolMessage{
		ID:          m.ID,
		Content:     m.Content,
		Timestamp:   m.Timestamp.UnixMilli(),
		Embeds:      m.Embeds,
		Attachments: m.Attachments,
	}
	if m.Author != nil {
		out.Author = spoolAuthor{
			ID:            m.Author.ID,
			Username:      m.Author.Username,
			Discriminator: m.Author.Discriminator,
			Avatar:        m.Author.Avatar,
			Bot:           m.Author.Bot,
		}
	}
	if m.EditedTimestamp != nil {
		ts := m.EditedTimestamp.UnixMilli()
		out.Edited = &ts
	}
	return out
}

// cleanupMarker is local recovery state, not database schema.
type cleanupMarker struct {
	Server    int64  `json:"server"`
	Channel   string `json:"channel"`
	Mode      string `json:"mode"`
	State     string `json:"state"`
	Deleted   bool   `json:"deleted"`
	RowID     int32  `json:"row_id,omitempty"`
	Logged    bool   `json:"logged,omitempty"`
	ClosedBy  string `json:"closed_by,omitempty"`
	Opener    string `json:"opener,omitempty"`
	OpenedAt  int64  `json:"opened_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func markerPath(server int64, channel string) (string, error) {
	dir, err := spoolDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cleanup", fmt.Sprintf("%d__%s.json", server, channel)), nil
}

func writeMarker(m cleanupMarker) error {
	path, err := markerPath(m.Server, m.Channel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	m.UpdatedAt = time.Now().Unix()
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func deleteMarker(server int64, channel string) {
	path, err := markerPath(server, channel)
	if err == nil {
		_ = os.Remove(path)
	}
}

func listMarkers() ([]cleanupMarker, error) {
	dir, err := spoolDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, "cleanup"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]cleanupMarker, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, "cleanup", entry.Name()))
		if err != nil {
			continue
		}
		var m cleanupMarker
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}
