package canvas

import "testing"

func has(pts [][2]int, x, y int) bool {
	for _, p := range pts {
		if p[0] == x && p[1] == y {
			return true
		}
	}
	return false
}

func TestLineTouchesEveryColumn(t *testing.T) {
	pts := LinePoints(2, 1, 9, 4)
	if !has(pts, 2, 1) || !has(pts, 9, 4) {
		t.Fatalf("line misses an endpoint: %v", pts)
	}
	if len(pts) != 8 {
		t.Fatalf("line from x=2 to x=9 has %d pixels, want 8", len(pts))
	}
	for x := 2; x <= 9; x++ {
		found := false
		for _, p := range pts {
			if p[0] == x {
				found = true
			}
		}
		if !found {
			t.Fatalf("no pixel in column %d: %v", x, pts)
		}
	}
	if back := LinePoints(9, 4, 2, 1); len(back) != len(pts) {
		t.Fatalf("reversed line has %d pixels, want %d", len(back), len(pts))
	}
}

func TestRectOutlineAndFill(t *testing.T) {
	if got := len(RectPoints(4, 4, 1, 2, false)); got != 10 {
		t.Fatalf("4x3 outline has %d pixels, want 10", got)
	}
	if got := len(RectPoints(1, 2, 4, 4, true)); got != 12 {
		t.Fatalf("4x3 filled has %d pixels, want 12", got)
	}
}

func TestEllipseIsRound(t *testing.T) {
	pts := EllipsePoints(10, 10, 3, 3, true)
	want := map[int]int{7: 3, 8: 5, 9: 7, 10: 7, 11: 7, 12: 5, 13: 3}
	rows := map[int]int{}
	for _, p := range pts {
		rows[p[1]]++
	}
	for y, n := range want {
		if rows[y] != n {
			t.Fatalf("row %d has %d pixels, want %d (rows %v)", y, rows[y], n, rows)
		}
	}
	outline := EllipsePoints(10, 10, 3, 3, false)
	if len(outline) != 16 {
		t.Fatalf("radius-3 outline has %d pixels, want 16", len(outline))
	}
	if has(outline, 10, 10) {
		t.Fatal("outline includes the center")
	}
}

func TestEllipseOnAHalfCenterIsItsOwnReflection(t *testing.T) {
	pts := EllipsePoints(15.5, 10, 9.5, 4, true)
	if !has(pts, 6, 10) || !has(pts, 25, 10) || has(pts, 5, 10) || has(pts, 26, 10) {
		t.Fatalf("ellipse at 15.5 with radius 9.5 does not span 6..25: %v", pts)
	}
	if mirrored := MirrorX(pts, 32); len(mirrored) != len(pts) {
		t.Fatalf("ellipse gained %d pixels when mirrored - it was not symmetric", len(mirrored)-len(pts))
	}
}

func TestMirrorKeepsEveryPixelOnce(t *testing.T) {
	got := MirrorX([][2]int{{1, 0}, {3, 0}, {4, 0}}, 8)
	if len(got) != 4 {
		t.Fatalf("mirror of 1,3,4 on an 8-wide canvas = %v, want 4 pixels", got)
	}
	if !has(got, 6, 0) || !has(got, 4, 0) || !has(got, 3, 0) || !has(got, 1, 0) {
		t.Fatalf("mirror = %v", got)
	}
	if got := MirrorX([][2]int{{3, 0}}, 7); len(got) != 1 {
		t.Fatalf("center pixel mirrored onto itself: %v", got)
	}
}
