package bridge

import "github.com/kimjungminn24/pixel-duet/internal/canvas"

// region is x,y,w,h; all zero means the whole canvas.
type region struct {
	X int `json:"x,omitempty" jsonschema:"region left; omit x,y,w,h for the whole canvas"`
	Y int `json:"y,omitempty" jsonschema:"region top"`
	W int `json:"w,omitempty" jsonschema:"region width"`
	H int `json:"h,omitempty" jsonschema:"region height"`
}

func (r region) rect() canvas.Rect { return canvas.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H} }

type lit struct {
	Light string `json:"light,omitempty" jsonschema:"where the light comes from: top-left (default), top-right, bottom-left, bottom-right"`
	Width int    `json:"width,omitempty" jsonschema:"how many pixels deep the band is; 1 by default"`
}

type point struct {
	X int `json:"x" jsonschema:"column, 0 at the left"`
	Y int `json:"y" jsonschema:"row, 0 at the top"`
}

type strokeArgs struct {
	Color  int     `json:"color" jsonschema:"palette index 1-32, or 0 to erase"`
	Points []point `json:"points" jsonschema:"pixels this one stroke covers"`
	Mirror bool    `json:"mirror,omitempty" jsonschema:"also paint the mirrored side"`
}

type lineArgs struct {
	Color  int  `json:"color" jsonschema:"palette index 1-32, or 0 to erase"`
	X0     int  `json:"x0" jsonschema:"start column"`
	Y0     int  `json:"y0" jsonschema:"start row"`
	X1     int  `json:"x1" jsonschema:"end column"`
	Y1     int  `json:"y1" jsonschema:"end row"`
	Mirror bool `json:"mirror,omitempty" jsonschema:"also paint the mirrored side"`
}

type rectArgs struct {
	Color  int  `json:"color" jsonschema:"palette index 1-32, or 0 to erase"`
	X0     int  `json:"x0" jsonschema:"one corner's column"`
	Y0     int  `json:"y0" jsonschema:"one corner's row"`
	X1     int  `json:"x1" jsonschema:"the opposite corner's column, inclusive"`
	Y1     int  `json:"y1" jsonschema:"the opposite corner's row, inclusive"`
	Filled bool `json:"filled,omitempty" jsonschema:"paint the inside too, not just the outline"`
	Mirror bool `json:"mirror,omitempty" jsonschema:"also paint the mirrored side"`
}

type ellipseArgs struct {
	Color  int     `json:"color" jsonschema:"palette index 1-32, or 0 to erase"`
	CX     float64 `json:"cx" jsonschema:"center column, halves allowed"`
	CY     float64 `json:"cy" jsonschema:"center row, halves allowed"`
	RX     float64 `json:"rx" jsonschema:"horizontal radius, halves allowed"`
	RY     float64 `json:"ry" jsonschema:"vertical radius, halves allowed"`
	Filled bool    `json:"filled,omitempty" jsonschema:"paint the inside too, not just the outline"`
	Mirror bool    `json:"mirror,omitempty" jsonschema:"also paint the mirrored side"`
}

type fillArgs struct {
	X     int `json:"x" jsonschema:"column to start the fill from"`
	Y     int `json:"y" jsonschema:"row to start the fill from"`
	Color int `json:"color" jsonschema:"palette index to fill with"`
}

type shadeArgs struct {
	lit
	region
}

type highlightArgs struct {
	lit
	region
}

type outlineArgs struct {
	Color int  `json:"color,omitempty" jsonschema:"palette index to ring in; omit for each shape's own darker shade"`
	Off   bool `json:"off,omitempty" jsonschema:"take the outline back off instead of putting one on"`
	region
}

type ditherArgs struct {
	From int `json:"from" jsonschema:"the color to thin out"`
	To   int `json:"to" jsonschema:"the color to mix in on every other pixel"`
	X    int `json:"x" jsonschema:"left column of the strip to dither"`
	Y    int `json:"y" jsonschema:"top row of the strip"`
	W    int `json:"w" jsonschema:"width of the strip"`
	H    int `json:"h" jsonschema:"height of the strip"`
}

type viewArgs struct {
	region
}

type readArgs struct {
	region
	Compact *bool `json:"compact,omitempty" jsonschema:"leave unset for whichever form is shorter; true forces run-length rows (0*8 4*16 0*8), false forces one digit per pixel"`
}

type spriteArgs struct {
	Name string `json:"name" jsonschema:"the sprite's name, as list_sprites shows it"`
}

type saveSpriteArgs struct {
	Name string `json:"name" jsonschema:"letters, digits, - and _; saving over a name replaces it"`
	region
}

type stampArgs struct {
	Name string `json:"name" jsonschema:"the sprite to place"`
	X    int    `json:"x" jsonschema:"column its left edge lands on"`
	Y    int    `json:"y" jsonschema:"row its top edge lands on"`
}

type sayArgs struct {
	Text string `json:"text" jsonschema:"one short line to show beside the canvas"`
}
