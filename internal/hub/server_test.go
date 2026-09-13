package hub

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

func dial(t *testing.T, addr, hello string) (*bufio.Reader, func(string)) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	r := bufio.NewReader(conn)
	send := func(line string) {
		t.Helper()
		if _, err := conn.Write([]byte(line + "\n")); err != nil {
			t.Fatalf("write %q: %v", line, err)
		}
	}
	if hello != "" {
		send(hello)
	}
	return r, send
}

func nextResult(t *testing.T, r *bufio.Reader) string {
	t.Helper()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "state ") {
			if _, _, err := proto.ReadState(r, line); err != nil {
				t.Fatalf("state: %v", err)
			}
			continue
		}
		if strings.HasPrefix(line, "log ") {
			if _, err := proto.ReadLog(r, line); err != nil {
				t.Fatalf("log: %v", err)
			}
			continue
		}
		if line == "paused" || line == "ok" || strings.HasPrefix(line, "ok ") || strings.HasPrefix(line, "err ") {
			return line
		}
	}
}

func TestDrawerIsHeldWhilePersonEditsThenLearnsWhatHappened(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, 5*time.Second).Serve(ln)
	addr := ln.Addr().String()

	person, personSend := dial(t, addr, "hello view")
	if got := nextResult(t, person); got != "ok" {
		t.Fatalf("hello reply = %q", got)
	}
	model, modelSend := dial(t, addr, "")

	personSend("pause")
	if got := nextResult(t, person); got != "ok" {
		t.Fatalf("pause reply = %q", got)
	}

	held := make(chan string, 1)
	go func() { held <- nextResult(t, model) }()
	modelSend("set 1 1 5")

	select {
	case got := <-held:
		t.Fatalf("brush stroke landed during the pause: %q", got)
	case <-time.After(150 * time.Millisecond):
	}

	personSend("set 0 0 7")
	nextResult(t, person)
	personSend("note make the eyes bigger")
	nextResult(t, person)
	personSend("resume")

	select {
	case got := <-held:
		for _, want := range []string{"ok changed=1", "edits=1 box=0,0-0,0", "note=make the eyes bigger"} {
			if !strings.Contains(got, want) {
				t.Fatalf("reply %q is missing %q", got, want)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("brush stroke never came back after resume")
	}
}

func TestPersonCanKeepEditingWhilePaused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, 5*time.Second).Serve(ln)

	person, send := dial(t, ln.Addr().String(), "hello view")
	nextResult(t, person)
	send("pause")
	nextResult(t, person)

	done := make(chan string, 1)
	go func() { done <- nextResult(t, person) }()
	send("set 3 3 4")

	select {
	case got := <-done:
		if !strings.Contains(got, "changed=1") {
			t.Fatalf("edit reply = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("a person's edit was held by their own pause")
	}
}

func TestDisconnectedClientIsDropped(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	h := New(8, time.Second)
	go h.Serve(ln)

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	waitFor(t, "client to register", func() bool { return h.clientCount() == 1 })
	conn.Close()
	waitFor(t, "client to be dropped", func() bool { return h.clientCount() == 0 })
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func nextLog(t *testing.T, r *bufio.Reader) []proto.LogEntry {
	t.Helper()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "log ") {
			entries, err := proto.ReadLog(r, line)
			if err != nil {
				t.Fatalf("log: %v", err)
			}
			return entries
		}
		if strings.HasPrefix(line, "state ") {
			if _, _, err := proto.ReadState(r, line); err != nil {
				t.Fatalf("state: %v", err)
			}
		}
	}
}

func TestSayAndNoteShareOneLog(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, time.Second).Serve(ln)
	addr := ln.Addr().String()

	person, personSend := dial(t, addr, "hello view")
	nextLog(t, person) // the empty log sent on connect
	_, modelSend := dial(t, addr, "")

	modelSend("say drawing the roof")
	got := nextLog(t, person)
	if len(got) != 1 || got[0].Who != "claude" || got[0].Text != "drawing the roof" {
		t.Fatalf("log after the model spoke = %+v", got)
	}

	personSend("note make it red")
	got = nextLog(t, person)
	if len(got) != 2 || got[1].Who != "you" || got[1].Text != "make it red" {
		t.Fatalf("log after the person replied = %+v", got)
	}
}

func TestSizeStartsOverOnABlankCanvas(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, time.Second).Serve(ln)

	person, send := dial(t, ln.Addr().String(), "hello view")
	nextResult(t, person)

	send("set 63 63 5")
	if got := nextResult(t, person); got != "ok changed=0" {
		t.Fatalf("pixel outside an 8x8 canvas landed: %q", got)
	}
	send("size 64")
	if got := nextResult(t, person); got != "ok" {
		t.Fatalf("size reply = %q", got)
	}
	send("set 63 63 5")
	if got := nextResult(t, person); got != "ok changed=1" {
		t.Fatalf("pixel at the far corner of a 64x64 canvas: %q", got)
	}
	for _, bad := range []string{"size 4", "size 999", "size big"} {
		send(bad)
		if got := nextResult(t, person); !strings.HasPrefix(got, "err ") {
			t.Fatalf("%q was accepted: %q", bad, got)
		}
	}
}

func TestListenIsHeldUntilThePersonSpeaks(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, 5*time.Second).Serve(ln)
	addr := ln.Addr().String()

	person, personSend := dial(t, addr, "hello view")
	nextResult(t, person)
	model, modelSend := dial(t, addr, "")

	held := make(chan string, 1)
	go func() { held <- nextResult(t, model) }()
	modelSend("listen")

	select {
	case got := <-held:
		t.Fatalf("listen answered before the person said anything: %q", got)
	case <-time.After(150 * time.Millisecond):
	}

	personSend("note draw a cat")
	select {
	case got := <-held:
		if got != "ok note=draw a cat" {
			t.Fatalf("listen reply = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("listen never came back after the person spoke")
	}
}

func TestListenGivesUpAfterTheWaitWindow(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, 100*time.Millisecond).Serve(ln)

	model, send := dial(t, ln.Addr().String(), "")
	send("listen")
	if got := nextResult(t, model); got != "ok" {
		t.Fatalf("listen reply after the window = %q", got)
	}
}

func TestNotesQueueUntilAStrokeReadsThem(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go New(8, time.Second).Serve(ln)
	addr := ln.Addr().String()

	person, personSend := dial(t, addr, "hello view")
	nextResult(t, person)
	model, modelSend := dial(t, addr, "")

	personSend("note make it red")
	nextResult(t, person)
	personSend("note and bigger")
	nextResult(t, person)

	modelSend("set 1 1 5")
	if got := nextResult(t, model); got != "ok changed=1 note=make it red / and bigger" {
		t.Fatalf("stroke reply = %q", got)
	}

	modelSend("set 2 2 5")
	if got := nextResult(t, model); got != "ok changed=1" {
		t.Fatalf("second stroke reply = %q", got)
	}
}
