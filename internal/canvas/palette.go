package canvas

import "fmt"

type RGB struct{ R, G, B uint8 }

func (c RGB) Hex() string { return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B) }

func ParseHex(s string) (RGB, error) {
	var c RGB
	if len(s) != 6 {
		return c, fmt.Errorf("bad color %q: want rrggbb", s)
	}
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
		return c, fmt.Errorf("bad color %q: %w", s, err)
	}
	return c, nil
}

const PaletteSize = 33

// Palette is DB32. Slot 0 is transparent; viewers draw a checkerboard for it.
var Palette = []RGB{
	{}, // 0: empty
	{0, 0, 0}, {34, 32, 52}, {69, 40, 60}, {102, 57, 49},
	{143, 86, 59}, {223, 113, 38}, {217, 160, 102}, {238, 195, 154},
	{251, 242, 54}, {153, 229, 80}, {106, 190, 48}, {55, 148, 110},
	{75, 105, 47}, {82, 75, 36}, {50, 60, 57}, {63, 63, 116},
	{48, 96, 130}, {91, 110, 225}, {99, 155, 255}, {95, 205, 228},
	{203, 219, 252}, {255, 255, 255}, {155, 173, 183}, {132, 126, 135},
	{105, 106, 106}, {89, 86, 82}, {118, 66, 138}, {172, 50, 50},
	{217, 87, 99}, {215, 123, 186}, {143, 151, 74}, {138, 111, 48},
}

// digits: one character per palette slot.
const digits = "0123456789abcdefghijklmnopqrstuvw"

func ValidColor(c int) bool { return c >= 0 && c < PaletteSize }

// Shade and Light map each slot one step down or up its DB32 ramp.
// A slot mapped to itself is at the end of its ramp.
var (
	Shade = [PaletteSize]uint8{
		0, 1, 1, 2, 3, 4, 5, 5, 7, 6, 11, 13, 15, 15, 15, 2, 2, 16, 16, 17,
		19, 20, 23, 24, 25, 26, 1, 3, 3, 28, 27, 14, 14,
	}
	Light = [PaletteSize]uint8{
		0, 26, 16, 4, 5, 7, 9, 8, 22, 22, 9, 10, 20, 11, 32, 12, 18, 19, 19,
		20, 21, 22, 22, 22, 23, 24, 25, 30, 29, 30, 30, 10, 31,
	}
)
