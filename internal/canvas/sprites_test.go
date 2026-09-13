package canvas

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpritesRoundTripAndKeepTheirFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sprites")
	c := NewCanvas(4, 3)
	c.Set(1, 1, 9)
	c.Set(3, 2, 12)
	path, err := SaveSprite(dir, "dot", c)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if !filepath.IsAbs(path) || filepath.Base(path) != "dot.txt" {
		t.Fatalf("save reported %q, want an absolute path ending in dot.txt", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("nothing at the reported path: %v", err)
	}
	got, err := LoadSprite(dir, "dot")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.String() != c.String() {
		t.Fatalf("loaded sprite =\n%s\nwant\n%s", got, c)
	}
	list, err := ListSprites(dir)
	if err != nil || len(list) != 1 || list[0] != (SpriteInfo{"dot", 4, 3}) {
		t.Fatalf("list = %+v, %v", list, err)
	}
	for _, bad := range []string{"../etc", "a/b", "", "sp ace"} {
		if _, err := SaveSprite(dir, bad, c); err == nil {
			t.Fatalf("saved a sprite named %q", bad)
		}
	}
	if list, err := ListSprites(filepath.Join(dir, "missing")); err != nil || len(list) != 0 {
		t.Fatalf("a missing folder should list as empty, got %+v, %v", list, err)
	}
}

func TestStampSkipsEmptyPixels(t *testing.T) {
	s := NewCanvas(2, 2)
	s.Set(0, 0, 9)
	s.Set(1, 1, 12)
	got := StampStrokes(s, 10, 20)
	if count(got) != 2 || !has(got[0].Pts, 10, 20) || !has(got[1].Pts, 11, 21) {
		t.Fatalf("stamp = %+v", got)
	}
	c := Crop(NewCanvas(8, 8), Rect{6, 6, 4, 4})
	if c.W != 4 || c.H != 4 {
		t.Fatalf("crop past the edge = %dx%d, want the asked-for 4x4", c.W, c.H)
	}
}
