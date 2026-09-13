package canvas

import (
	"fmt"
	"math"
	"strings"
)

// Canvas is the pixel grid: one palette index per cell, row-major.
type Canvas struct {
	W, H  int
	Cells []uint8
}

func NewCanvas(w, h int) *Canvas {
	return &Canvas{W: w, H: h, Cells: make([]uint8, w*h)}
}

func (c *Canvas) clone() *Canvas {
	return &Canvas{W: c.W, H: c.H, Cells: append([]uint8(nil), c.Cells...)}
}

func (c *Canvas) in(x, y int) bool { return x >= 0 && x < c.W && y >= 0 && y < c.H }

// At returns 0 for anything off the canvas.
func (c *Canvas) At(x, y int) uint8 {
	if !c.in(x, y) {
		return 0
	}
	return c.Cells[y*c.W+x]
}

// Set reports whether the pixel changed.
func (c *Canvas) Set(x, y int, color uint8) bool {
	if !c.in(x, y) || !ValidColor(int(color)) {
		return false
	}
	i := y*c.W + x
	if c.Cells[i] == color {
		return false
	}
	c.Cells[i] = color
	return true
}

func (c *Canvas) Clear() {
	clear(c.Cells)
}

// Fill flood fills from (x,y) and returns the count and box of changed pixels.
func (c *Canvas) Fill(x, y int, to uint8) (int, Box) {
	box := EmptyBox()
	from := c.At(x, y)
	if !c.in(x, y) || from == to || !ValidColor(int(to)) {
		return 0, box
	}
	n := 0
	stack := [][2]int{{x, y}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !c.in(p[0], p[1]) || c.Cells[p[1]*c.W+p[0]] != from {
			continue
		}
		c.Cells[p[1]*c.W+p[0]] = to
		box.Add(p[0], p[1])
		n++
		stack = append(stack,
			[2]int{p[0] + 1, p[1]}, [2]int{p[0] - 1, p[1]},
			[2]int{p[0], p[1] + 1}, [2]int{p[0], p[1] - 1})
	}
	return n, box
}

// Rect is a window onto the canvas. Zero W or H means the whole canvas.
type Rect struct{ X, Y, W, H int }

func (r Rect) Clip(c *Canvas) Rect {
	if r.W <= 0 || r.H <= 0 {
		return Rect{0, 0, c.W, c.H}
	}
	return r
}

func (c *Canvas) String() string {
	return c.Digits(Rect{})
}

// Digits renders a window as one digit per pixel, one line per row.
func (c *Canvas) Digits(r Rect) string {
	r = r.Clip(c)
	var b strings.Builder
	b.Grow(r.H * (r.W + 1))
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			b.WriteByte(digits[c.At(x, y)])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Runs renders a window as run-length rows: "y3-9: 0*8 4*16 0*8".
func (c *Canvas) Runs(r Rect) string {
	r = r.Clip(c)
	rows := make([]string, r.H)
	for i := range rows {
		rows[i] = c.runRow(r.Y+i, r.X, r.X+r.W)
	}
	var b strings.Builder
	for i := 0; i < r.H; {
		n := 1
		for i+n < r.H && rows[i+n] == rows[i] {
			n++
		}
		y := r.Y + i
		if n == 1 {
			fmt.Fprintf(&b, "y%d: %s\n", y, rows[i])
		} else {
			fmt.Fprintf(&b, "y%d-%d: %s\n", y, y+n-1, rows[i])
		}
		i += n
	}
	return b.String()
}

func (c *Canvas) runRow(y, x0, x1 int) string {
	var b strings.Builder
	for x := x0; x < x1; {
		d := c.At(x, y)
		n := 1
		for x+n < x1 && c.At(x+n, y) == d {
			n++
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteByte(digits[d])
		if n > 1 {
			fmt.Fprintf(&b, "*%d", n)
		}
		x += n
	}
	return b.String()
}

// ParseGrid is the inverse of String.
func ParseGrid(s string) (*Canvas, error) {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, fmt.Errorf("empty grid")
	}
	c := NewCanvas(len(lines[0]), len(lines))
	for y, line := range lines {
		if len(line) != c.W {
			return nil, fmt.Errorf("row %d has %d cells, want %d", y, len(line), c.W)
		}
		for x := 0; x < c.W; x++ {
			i := strings.IndexByte(digits, line[x])
			if i < 0 {
				return nil, fmt.Errorf("row %d col %d: %q is not a palette digit", y, x, line[x])
			}
			c.Cells[y*c.W+x] = uint8(i)
		}
	}
	return c, nil
}

func Crop(c *Canvas, r Rect) *Canvas {
	r = r.Clip(c)
	out := NewCanvas(r.W, r.H)
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			out.Cells[y*r.W+x] = c.At(r.X+x, r.Y+y)
		}
	}
	return out
}

// Box is an inclusive pixel rectangle that starts empty and grows with Add.
type Box struct{ X0, Y0, X1, Y1 int }

func EmptyBox() Box { return Box{X0: math.MaxInt, Y0: math.MaxInt, X1: -1, Y1: -1} }

func (b Box) Empty() bool { return b.X1 < b.X0 }

func (b *Box) Add(x, y int) {
	b.X0, b.Y0 = min(b.X0, x), min(b.Y0, y)
	b.X1, b.Y1 = max(b.X1, x), max(b.Y1, y)
}

func (b *Box) Union(o Box) {
	if !o.Empty() {
		b.Add(o.X0, o.Y0)
		b.Add(o.X1, o.Y1)
	}
}

func (b Box) String() string { return fmt.Sprintf("%d,%d-%d,%d", b.X0, b.Y0, b.X1, b.Y1) }
