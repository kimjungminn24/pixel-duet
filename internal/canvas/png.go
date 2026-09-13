package canvas

import (
	"bytes"
	"image"
	"image/png"
)

// checker is the pair of greys drawn for empty pixels.
var checker = [2]RGB{{38, 38, 42}, {58, 58, 64}}

const checkerSize = 2

func pixelColor(pal []RGB, c *Canvas, x, y int) RGB {
	if i := c.At(x, y); i != 0 {
		return pal[i]
	}
	return checker[(x/checkerSize+y/checkerSize)%2]
}

// RenderPNG draws the canvas at px pixels per cell.
func RenderPNG(c *Canvas, pal []RGB, px int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, c.W*px, c.H*px))
	for y := 0; y < c.H; y++ {
		top := y * px * img.Stride
		row := img.Pix[top : top+img.Stride]
		for x := 0; x < c.W; x++ {
			rgb := pixelColor(pal, c, x, y)
			for dx := 0; dx < px; dx++ {
				i := (x*px + dx) * 4
				row[i], row[i+1], row[i+2], row[i+3] = rgb.R, rgb.G, rgb.B, 255
			}
		}
		for dy := 1; dy < px; dy++ {
			copy(img.Pix[top+dy*img.Stride:], row)
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// PNGScale keeps the long side near 512px.
func PNGScale(c *Canvas) int {
	return max(2, 512/max(c.W, c.H))
}
