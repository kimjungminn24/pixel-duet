package bridge

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

// hubConn is the bridge's connection to the hub, opened lazily and dropped on error.
type hubConn struct {
	mu   sync.Mutex
	addr string
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

func (d *hubConn) ensure() error {
	if d.conn != nil {
		return nil
	}
	conn, err := net.Dial("tcp", d.addr)
	if err != nil {
		return fmt.Errorf("cannot reach the pixelduet server at %s (start it with: pixelduet serve): %w", d.addr, err)
	}
	d.conn, d.r, d.w = conn, bufio.NewReader(conn), bufio.NewWriter(conn)
	return nil
}

func (d *hubConn) drop() {
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}
}

func (d *hubConn) send(line string) error {
	if _, err := d.w.WriteString(line + "\n"); err != nil {
		return err
	}
	return d.w.Flush()
}

// cmd sends one line and waits for its reply, which may take a while during a pause.
func (d *hubConn) cmd(line string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.ensure(); err != nil {
		return "", err
	}
	if err := d.send(line); err != nil {
		d.drop()
		return "", err
	}
	res, err := proto.AwaitResult(d.r)
	if err != nil {
		d.drop()
		return "", fmt.Errorf("lost the server mid-command: %w", err)
	}
	return res, nil
}

// canvas returns the grid with the live palette that precedes it.
func (d *hubConn) canvas() (c *canvas.Canvas, paused bool, pal []canvas.RGB, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.ensure(); err != nil {
		return nil, false, nil, err
	}
	if err := d.send("get"); err != nil {
		d.drop()
		return nil, false, nil, err
	}
	pal = canvas.Palette
	for {
		m, err := proto.ReadMessage(d.r)
		if err != nil {
			d.drop()
			return nil, false, nil, fmt.Errorf("lost the server mid-command: %w", err)
		}
		switch m.Kind {
		case proto.MsgPalette:
			if len(m.Palette) == len(canvas.Palette) {
				pal = m.Palette
			}
		case proto.MsgState:
			return m.Canvas, m.Paused, pal, nil
		}
	}
}
