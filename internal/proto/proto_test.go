package proto

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
)

func TestReadMessageConsumesWholeBodies(t *testing.T) {
	c := canvas.NewCanvas(3, 2)
	c.Set(1, 1, 9)
	wire := EncodePalette(canvas.Palette) +
		EncodeState(c, true) +
		EncodeLog([]LogEntry{{"claude", "roof next"}, {"you", "make it red"}}) +
		EncodeSpeed(250*time.Millisecond) +
		"ok changed=1\n"
	r := bufio.NewReader(strings.NewReader(wire))

	m, err := ReadMessage(r)
	if err != nil || m.Kind != MsgPalette || len(m.Palette) != len(canvas.Palette) || m.Palette[9] != canvas.Palette[9] {
		t.Fatalf("palette = %+v, %v", m, err)
	}
	m, err = ReadMessage(r)
	if err != nil || m.Kind != MsgState || !m.Paused || m.Canvas.String() != c.String() {
		t.Fatalf("state = %+v, %v", m, err)
	}
	m, err = ReadMessage(r)
	if err != nil || m.Kind != MsgLog || len(m.Entries) != 2 || m.Entries[1] != (LogEntry{"you", "make it red"}) {
		t.Fatalf("log = %+v, %v", m, err)
	}
	m, err = ReadMessage(r)
	if err != nil || m.Kind != MsgSpeed || m.Delay != 250*time.Millisecond {
		t.Fatalf("speed = %+v, %v", m, err)
	}
	m, err = ReadMessage(r)
	if err != nil || m.Kind != MsgReply || m.Line != "ok changed=1" {
		t.Fatalf("Reply = %+v, %v", m, err)
	}
}

func TestParseReplyKeepsFieldsApartFromTheNote(t *testing.T) {
	box := canvas.EmptyBox()
	box.Add(9, 8)
	box.Add(13, 10)
	line := OKLine(1, Handoff{Waited: 12 * time.Second, Edits: 5, Box: box, Note: "no edits=please / bigger"})
	if line != "ok changed=1 waited=12s edits=5 box=9,8-13,10 note=no edits=please / bigger" {
		t.Fatalf("OKLine = %q", line)
	}
	p := ParseReply(line)
	if p.Changed != 1 || p.Waited != 12*time.Second || p.Edits != 5 || p.Box != "9,8-13,10" ||
		!p.HasNote || p.Note != "no edits=please / bigger" {
		t.Fatalf("ParseReply = %+v", p)
	}
	if p := ParseReply("ok changed=3 note=looks edits=fine"); p.Edits != 0 || p.Note != "looks edits=fine" {
		t.Fatalf("a note that mentions edits was taken for a field: %+v", p)
	}
	if p := ParseReply("ok"); p.HasNote || p.Changed != 0 {
		t.Fatalf("bare ok = %+v", p)
	}
}

func TestStrokeRoundTrip(t *testing.T) {
	pts := [][2]int{{1, 2}, {3, 4}}
	line := EncodeStroke(7, pts)
	if line != "stroke 7 1,2 3,4" {
		t.Fatalf("EncodeStroke = %q", line)
	}
	got, err := ParsePoints(strings.Fields(line)[2:])
	if err != nil || len(got) != 2 || got[1] != [2]int{3, 4} {
		t.Fatalf("ParsePoints = %v, %v", got, err)
	}
	if _, err := ParsePoints(nil); err == nil {
		t.Fatal("a stroke with no points was accepted")
	}
}
