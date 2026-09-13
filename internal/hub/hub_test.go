package hub

import (
	"testing"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

func TestPausedDrawWaitsThenLearnsWhatChanged(t *testing.T) {
	h := New(4, 2*time.Second)
	h.setPaused(true)

	type result struct {
		ho proto.Handoff
		ok bool
	}
	got := make(chan result, 1)
	go func() {
		ho, ok := h.waitIfPaused()
		got <- result{ho, ok}
	}()

	time.Sleep(50 * time.Millisecond)
	if _, err := h.applyDraw([]string{"set", "1", "1", "5"}, true); err != nil {
		t.Fatalf("applyDraw: %v", err)
	}
	h.addNote("make the eyes bigger")

	select {
	case r := <-got:
		t.Fatalf("drawer was not held during the pause: %+v", r)
	default:
	}

	h.setPaused(false)
	r := <-got
	if !r.ok {
		t.Fatal("drawer gave up instead of resuming")
	}
	if r.ho.Edits != 1 {
		t.Fatalf("edits = %d, want 1", r.ho.Edits)
	}
	if r.ho.Box.String() != "1,1-1,1" {
		t.Fatalf("box = %s, want the one edited pixel 1,1-1,1", r.ho.Box)
	}
	if r.ho.Note != "make the eyes bigger" {
		t.Fatalf("note = %q, want the message left during the pause", r.ho.Note)
	}
}

func TestPausedDrawGivesUpAfterMaxWait(t *testing.T) {
	h := New(4, 100*time.Millisecond)
	h.setPaused(true)
	if _, ok := h.waitIfPaused(); ok {
		t.Fatal("expected the drawer to give up while still paused")
	}
}
