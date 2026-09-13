package proto

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
)

// Wire format, one command or Reply per line:
//
//	client -> server   hello [view]    view clients edit as a person; never held
//	                   set <x> <y> <c>
//	                   stroke <c> <x>,<y> <x>,<y> ...
//	                   fill <x> <y> <c>
//	                   clear | get | pause | resume
//	                   size <n>
//	                   setcolor <i> <rrggbb>
//	                   speed <ms>
//	                   say <text>
//	                   note <text>     queued until a drawer reads it
//	                   listen          blocks until the next note
//
//	server -> client   state <w> <h> <paused>
//	                   <one digit per pixel, h lines>
//	                   .
//	                   log <n>
//	                   <who> <text>   (n lines)
//	                   .
//	                   palette - <rrggbb> ...
//	                   speed <ms>
//	                   ok [changed=<n>] [waited=<d>] [edits=<n> box=<x0>,<y0>-<x1>,<y1>] [note=<text>]
//	                   paused | err <Message>
//
// state, log, palette and speed are pushed on connect and on change.
// edits/box describe what a view client changed while a drawer was held.
// note carries every note queued since the last Reply, joined with " / ".

const endMarker = "."

const NoteSep = " / "

type Kind int

const (
	MsgReply Kind = iota
	MsgState
	MsgLog
	MsgPalette
	MsgSpeed
	MsgOther
)

// Message is one server Message with its body consumed.
type Message struct {
	Kind    Kind
	Line    string
	Canvas  *canvas.Canvas
	Paused  bool
	Entries []LogEntry
	Palette []canvas.RGB
	Delay   time.Duration
}

func ReadMessage(r *bufio.Reader) (Message, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return Message{}, err
	}
	m := Message{Line: strings.TrimSpace(line)}
	switch {
	case IsReply(m.Line):
		m.Kind = MsgReply
	case strings.HasPrefix(m.Line, "state "):
		m.Kind = MsgState
		m.Canvas, m.Paused, err = ReadState(r, m.Line)
	case strings.HasPrefix(m.Line, "log "):
		m.Kind = MsgLog
		m.Entries, err = ReadLog(r, m.Line)
	case strings.HasPrefix(m.Line, "palette "):
		m.Kind = MsgPalette
		m.Palette, err = ParsePalette(m.Line)
	case strings.HasPrefix(m.Line, "speed "):
		m.Kind = MsgSpeed
		m.Delay, err = ParseSpeed(m.Line)
	default:
		m.Kind = MsgOther
	}
	return m, err
}

func IsReply(line string) bool {
	return line == "ok" || line == "paused" ||
		strings.HasPrefix(line, "ok ") || strings.HasPrefix(line, "err ")
}

// AwaitResult skips pushes until a Reply line.
func AwaitResult(r *bufio.Reader) (string, error) {
	for {
		m, err := ReadMessage(r)
		if err != nil {
			return "", err
		}
		if m.Kind == MsgReply {
			return m.Line, nil
		}
	}
}

func AwaitState(r *bufio.Reader) (*canvas.Canvas, bool, error) {
	for {
		m, err := ReadMessage(r)
		if err != nil {
			return nil, false, err
		}
		if m.Kind == MsgState {
			return m.Canvas, m.Paused, nil
		}
	}
}

func readBody(r *bufio.Reader, n int, what string) ([]string, error) {
	lines := make([]string, 0, n)
	for i := 0; i < n; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		lines = append(lines, strings.TrimRight(line, "\r\n"))
	}
	end, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(end) != endMarker {
		return nil, fmt.Errorf("%s did not end with %q", what, endMarker)
	}
	return lines, nil
}

func EncodeState(c *canvas.Canvas, paused bool) string {
	p := 0
	if paused {
		p = 1
	}
	return fmt.Sprintf("state %d %d %d\n%s%s\n", c.W, c.H, p, c, endMarker)
}

func ReadState(r *bufio.Reader, header string) (*canvas.Canvas, bool, error) {
	var w, h, p int
	if _, err := fmt.Sscanf(header, "state %d %d %d", &w, &h, &p); err != nil {
		return nil, false, fmt.Errorf("bad state header %q: %w", header, err)
	}
	rows, err := readBody(r, h, "state")
	if err != nil {
		return nil, false, err
	}
	c, err := canvas.ParseGrid(strings.Join(rows, "\n") + "\n")
	if err != nil {
		return nil, false, err
	}
	return c, p == 1, nil
}

type LogEntry struct {
	Who  string // "claude" or "you"
	Text string
}

func EncodeLog(entries []LogEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "log %d\n", len(entries))
	for _, e := range entries {
		fmt.Fprintf(&b, "%s %s\n", e.Who, e.Text)
	}
	b.WriteString(endMarker + "\n")
	return b.String()
}

func ReadLog(r *bufio.Reader, header string) ([]LogEntry, error) {
	var n int
	if _, err := fmt.Sscanf(header, "log %d", &n); err != nil {
		return nil, fmt.Errorf("bad log header %q: %w", header, err)
	}
	lines, err := readBody(r, n, "log")
	if err != nil {
		return nil, err
	}
	entries := make([]LogEntry, 0, n)
	for _, line := range lines {
		who, text, _ := strings.Cut(line, " ")
		entries = append(entries, LogEntry{Who: who, Text: text})
	}
	return entries, nil
}

func EncodePalette(pal []canvas.RGB) string {
	parts := make([]string, len(pal))
	parts[0] = "-"
	for i := 1; i < len(pal); i++ {
		parts[i] = pal[i].Hex()
	}
	return "palette " + strings.Join(parts, " ") + "\n"
}

func ParsePalette(line string) ([]canvas.RGB, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "palette" {
		return nil, fmt.Errorf("bad palette line %q", line)
	}
	pal := make([]canvas.RGB, len(fields)-1)
	for i, f := range fields[2:] {
		c, err := canvas.ParseHex(f)
		if err != nil {
			return nil, err
		}
		pal[i+1] = c
	}
	return pal, nil
}

func EncodeSpeed(d time.Duration) string {
	return fmt.Sprintf("speed %d\n", d.Milliseconds())
}

func ParseSpeed(line string) (time.Duration, error) {
	var ms int
	if _, err := fmt.Sscanf(line, "speed %d", &ms); err != nil {
		return 0, fmt.Errorf("bad speed line %q: %w", line, err)
	}
	return time.Duration(ms) * time.Millisecond, nil
}

func EncodeStroke(color uint8, pts [][2]int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "stroke %d", color)
	for _, p := range pts {
		fmt.Fprintf(&b, " %d,%d", p[0], p[1])
	}
	return b.String()
}

func ParsePoints(fields []string) ([][2]int, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("stroke needs at least one point")
	}
	pts := make([][2]int, 0, len(fields))
	for _, f := range fields {
		var x, y int
		if _, err := fmt.Sscanf(f, "%d,%d", &x, &y); err != nil {
			return nil, fmt.Errorf("bad point %q", f)
		}
		pts = append(pts, [2]int{x, y})
	}
	return pts, nil
}

// Handoff is what a held drawer is told on release.
type Handoff struct {
	Waited time.Duration
	Edits  int
	Box    canvas.Box
	Note   string
}

func OKLine(changed int, ho Handoff) string {
	return fmt.Sprintf("ok changed=%d%s", changed, OKTail(ho))
}

// OKTail puts note= last because the text may contain anything.
func OKTail(ho Handoff) string {
	var b strings.Builder
	if ho.Waited > time.Second {
		fmt.Fprintf(&b, " waited=%s", ho.Waited.Round(time.Second))
	}
	if ho.Edits > 0 {
		fmt.Fprintf(&b, " edits=%d", ho.Edits)
		if !ho.Box.Empty() {
			fmt.Fprintf(&b, " box=%s", ho.Box)
		}
	}
	if ho.Note != "" {
		fmt.Fprintf(&b, " note=%s", ho.Note)
	}
	return b.String()
}

type Reply struct {
	Changed int
	Waited  time.Duration
	Edits   int
	Box     string
	Note    string
	HasNote bool
}

func ParseReply(line string) Reply {
	var p Reply
	head, note, ok := strings.Cut(line, " note=")
	p.Note, p.HasNote = note, ok
	for _, f := range strings.Fields(head) {
		k, v, _ := strings.Cut(f, "=")
		switch k {
		case "changed":
			p.Changed, _ = strconv.Atoi(v)
		case "waited":
			p.Waited, _ = time.ParseDuration(v)
		case "edits":
			p.Edits, _ = strconv.Atoi(v)
		case "box":
			p.Box = v
		}
	}
	return p
}
