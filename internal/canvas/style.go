package canvas

import "slices"

// neighbors is 4-connected on purpose: diagonals would square off outline corners.
var neighbors = [4][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}

type Stroke struct {
	Color uint8
	Pts   [][2]int
}

// strokeSet groups pixels by color in first-seen order.
type strokeSet struct {
	order []uint8
	by    map[uint8][][2]int
}

func (s *strokeSet) add(color uint8, x, y int) {
	if s.by == nil {
		s.by = make(map[uint8][][2]int)
	}
	if _, seen := s.by[color]; !seen {
		s.order = append(s.order, color)
	}
	s.by[color] = append(s.by[color], [2]int{x, y})
}

func (s *strokeSet) strokes() []Stroke {
	out := make([]Stroke, 0, len(s.order))
	for _, color := range s.order {
		out = append(out, Stroke{color, s.by[color]})
	}
	return out
}

func oneStroke(color uint8, pts [][2]int) []Stroke {
	if len(pts) == 0 {
		return nil
	}
	return []Stroke{{color, pts}}
}

// OutlineStrokes rings painted pixels with color, or with each neighbor's Shade when color is 0.
func OutlineStrokes(c *Canvas, r Rect, color uint8) []Stroke {
	r = r.Clip(c)
	var set strokeSet
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			if !c.in(x, y) || c.At(x, y) != 0 {
				continue
			}
			for _, d := range neighbors {
				n := c.At(x+d[0], y+d[1])
				if n == 0 {
					continue
				}
				if color != 0 {
					set.add(color, x, y)
				} else {
					set.add(Shade[n], x, y)
				}
				break
			}
		}
	}
	return set.strokes()
}

// UnoutlineStrokes erases pixels that touch transparency and are the ring of what they sit against.
func UnoutlineStrokes(c *Canvas, r Rect, color uint8) []Stroke {
	r = r.Clip(c)
	var pts [][2]int
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			col := c.At(x, y)
			if !c.in(x, y) || col == 0 || (color != 0 && col != color) {
				continue
			}
			// With a named color the ring is identified by color alone.
			bare, ringing := false, color != 0
			for _, d := range neighbors {
				switch n := c.At(x+d[0], y+d[1]); {
				case n == 0:
					bare = true
				case n != col && Shade[n] == col:
					ringing = true
				}
			}
			if bare && ringing {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	return oneStroke(0, pts)
}

// LightDir returns the shadow direction for a named light.
func LightDir(light string) (dx, dy int, ok bool) {
	switch light {
	case "", "top-left":
		return 1, 1, true
	case "top-right":
		return -1, 1, true
	case "bottom-left":
		return 1, -1, true
	case "bottom-right":
		return -1, -1, true
	}
	return 0, 0, false
}

func ShadeStrokes(c *Canvas, r Rect, dx, dy, width int) []Stroke {
	return toneStrokes(c, r, dx, dy, width, Shade[:])
}

func HighlightStrokes(c *Canvas, r Rect, dx, dy, width int) []Stroke {
	return toneStrokes(c, r, -dx, -dy, width, Light[:])
}

// shadowStrokes offsets the silhouette into empty pixels.
func shadowStrokes(c *Canvas, r Rect, dx, dy, distance int, color uint8) []Stroke {
	r = r.Clip(c)
	var pts [][2]int
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			tx, ty := x+dx*distance, y+dy*distance
			if c.At(x, y) != 0 && c.in(tx, ty) && c.At(tx, ty) == 0 {
				pts = append(pts, [2]int{tx, ty})
			}
		}
	}
	return oneStroke(color, pts)
}

// toneStrokes steps the edge of each shape facing (dx,dy) along ramp, width times.
// Existing bands are skipped, so repeated passes widen inward instead of stacking.
func toneStrokes(c *Canvas, r Rect, dx, dy, width int, ramp []uint8) []Stroke {
	r = r.Clip(c)
	work := c.clone()
	outline := terminalOutlineMask(work, ramp)
	isOutline := func(x, y int) bool { return work.in(x, y) && outline[y*work.W+x] }

	type toned struct {
		x, y  int
		Color uint8
	}
	var set strokeSet
	for pass := 0; pass < max(width, 1); pass++ {
		band := bandMask(work, ramp)
		var found []toned
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				col := work.At(x, y)
				if !work.in(x, y) || col == 0 || ramp[col] == col || band[y*work.W+x] {
					continue
				}
				next := ramp[col]
				ahead, below := work.At(x+dx, y), work.At(x, y+dy)
				onEdge := ahead == 0 || ahead == next || below == 0 || below == next ||
					isOutline(x+dx, y) || isOutline(x, y+dy)
				if onEdge {
					found = append(found, toned{x, y, next})
				}
			}
		}
		for _, f := range found {
			work.Set(f.x, f.y, f.Color)
			set.add(f.Color, f.x, f.y)
		}
	}
	return set.strokes()
}

// terminalOutlineMask marks ramp-terminal patches that touch transparency (outer rings, not pupils).
func terminalOutlineMask(c *Canvas, ramp []uint8) []bool {
	return patchMask(c, func(col uint8, around []uint8) bool {
		return ramp[col] == col && slices.Contains(around, 0)
	})
}

// bandMask marks patches that are the next tone of a color they border.
func bandMask(c *Canvas, ramp []uint8) []bool {
	return patchMask(c, func(col uint8, around []uint8) bool {
		for _, n := range around {
			if n != 0 && n != col && ramp[n] == col {
				return true
			}
		}
		return false
	})
}

// patchMask marks each connected same-color patch for which keep(color, neighboring colors) is true.
// around contains 0 for empty pixels and for the canvas edge.
func patchMask(c *Canvas, keep func(col uint8, around []uint8) bool) []bool {
	mask := make([]bool, len(c.Cells))
	seen := make([]bool, len(c.Cells))
	var patch, stack []int
	var around []uint8
	for start, col := range c.Cells {
		if seen[start] || col == 0 {
			continue
		}
		var met [PaletteSize]bool
		patch, stack, around = patch[:0], append(stack[:0], start), around[:0]
		seen[start] = true
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			patch = append(patch, i)
			x, y := i%c.W, i/c.W
			for _, d := range neighbors {
				nx, ny := x+d[0], y+d[1]
				n := c.At(nx, ny)
				if n == col && c.in(nx, ny) {
					if j := ny*c.W + nx; !seen[j] {
						seen[j] = true
						stack = append(stack, j)
					}
					continue
				}
				if !met[n] {
					met[n] = true
					around = append(around, n)
				}
			}
		}
		if keep(col, around) {
			for _, i := range patch {
				mask[i] = true
			}
		}
	}
	return mask
}

// DitherStrokes replaces every other pixel of from with to on a checkerboard.
func DitherStrokes(c *Canvas, r Rect, from, to uint8) []Stroke {
	r = r.Clip(c)
	var pts [][2]int
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			if c.in(x, y) && c.At(x, y) == from && (x+y)%2 == 0 {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	return oneStroke(to, pts)
}

// Asymmetry counts pixels that differ from their reflection across the vertical center.
func Asymmetry(c *Canvas) (int, Box) {
	n, box := 0, EmptyBox()
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W/2; x++ {
			if c.At(x, y) != c.At(c.W-1-x, y) {
				n++
				box.Add(x, y)
				box.Add(c.W-1-x, y)
			}
		}
	}
	return n, box
}
