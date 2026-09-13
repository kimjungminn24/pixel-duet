package canvas

import "math"

// LinePoints is Bresenham from (x0,y0) to (x1,y1), inclusive.
func LinePoints(x0, y0, x1, y1 int) [][2]int {
	dx, dy := max(x1-x0, x0-x1), -max(y1-y0, y0-y1)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	e := dx + dy
	var pts [][2]int
	for {
		pts = append(pts, [2]int{x0, y0})
		if x0 == x1 && y0 == y1 {
			return pts
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

func RectPoints(x0, y0, x1, y1 int, filled bool) [][2]int {
	x0, x1 = min(x0, x1), max(x0, x1)
	y0, y1 = min(y0, y1), max(y0, y1)
	var pts [][2]int
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if filled || x == x0 || x == x1 || y == y0 || y == y1 {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	return pts
}

// EllipsePoints spans 2*rx+1 pixels; center and radii may be halves so a
// shape can sit on the mirror axis of an even-width canvas (cx=15.5 on 32).
// The outline is every inside pixel with an outside 4-neighbor.
func EllipsePoints(cx, cy, rx, ry float64, filled bool) [][2]int {
	rx, ry = math.Abs(rx), math.Abs(ry)
	inside := func(x, y int) bool {
		if rx == 0 || ry == 0 {
			return math.Abs(float64(x)-cx) <= rx && math.Abs(float64(y)-cy) <= ry
		}
		fx := (float64(x) - cx) / (rx + 0.5)
		fy := (float64(y) - cy) / (ry + 0.5)
		return fx*fx+fy*fy <= 1
	}
	var pts [][2]int
	for y := int(math.Floor(cy - ry)); y <= int(math.Ceil(cy+ry)); y++ {
		for x := int(math.Floor(cx - rx)); x <= int(math.Ceil(cx+rx)); x++ {
			if !inside(x, y) {
				continue
			}
			edge := !inside(x-1, y) || !inside(x+1, y) || !inside(x, y-1) || !inside(x, y+1)
			if filled || edge {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	return pts
}

// MirrorX adds each point's reflection across the vertical center of a w-wide canvas, without duplicates.
func MirrorX(pts [][2]int, w int) [][2]int {
	out := make([][2]int, 0, 2*len(pts))
	seen := make(map[[2]int]bool, 2*len(pts))
	add := func(p [2]int) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, p := range pts {
		add(p)
		add([2]int{w - 1 - p[0], p[1]})
	}
	return out
}
