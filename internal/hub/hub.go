package hub

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

// Hub owns the canvas and everything shared between clients, under one lock.
type Hub struct {
	mu       sync.Mutex
	canvas   *canvas.Canvas
	paused   bool
	resumeCh chan struct{} // closed on resume
	maxWait  time.Duration

	edits     int // pixels a view client changed during the current pause
	editBox   canvas.Box
	lastEdits int // frozen at resume for the drawer that was held
	lastBox   canvas.Box

	notes  []string
	noteCh chan struct{} // closed when a note arrives

	log []proto.LogEntry

	palette []canvas.RGB
	delay   time.Duration // gap between drawer strokes

	clients map[*client]bool
}

const (
	minSize = 8
	maxSize = 128
)

const maxDelay = 3 * time.Second

const logMax = 60

func New(size int, maxWait time.Duration) *Hub {
	return &Hub{
		canvas:   canvas.NewCanvas(size, size),
		resumeCh: make(chan struct{}),
		noteCh:   make(chan struct{}),
		editBox:  canvas.EmptyBox(),
		lastBox:  canvas.EmptyBox(),
		maxWait:  maxWait,
		clients:  make(map[*client]bool),
		palette:  append([]canvas.RGB(nil), canvas.Palette...),
	}
}

func (h *Hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.close()
}

func (h *Hub) clientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// broadcastLocked runs under h.mu so pushes arrive in state order; send never blocks.
func (h *Hub) broadcastLocked(msg string) {
	for c := range h.clients {
		c.send(msg)
	}
}

func (h *Hub) pushState() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcastLocked(proto.EncodeState(h.canvas, h.paused))
}

func (h *Hub) pushLog() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcastLocked(proto.EncodeLog(h.log))
}

func (h *Hub) pushPalette() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcastLocked(proto.EncodePalette(h.palette))
	h.broadcastLocked(proto.EncodeState(h.canvas, h.paused))
}

func (h *Hub) pushSpeed() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.broadcastLocked(proto.EncodeSpeed(h.delay))
}

// greet sends a client the full state; palette first so the state can be painted.
func (h *Hub) greet(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c.send(proto.EncodePalette(h.palette))
	c.send(proto.EncodeState(h.canvas, h.paused))
	c.send(proto.EncodeLog(h.log))
	c.send(proto.EncodeSpeed(h.delay))
}

func (h *Hub) setColor(i int, c canvas.RGB) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i < 1 || i >= len(h.palette) {
		return fmt.Errorf("color index %d is out of range (1-%d)", i, len(h.palette)-1)
	}
	h.palette[i] = c
	return nil
}

// resize starts over on a blank canvas; palette, log and pause carry on.
func (h *Hub) resize(n int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.canvas = canvas.NewCanvas(n, n)
}

func (h *Hub) setDelay(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.delay = max(0, min(d, maxDelay))
}

func (h *Hub) getDelay() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.delay
}

func (h *Hub) setPaused(p bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if p == h.paused {
		return
	}
	h.paused = p
	if p {
		h.edits, h.editBox = 0, canvas.EmptyBox()
		h.resumeCh = make(chan struct{})
		return
	}
	h.lastEdits, h.lastBox = h.edits, h.editBox
	close(h.resumeCh)
}

// waitIfPaused holds a drawer until resume or maxWait, and collects queued notes either way.
func (h *Hub) waitIfPaused() (proto.Handoff, bool) {
	h.mu.Lock()
	if !h.paused {
		ho := proto.Handoff{Note: h.takeNotesLocked()}
		h.mu.Unlock()
		return ho, true
	}
	ch := h.resumeCh
	h.mu.Unlock()

	start := time.Now()
	select {
	case <-ch:
	case <-time.After(h.maxWait):
		return proto.Handoff{Waited: time.Since(start)}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return proto.Handoff{
		Waited: time.Since(start),
		Edits:  h.lastEdits,
		Box:    h.lastBox,
		Note:   h.takeNotesLocked(),
	}, true
}

func (h *Hub) addNote(s string) {
	if strings.TrimSpace(s) == "" {
		return
	}
	h.mu.Lock()
	h.notes = append(h.notes, s)
	close(h.noteCh)
	h.noteCh = make(chan struct{})
	h.mu.Unlock()
	h.say("you", s)
}

func (h *Hub) takeNotesLocked() string {
	s := strings.Join(h.notes, proto.NoteSep)
	h.notes = nil
	return s
}

// awaitNote blocks until a note arrives or maxWait passes.
func (h *Hub) awaitNote() (note string, waited time.Duration) {
	h.mu.Lock()
	if len(h.notes) > 0 {
		note = h.takeNotesLocked()
		h.mu.Unlock()
		return note, 0
	}
	ch := h.noteCh
	h.mu.Unlock()

	start := time.Now()
	select {
	case <-ch:
	case <-time.After(h.maxWait):
		return "", time.Since(start)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.takeNotesLocked(), time.Since(start)
}

func (h *Hub) say(who, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.log = append(h.log, proto.LogEntry{Who: who, Text: text})
	if len(h.log) > logMax {
		h.log = h.log[len(h.log)-logMax:]
	}
	h.broadcastLocked(proto.EncodeLog(h.log))
}

// applyDraw runs one drawing command and returns the number of changed pixels.
func (h *Hub) applyDraw(fields []string, fromView bool) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	n, box := 0, canvas.EmptyBox()
	switch cmd, args := fields[0], fields[1:]; cmd {
	case "set":
		x, y, col, err := parseXYColor(args, "set <x> <y> <color>")
		if err != nil {
			return 0, err
		}
		if h.canvas.Set(x, y, col) {
			n = 1
			box.Add(x, y)
		}
	case "fill":
		x, y, col, err := parseXYColor(args, "fill <x> <y> <color>")
		if err != nil {
			return 0, err
		}
		n, box = h.canvas.Fill(x, y, col)
	case "stroke":
		if len(args) < 2 {
			return 0, fmt.Errorf("usage: stroke <color> <x>,<y> ...")
		}
		col, err := strconv.Atoi(args[0])
		if err != nil {
			return 0, fmt.Errorf("bad color %q", args[0])
		}
		pts, err := proto.ParsePoints(args[1:])
		if err != nil {
			return 0, err
		}
		for _, p := range pts {
			if h.canvas.Set(p[0], p[1], uint8(col)) {
				n++
				box.Add(p[0], p[1])
			}
		}
	case "clear":
		h.canvas.Clear()
	default:
		return 0, fmt.Errorf("unknown command %q", cmd)
	}

	if fromView && h.paused {
		h.edits += n
		h.editBox.Union(box)
	}
	return n, nil
}

func parseXYColor(args []string, usage string) (x, y int, color uint8, err error) {
	if len(args) != 3 {
		return 0, 0, 0, fmt.Errorf("usage: %s", usage)
	}
	var v [3]int
	for i, a := range args {
		if v[i], err = strconv.Atoi(a); err != nil {
			return 0, 0, 0, fmt.Errorf("bad number %q", a)
		}
	}
	return v[0], v[1], uint8(v[2]), nil
}
