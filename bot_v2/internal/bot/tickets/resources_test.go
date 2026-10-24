package tickets

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// Benchmarks exclude Discord/DB I/O; they measure the queue's own overhead.
func BenchmarkQueueWaiting500(b *testing.B) {
	writer := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(writer)
	q := newTicketQueue(500, nil, func(*ticketJob, string) {})
	defer q.cancel()
	keys := make([]string, 500)
	for n := range keys {
		keys[n] = fmt.Sprint(n)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, key := range keys {
			if err := q.enqueue(jobFor(key)); err != nil {
				b.Fatal(err)
			}
		}
		q.mu.Lock()
		for _, j := range q.waiting {
			j.timer.Stop()
			j.cancel()
		}
		clear(q.pending)
		clear(q.waiting)
		q.waiting = q.waiting[:0]
		q.mu.Unlock()
	}
}

func BenchmarkQuestionCacheHit(b *testing.B) {
	questions := []QuestionData{{Label: strings.Repeat("question", 20), Style: "paragraph", IsRequired: true}}
	c := questionCache{entries: make(map[int32]questionSnapshot), load: func(context.Context, int32) (questionSnapshot, error) {
		return questionSnapshot{title: "Ticket questions", questions: questions}, nil
	}}
	ctx := context.Background()
	c.get(ctx, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		value, err := c.get(ctx, 1)
		if err != nil {
			b.Fatal(err)
		}
		_ = value.version
	}
}

func TestQueueRetainedMemory(t *testing.T) {
	writer := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(writer)
	q := newTicketQueue(500, nil, func(*ticketJob, string) {})
	defer func() {
		q.mu.Lock()
		for _, j := range q.waiting {
			j.timer.Stop()
			j.cancel()
		}
		q.mu.Unlock()
		q.cancel()
	}()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for n := 0; n < 500; n++ {
		i := interaction()
		i.Member.Roles = make([]string, 50)
		i.Message = &discordgo.Message{Content: strings.Repeat("panel message", 1000)}
		j := jobFor(fmt.Sprint(n))
		j.interaction = i
		j.qna = []QnA{{Question: "Question", Answer: strings.Repeat("a", 4000)}}
		if err := q.enqueue(j); err != nil {
			t.Fatal(err)
		}
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	t.Logf("retained heap for 500 waiting jobs, each with a 4,000-byte answer: %.2f MiB", float64(int64(after.HeapAlloc)-int64(before.HeapAlloc))/(1024*1024))
	runtime.KeepAlive(q)
}

func TestOpeningInteractionKeepsOnlyTicketFields(t *testing.T) {
	original := interaction()
	original.ChannelID = "source-channel"
	original.Member.Roles = []string{"staff"}
	original.Message = &discordgo.Message{Content: "large source message"}
	original.Data = discordgo.ModalSubmitInteractionData{CustomID: "form"}
	compact := openingInteraction(original)
	if compact == original || compact.Interaction == original.Interaction || compact.Member == original.Member {
		t.Fatal("queue kept original payload")
	}
	if compact.ID != original.ID || compact.AppID != original.AppID || compact.Token != original.Token || compact.GuildID != original.GuildID || compact.ChannelID != original.ChannelID || compact.Member.User.ID != original.Member.User.ID || compact.Member.User.Username != original.Member.User.Username {
		t.Fatal("ticket identity, channel naming, or reply credentials changed")
	}
	if compact.Message != nil || compact.Data != nil || compact.Member.Roles != nil {
		t.Fatal("unused payload retained")
	}
	if original.Message == nil || original.Data == nil || len(original.Member.Roles) != 1 {
		t.Fatal("compaction mutated original event")
	}
}

func TestQueueReleasesAnswersAndReusesTimer(t *testing.T) {
	q := newTicketQueue(1, nil, func(*ticketJob, string) {})
	t.Cleanup(func() { stopTestQueue(t, q) })
	j := jobFor("user")
	j.qna = []QnA{{Answer: strings.Repeat("x", 4000)}}
	if err := q.enqueue(j); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	timer := j.timer
	q.armLocked(j)
	if timer != j.timer {
		t.Error("schedule refresh allocated another timer")
	}
	q.cancelLocked(j, windowCanceledMessage)
	if j.qna != nil {
		t.Error("canceled job retained answers for notification")
	}
	q.mu.Unlock()
}

func TestCachedQuestionVersionMatchesExistingForms(t *testing.T) {
	questions := []QuestionData{{Label: "Question", Style: "paragraph", IsRequired: true}}
	c := questionCache{entries: make(map[int32]questionSnapshot), load: func(context.Context, int32) (questionSnapshot, error) {
		return questionSnapshot{title: "Title", questions: questions}, nil
	}}
	value, err := c.get(context.Background(), 1)
	if err != nil || value.version != questionVersion("Title", questions) {
		t.Fatal("cached version changed form compatibility")
	}
}
