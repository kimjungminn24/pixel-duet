package canvas

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestRenderPNGShowsPaletteColors(t *testing.T) {
	c := NewCanvas(4, 4)
	c.Set(1, 2, 9) // red
	pal := append([]RGB(nil), Palette...)
	pal[9] = RGB{10, 20, 30} // the person recolored slot 9

	img, err := png.Decode(bytes.NewReader(RenderPNG(c, pal, 8)))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 32 || b.Dy() != 32 {
		t.Fatalf("picture is %dx%d, want 32x32", b.Dx(), b.Dy())
	}
	r, g, bl, _ := img.At(1*8+3, 2*8+3).RGBA()
	if r>>8 != 10 || g>>8 != 20 || bl>>8 != 30 {
		t.Fatalf("painted pixel = %d,%d,%d, want the live palette color 10,20,30", r>>8, g>>8, bl>>8)
	}
	r, g, bl, _ = img.At(0, 0).RGBA()
	if uint8(r>>8) != checker[0].R || uint8(g>>8) != checker[0].G || uint8(bl>>8) != checker[0].B {
		t.Fatalf("empty pixel = %d,%d,%d, want the checkerboard", r>>8, g>>8, bl>>8)
	}
}

func TestPNGScaleKeepsPicturesAround512(t *testing.T) {
	for _, tc := range []struct{ size, want int }{{32, 16}, {64, 8}, {128, 4}} {
		if got := PNGScale(NewCanvas(tc.size, tc.size)); got != tc.want {
			t.Fatalf("scale for %d = %d, want %d", tc.size, got, tc.want)
		}
	}
}

func TestEmptyPixelsShowAPatternPaintedBlackDoesNot(t *testing.T) {
	grey := func(img image.Image, x, y int) RGB {
		r, g, b, _ := img.At(x, y).RGBA()
		return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
	}
	decode := func(c *Canvas) image.Image {
		img, err := png.Decode(bytes.NewReader(RenderPNG(c, Palette, 4)))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		return img
	}

	empty := decode(NewCanvas(8, 8))
	if grey(empty, 0, 0) == grey(empty, checkerSize*4, 0) {
		t.Fatal("neighbouring checker squares should not be the same grey")
	}

	c := NewCanvas(8, 8)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			c.Set(x, y, 1) // 1 is black, a real color
		}
	}
	if got, want := pixelColor(Palette, c, 0, 0), (RGB{0, 0, 0}); got != want {
		t.Fatalf("painted black rendered as %+v, want %+v", got, want)
	}
	painted := decode(c)
	b := painted.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if grey(painted, x, y) == checker[1] {
				t.Fatalf("black canvas shows the empty pattern at %d,%d", x, y)
			}
		}
	}
}
