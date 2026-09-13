package hub

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

const DefaultAddr = "127.0.0.1:7777"

// client is one connection; all writes go through out so one goroutine owns the socket.
type client struct {
	out    chan string
	view   bool // edits as a person; never held by a pause
	mu     sync.Mutex
	closed bool
}

// send drops the message if the client is behind; the next state push is complete anyway.
func (c *client) send(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	select {
	case c.out <- msg:
	default:
	}
}

func (c *client) reply(line string) { c.send(line + "\n") }

func (c *client) fail(err error) { c.reply("err " + err.Error()) }

func (c *client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.out)
	}
}

func (c *client) speaker() string {
	if c.view {
		return "you"
	}
	return "claude"
}

func Run(addr string, size int, maxWait time.Duration) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("pixelduet server on %s, canvas %dx%d", addr, size, size)
	return New(size, maxWait).Serve(ln)
}

func (h *Hub) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go h.serveConn(conn)
	}
}

func (h *Hub) serveConn(conn net.Conn) {
	defer conn.Close()
	c := &client{out: make(chan string, 8)}
	h.add(c)

	done := make(chan struct{})
	go func() {
		defer close(done)
		w := bufio.NewWriter(conn)
		for msg := range c.out {
			if _, err := w.WriteString(msg); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}
		}
	}()

	h.greet(c)
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		h.handle(c, strings.TrimSpace(line))
	}

	// remove closes c.out, which is what ends the writer goroutine.
	h.remove(c)
	<-done
}

// handle replies before pushing, so a caller never sees the state its own command caused first.
func (h *Hub) handle(c *client, line string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}
	cmd, args := fields[0], fields[1:]
	switch cmd {
	case "hello":
		c.view = len(args) > 0 && args[0] == "view"
		c.reply("ok")
	case "get":
		h.greet(c)
	case "say":
		h.say(c.speaker(), textAfter(line, cmd))
		c.reply("ok")
	case "note":
		h.addNote(textAfter(line, cmd))
		c.reply("ok")
	case "listen":
		note, waited := h.awaitNote()
		c.reply("ok" + proto.OKTail(proto.Handoff{Waited: waited, Note: note}))
	case "pause", "resume":
		h.setPaused(cmd == "pause")
		c.reply("ok")
		h.pushState()
	case "size":
		n, err := parseSize(args)
		if err != nil {
			c.fail(err)
			return
		}
		h.resize(n)
		c.reply("ok")
		h.pushState()
	case "setcolor":
		i, col, err := parseSetColor(args)
		if err == nil {
			err = h.setColor(i, col)
		}
		if err != nil {
			c.fail(err)
			return
		}
		c.reply("ok")
		h.pushPalette()
	case "speed":
		d, err := parseDelay(args)
		if err != nil {
			c.fail(err)
			return
		}
		h.setDelay(d)
		c.reply("ok")
		h.pushSpeed()
	case "set", "stroke", "fill", "clear":
		h.handleDraw(c, fields)
	default:
		c.reply("err unknown command")
	}
}

func (h *Hub) handleDraw(c *client, fields []string) {
	ho := proto.Handoff{}
	if !c.view {
		var ok bool
		if ho, ok = h.waitIfPaused(); !ok {
			c.reply("paused")
			return
		}
		time.Sleep(h.getDelay())
	}
	n, err := h.applyDraw(fields, c.view)
	if err != nil {
		c.fail(err)
		return
	}
	c.reply(proto.OKLine(n, ho))
	h.pushState()
}

func textAfter(line, cmd string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, cmd))
}

func parseSize(args []string) (int, error) {
	usage := fmt.Errorf("usage: size <n>, %d to %d", minSize, maxSize)
	if len(args) != 1 {
		return 0, usage
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < minSize || n > maxSize {
		return 0, usage
	}
	return n, nil
}

func parseSetColor(args []string) (int, canvas.RGB, error) {
	usage := fmt.Errorf("usage: setcolor <index> <rrggbb>")
	if len(args) != 2 {
		return 0, canvas.RGB{}, usage
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, canvas.RGB{}, usage
	}
	col, err := canvas.ParseHex(args[1])
	if err != nil {
		return 0, canvas.RGB{}, usage
	}
	return i, col, nil
}

func parseDelay(args []string) (time.Duration, error) {
	usage := fmt.Errorf("usage: speed <milliseconds>")
	if len(args) != 1 {
		return 0, usage
	}
	ms, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, usage
	}
	return time.Duration(ms) * time.Millisecond, nil
}
