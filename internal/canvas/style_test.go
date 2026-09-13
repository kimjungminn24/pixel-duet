package canvas

import "testing"

func square(size, x0, y0, x1, y1 int, color uint8) *Canvas {
	c := NewCanvas(size, size)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c.Set(x, y, color)
		}
	}
	return c
}

func count(strokes []Stroke) int {
	n := 0
	for _, s := range strokes {
		n += len(s.Pts)
	}
	return n
}

func TestOutlineRingsTheFigureInItsShade(t *testing.T) {
	c := square(7, 2, 2, 4, 4, 11) // green
	got := OutlineStrokes(c, Rect{}, 0)
	if len(got) != 1 || got[0].Color != 13 || len(got[0].Pts) != 12 {
		t.Fatalf("outline = %+v, want 12 dark-green pixels", got)
	}
	for _, p := range got[0].Pts {
		if (p[0] == 1 || p[0] == 5) && (p[1] == 1 || p[1] == 5) {
			t.Fatalf("outline includes corner %v", p)
		}
	}
	if got := OutlineStrokes(c, Rect{}, 1); got[0].Color != 1 {
		t.Fatalf("outline color = %d, want black as asked", got[0].Color)
	}
	for _, p := range got[0].Pts {
		c.Set(p[0], p[1], 13)
	}
	if got := OutlineStrokes(c, Rect{}, 0); got[0].Color != Shade[13] {
		t.Fatalf("second outline is %d, want the shade of the outline", got[0].Color)
	}
}

func TestUnoutlineTakesTheRingOffAndLeavesTheFigure(t *testing.T) {
	c := square(9, 3, 3, 5, 5, 11) // green
	c.Set(4, 4, 1)                 // a black pupil in the middle
	before := c.String()

	for _, s := range OutlineStrokes(c, Rect{}, 1) { // black outline
		for _, p := range s.Pts {
			c.Set(p[0], p[1], s.Color)
		}
	}
	if c.String() == before {
		t.Fatal("outline painted nothing")
	}
	got := UnoutlineStrokes(c, Rect{}, 1)
	if len(got) != 1 || got[0].Color != 0 || len(got[0].Pts) != 12 {
		t.Fatalf("unoutline = %+v, want 12 pixels erased", got)
	}
	for _, p := range got[0].Pts {
		c.Set(p[0], p[1], 0)
	}
	if c.String() != before {
		t.Fatalf("the canvas did not come back:\n%s\nwant\n%s", c, before)
	}

	c = square(9, 3, 3, 5, 5, 11)
	for _, s := range ShadeStrokes(c, Rect{}, 1, 1, 1) {
		for _, p := range s.Pts {
			c.Set(p[0], p[1], s.Color)
		}
	}
	shaded := c.String()
	for _, s := range OutlineStrokes(c, Rect{}, 0) {
		for _, p := range s.Pts {
			c.Set(p[0], p[1], s.Color)
		}
	}
	for _, s := range UnoutlineStrokes(c, Rect{}, 0) {
		for _, p := range s.Pts {
			c.Set(p[0], p[1], 0)
		}
	}
	if c.String() != shaded {
		t.Fatalf("unoutline ate the shadow band:\n%s\nwant\n%s", c, shaded)
	}
}

func TestShadeDarkensTheSideAwayFromTheLight(t *testing.T) {
	c := square(5, 1, 1, 3, 3, 11)
	got := ShadeStrokes(c, Rect{}, 1, 1, 1)
	if count(got) != 5 || got[0].Color != 13 {
		t.Fatalf("shade = %+v, want 5 dark-green pixels", got)
	}
	for _, p := range got[0].Pts {
		if p[0] != 3 && p[1] != 3 {
			t.Fatalf("shaded %v, which is on the lit side", p)
		}
	}
	two := ShadeStrokes(c, Rect{}, 1, 1, 2)
	if count(two) != 8 {
		t.Fatalf("shade two deep = %d pixels, want 8", count(two))
	}
	for _, p := range two[0].Pts {
		c.Set(p[0], p[1], 13)
	}
	again := ShadeStrokes(c, Rect{}, 1, 1, 1)
	if len(again) != 1 || again[0].Color != 13 || len(again[0].Pts) != 1 || again[0].Pts[0] != [2]int{1, 1} {
		t.Fatalf("a second shade pass = %+v, want only the last green pixel (1,1) in dark green", again)
	}
	got = ShadeStrokes(c, Rect{}, -1, 1, 1)
	for _, p := range got[0].Pts {
		if p[0] != 1 && p[1] != 3 {
			t.Fatalf("shaded %v with light from the top-right", p)
		}
	}
}

func TestShadeIgnoresBoundariesBetweenTwoFills(t *testing.T) {
	c := square(6, 0, 0, 5, 5, 11)
	for y := 2; y <= 3; y++ {
		for x := 2; x <= 3; x++ {
			c.Set(x, y, 16) // a peach patch in the middle
		}
	}
	got := ShadeStrokes(c, Rect{}, 1, 1, 1)
	for _, s := range got {
		for _, p := range s.Pts {
			if p[0] < 5 && p[1] < 5 {
				t.Fatalf("shaded %v next to the patch; only the canvas edge should shade", p)
			}
		}
	}
}

func TestHighlightLightensTheSideFacingTheLight(t *testing.T) {
	c := square(5, 1, 1, 3, 3, 11) // green lightens to lime
	got := HighlightStrokes(c, Rect{}, 1, 1, 1)
	if count(got) != 5 || got[0].Color != 10 {
		t.Fatalf("highlight = %+v, want 5 lime pixels", got)
	}
	for _, p := range got[0].Pts {
		if p[0] != 1 && p[1] != 1 {
			t.Fatalf("highlighted %v, which is on the shadow side", p)
		}
	}
	for _, p := range got[0].Pts {
		c.Set(p[0], p[1], 10)
	}
	if again := HighlightStrokes(c, Rect{}, 1, 1, 1); count(again) != 3 || again[0].Color != 10 {
		t.Fatalf("second highlight = %+v, want the band to grow inward by 3", again)
	}
}

func TestDitherCheckersOneColorInsideTheWindow(t *testing.T) {
	c := square(6, 0, 0, 5, 5, 11)
	c.Set(2, 2, 28)
	got := DitherStrokes(c, Rect{1, 1, 4, 4}, 11, 13)
	if len(got) != 1 || got[0].Color != 13 {
		t.Fatalf("dither = %+v", got)
	}
	if len(got[0].Pts) != 7 {
		t.Fatalf("dithered %d pixels, want 7", len(got[0].Pts))
	}
	for _, p := range got[0].Pts {
		if (p[0]+p[1])%2 != 0 || p[0] < 1 || p[0] > 4 || p[1] < 1 || p[1] > 4 || (p[0] == 2 && p[1] == 2) {
			t.Fatalf("dithered %v", p)
		}
	}
	if got := DitherStrokes(c, Rect{}, 20, 13); got != nil {
		t.Fatalf("dithering a color that is not there = %+v", got)
	}
}

func TestAsymmetryCountsPixelsThatDifferFromTheirReflection(t *testing.T) {
	c := NewCanvas(8, 4)
	if n, _ := Asymmetry(c); n != 0 {
		t.Fatalf("an empty canvas scored %d", n)
	}
	c.Set(1, 1, 9)
	c.Set(6, 1, 9) // its reflection: still symmetric
	if n, _ := Asymmetry(c); n != 0 {
		t.Fatalf("a mirrored pair scored %d", n)
	}
	c.Set(2, 3, 9)
	n, box := Asymmetry(c)
	if n != 1 || box.String() != "2,3-5,3" {
		t.Fatalf("one stray pixel scored %d in %s, want 1 in 2,3-5,3", n, box)
	}
}

func TestLightDirNamesTheFourCorners(t *testing.T) {
	for _, tc := range []struct {
		name   string
		dx, dy int
	}{{"", 1, 1}, {"top-left", 1, 1}, {"top-right", -1, 1}, {"bottom-left", 1, -1}, {"bottom-right", -1, -1}} {
		dx, dy, ok := LightDir(tc.name)
		if !ok || dx != tc.dx || dy != tc.dy {
			t.Fatalf("LightDir(%q) = %d,%d,%v", tc.name, dx, dy, ok)
		}
	}
	if _, _, ok := LightDir("sideways"); ok {
		t.Fatal("an unknown light was accepted")
	}
}

func TestShadeWorksInsideBlackOutline(t *testing.T) {
	for _, d := range [][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}} {
		c := square(12, 2, 2, 9, 9, 1)
		for y := 3; y <= 8; y++ {
			for x := 3; x <= 8; x++ {
				c.Set(x, y, 11)
			}
		}
		got := ShadeStrokes(c, Rect{}, d[0], d[1], 2)
		if count(got) != 20 {
			t.Fatalf("direction %v: shaded %d, want 20", d, count(got))
		}
		for _, s := range got {
			for _, p := range s.Pts {
				if c.At(p[0], p[1]) != 11 || s.Color != 13 {
					t.Fatalf("outline changed: %+v", s)
				}
			}
		}
	}
}

func TestShadeDoesNotTreatEnclosedBlackPupilAsOutline(t *testing.T) {
	c := square(11, 1, 1, 9, 9, 11)
	c.Set(5, 5, 1)
	for _, s := range ShadeStrokes(c, Rect{}, 1, 1, 1) {
		for _, p := range s.Pts {
			if p[0] != 9 && p[1] != 9 {
				t.Fatalf("shaded next to pupil: %v", p)
			}
		}
	}
}

func TestShadowPreservesForegroundAndClipsAtCanvasEdge(t *testing.T) {
	c := square(10, 2, 2, 5, 5, 1)
	for _, d := range [][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}} {
		got := shadowStrokes(c, Rect{}, d[0], d[1], 2, 1)
		if count(got) != 12 {
			t.Fatalf("direction %v: shadow %d, want 12", d, count(got))
		}
		for _, s := range got {
			for _, p := range s.Pts {
				if !c.in(p[0], p[1]) || c.At(p[0], p[1]) != 0 {
					t.Fatalf("shadow overwrites foreground: %v", p)
				}
			}
		}
	}
	c = square(4, 0, 0, 3, 3, 1)
	if got := shadowStrokes(c, Rect{}, 1, 1, 2, 1); len(got) != 0 {
		t.Fatal("full canvas has no room for a shadow")
	}
}
