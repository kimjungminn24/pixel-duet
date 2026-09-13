package canvas

import "testing"

func TestSetReportsWhetherThePixelChanged(t *testing.T) {
	c := NewCanvas(4, 3)
	if !c.Set(1, 2, 9) {
		t.Fatal("first write should report a change")
	}
	if c.Set(1, 2, 9) {
		t.Fatal("writing the same color again is not a change")
	}
	if got := c.At(1, 2); got != 9 {
		t.Fatalf("At(1,2) = %d, want 9", got)
	}
	if c.Set(-1, 0, 5) || c.Set(0, 0, 99) {
		t.Fatal("out of range writes should be dropped, not applied")
	}
}

func TestStringAndParseGridRoundTrip(t *testing.T) {
	c := NewCanvas(3, 2)
	c.Set(0, 0, 10)
	c.Set(2, 1, 15)
	if got, want := c.String(), "a00\n00f\n"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	back, err := ParseGrid(c.String())
	if err != nil {
		t.Fatalf("ParseGrid: %v", err)
	}
	if back.String() != c.String() {
		t.Fatalf("round trip changed the grid:\n%q\n%q", back.String(), c.String())
	}
	if _, err := ParseGrid("zz\nzz\n"); err == nil {
		t.Fatal("expected an error for non-palette digits")
	}
}

func TestFillStopsAtPixelsOfAnotherColor(t *testing.T) {
	c := NewCanvas(3, 3)
	for y := 0; y < 3; y++ {
		c.Set(1, y, 7) // a wall down the middle column
	}
	n, box := c.Fill(0, 0, 3)
	if n != 3 {
		t.Fatalf("Fill changed %d pixels, want 3", n)
	}
	if box.String() != "0,0-0,2" {
		t.Fatalf("Fill box = %s, want the left column 0,0-0,2", box)
	}
	if got := c.At(0, 2); got != 3 {
		t.Fatalf("left side not filled: At(0,2) = %d", got)
	}
	if got := c.At(2, 0); got != 0 {
		t.Fatalf("fill leaked past the wall: At(2,0) = %d", got)
	}
}

func TestCompactFoldsRunsAndRepeatedRows(t *testing.T) {
	c := NewCanvas(8, 4)
	for y := 1; y <= 2; y++ {
		for x := 2; x <= 5; x++ {
			c.Set(x, y, 4)
		}
	}
	c.Set(7, 3, 9)
	want := "y0: 0*8\ny1-2: 0*2 4*4 0*2\ny3: 0*7 9\n"
	if got := c.Runs(Rect{}); got != want {
		t.Fatalf("Compact =\n%s\nwant\n%s", got, want)
	}
	if got := c.Runs(Rect{4, 2, 4, 2}); got != "y2: 4*2 0*2\ny3: 0*3 9\n" {
		t.Fatalf("Compact of a window = %q", got)
	}
}

func TestRunLengthRowsWinOnFlatAreasAndLoseOnNoisyOnes(t *testing.T) {
	flat := NewCanvas(32, 32)
	for y := 4; y < 20; y++ {
		for x := 4; x < 20; x++ {
			flat.Set(x, y, 9)
		}
	}
	if runs, grid := flat.Runs(Rect{}), flat.Digits(Rect{}); len(runs) >= len(grid) {
		t.Fatalf("runs %d chars, digits %d: runs should win on a flat canvas", len(runs), len(grid))
	}

	noisy := NewCanvas(8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if (x+y)%2 == 0 {
				noisy.Set(x, y, 9)
			}
		}
	}
	if runs, grid := noisy.Runs(Rect{}), noisy.Digits(Rect{}); len(runs) <= len(grid) {
		t.Fatalf("runs %d chars, digits %d: digits should win on a checkerboard", len(runs), len(grid))
	}
}
