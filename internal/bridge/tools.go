package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

func (b *bridge) getCanvas(ctx context.Context, req *mcp.CallToolRequest, in viewArgs) (*mcp.CallToolResult, any, error) {
	c, paused, pal, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	shown, zoom := c, ""
	if r := in.rect(); r.W > 0 && r.H > 0 {
		shown = canvas.Crop(c, r)
		zoom = fmt.Sprintf("zoomed on x=%d y=%d w=%d h=%d at %d per canvas pixel\n",
			r.X, r.Y, r.W, r.H, canvas.PNGScale(shown))
	}
	state := "PLAYING"
	if paused {
		state = "PAUSED - your next stroke will wait for the person"
	}
	axis := float64(c.W-1) / 2
	mirror := fmt.Sprintf("mirror check: symmetric about x=%g", axis)
	if n, box := canvas.Asymmetry(c); n > 0 {
		mirror = fmt.Sprintf("mirror check: %d pixels differ from their reflection about x=%g, inside %s "+
			"(shade and highlight are one-sided by design, so judge this before them; a figure "+
			"facing you should be at 0 until then)", n, axis, box)
	}
	text := fmt.Sprintf("canvas %dx%d  %s\n%s%s\n"+
		"Judge shape and color from the picture below; for exact coordinates use read_pixels.\n",
		c.W, c.H, state, zoom, mirror)
	if b.viewer != "" {
		text += fmt.Sprintf("The person watches at %s; mention it if they seem not to have it open.\n", b.viewer)
	}
	return pictureResult(text, shown, pal), nil, nil
}

func (b *bridge) readPixels(ctx context.Context, req *mcp.CallToolRequest, in readArgs) (*mcp.CallToolResult, any, error) {
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	r := in.rect().Clip(c)
	grid, runs := c.Digits(r), c.Runs(r)
	compact := len(runs) < len(grid)
	if in.Compact != nil {
		compact = *in.Compact
	}
	body, form := grid, "one digit per pixel"
	if compact {
		body, form = runs, "run-length rows (0*8 4*16 0*8)"
	}
	head := fmt.Sprintf("x=%d y=%d w=%d h=%d as %s (first row below is y=%d, first column x=%d)\n",
		r.X, r.Y, r.W, r.H, form, r.Y, r.X)
	return textResult(head + body), nil, nil
}

func (b *bridge) drawPixels(ctx context.Context, req *mcp.CallToolRequest, in strokeArgs) (*mcp.CallToolResult, any, error) {
	pts := make([][2]int, 0, len(in.Points))
	for _, p := range in.Points {
		pts = append(pts, [2]int{p.X, p.Y})
	}
	return b.paint(in.Color, pts, in.Mirror)
}

func (b *bridge) drawLine(ctx context.Context, req *mcp.CallToolRequest, in lineArgs) (*mcp.CallToolResult, any, error) {
	return b.paint(in.Color, canvas.LinePoints(in.X0, in.Y0, in.X1, in.Y1), in.Mirror)
}

func (b *bridge) drawRect(ctx context.Context, req *mcp.CallToolRequest, in rectArgs) (*mcp.CallToolResult, any, error) {
	return b.paint(in.Color, canvas.RectPoints(in.X0, in.Y0, in.X1, in.Y1, in.Filled), in.Mirror)
}

func (b *bridge) drawEllipse(ctx context.Context, req *mcp.CallToolRequest, in ellipseArgs) (*mcp.CallToolResult, any, error) {
	return b.paint(in.Color, canvas.EllipsePoints(in.CX, in.CY, in.RX, in.RY, in.Filled), in.Mirror)
}

func (b *bridge) fillArea(ctx context.Context, req *mcp.CallToolRequest, in fillArgs) (*mcp.CallToolResult, any, error) {
	res, err := b.conn.cmd(fmt.Sprintf("fill %d %d %d", in.X, in.Y, in.Color))
	if err != nil {
		return nil, nil, err
	}
	return b.reportDraw(res)
}

func (b *bridge) clearCanvas(ctx context.Context, req *mcp.CallToolRequest, in any) (*mcp.CallToolResult, any, error) {
	res, err := b.conn.cmd("clear")
	if err != nil {
		return nil, nil, err
	}
	return b.reportDraw(res)
}

// paint sends one stroke; every shape tool ends here.
func (b *bridge) paint(color int, pts [][2]int, mirror bool) (*mcp.CallToolResult, any, error) {
	if !canvas.ValidColor(color) {
		return nil, nil, fmt.Errorf("color %d is not a palette index", color)
	}
	if len(pts) == 0 {
		return nil, nil, fmt.Errorf("a stroke needs at least one point")
	}
	if mirror {
		// Width is read fresh because the person can resize at any time.
		c, _, _, err := b.conn.canvas()
		if err != nil {
			return nil, nil, err
		}
		pts = canvas.MirrorX(pts, c.W)
	}
	res, err := b.conn.cmd(proto.EncodeStroke(uint8(color), pts))
	if err != nil {
		return nil, nil, err
	}
	return b.reportDraw(res)
}

func (b *bridge) outline(ctx context.Context, req *mcp.CallToolRequest, in outlineArgs) (*mcp.CallToolResult, any, error) {
	if !canvas.ValidColor(in.Color) {
		return nil, nil, fmt.Errorf("color %d is not a palette index", in.Color)
	}
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	if in.Off {
		return b.finish("un-outlined", canvas.UnoutlineStrokes(c, in.rect(), uint8(in.Color)))
	}
	return b.finish("outlined", canvas.OutlineStrokes(c, in.rect(), uint8(in.Color)))
}

func (b *bridge) shade(ctx context.Context, req *mcp.CallToolRequest, in shadeArgs) (*mcp.CallToolResult, any, error) {
	dx, dy, err := b.light(in.Light)
	if err != nil {
		return nil, nil, err
	}
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	return b.finish("shaded", canvas.ShadeStrokes(c, in.rect(), dx, dy, in.Width))
}

func (b *bridge) highlight(ctx context.Context, req *mcp.CallToolRequest, in highlightArgs) (*mcp.CallToolResult, any, error) {
	dx, dy, err := b.light(in.Light)
	if err != nil {
		return nil, nil, err
	}
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	return b.finish("highlighted", canvas.HighlightStrokes(c, in.rect(), dx, dy, in.Width))
}

func (b *bridge) dither(ctx context.Context, req *mcp.CallToolRequest, in ditherArgs) (*mcp.CallToolResult, any, error) {
	for _, col := range []int{in.From, in.To} {
		if !canvas.ValidColor(col) {
			return nil, nil, fmt.Errorf("color %d is not a palette index", col)
		}
	}
	if in.W <= 0 || in.H <= 0 {
		return nil, nil, fmt.Errorf("dither needs a strip: x, y, w and h")
	}
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	return b.finish("dithered", canvas.DitherStrokes(c, canvas.Rect{X: in.X, Y: in.Y, W: in.W, H: in.H}, uint8(in.From), uint8(in.To)))
}

func (b *bridge) light(name string) (dx, dy int, err error) {
	dx, dy, ok := canvas.LightDir(name)
	if !ok {
		return 0, 0, fmt.Errorf("light must be top-left, top-right, bottom-left or bottom-right, not %q", name)
	}
	return dx, dy, nil
}

// finish sends strokes one per color. A reply carrying edits or a note wins over the pixel count.
func (b *bridge) finish(what string, strokes []canvas.Stroke) (*mcp.CallToolResult, any, error) {
	total, held := 0, ""
	for _, s := range strokes {
		res, err := b.conn.cmd(proto.EncodeStroke(s.Color, s.Pts))
		if err != nil {
			return nil, nil, err
		}
		if res == "paused" || strings.HasPrefix(res, "err ") {
			return b.reportDraw(res)
		}
		if p := proto.ParseReply(res); p.Edits > 0 || p.HasNote {
			held = res
		}
		total += len(s.Pts)
	}
	if held != "" {
		return b.reportDraw(held)
	}
	return textResult(fmt.Sprintf("%s %d pixels", what, total)), nil, nil
}

func (b *bridge) listSprites(ctx context.Context, req *mcp.CallToolRequest, in any) (*mcp.CallToolResult, any, error) {
	list, err := canvas.ListSprites(b.sprites)
	if err != nil {
		return nil, nil, err
	}
	if len(list) == 0 {
		return textResult(fmt.Sprintf("The gallery at %s is empty. When the person likes a picture, "+
			"save_sprite it and the next session starts from something.", canvas.AbsPath(b.sprites))), nil, nil
	}
	var text strings.Builder
	fmt.Fprintf(&text, "%d sprites in %s - view_sprite one to see it, stamp to place it:\n",
		len(list), canvas.AbsPath(b.sprites))
	for _, s := range list {
		fmt.Fprintf(&text, "  %s  %dx%d\n", s.Name, s.W, s.H)
	}
	return textResult(text.String()), nil, nil
}

func (b *bridge) viewSprite(ctx context.Context, req *mcp.CallToolRequest, in spriteArgs) (*mcp.CallToolResult, any, error) {
	s, err := canvas.LoadSprite(b.sprites, in.Name)
	if err != nil {
		return nil, nil, err
	}
	path, _ := canvas.SpritePath(b.sprites, in.Name)
	text := fmt.Sprintf("sprite %s  %dx%d  %s\n%s", in.Name, s.W, s.H, canvas.AbsPath(path), s)
	return pictureResult(text, s, canvas.Palette), nil, nil
}

func (b *bridge) saveSprite(ctx context.Context, req *mcp.CallToolRequest, in saveSpriteArgs) (*mcp.CallToolResult, any, error) {
	c, _, _, err := b.conn.canvas()
	if err != nil {
		return nil, nil, err
	}
	s := canvas.Crop(c, in.rect())
	path, err := canvas.SaveSprite(b.sprites, in.Name, s)
	if err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("saved %s (%dx%d) to %s - tell the person that path",
		in.Name, s.W, s.H, path)), nil, nil
}

func (b *bridge) stamp(ctx context.Context, req *mcp.CallToolRequest, in stampArgs) (*mcp.CallToolResult, any, error) {
	s, err := canvas.LoadSprite(b.sprites, in.Name)
	if err != nil {
		return nil, nil, err
	}
	return b.finish("stamped", canvas.StampStrokes(s, in.X, in.Y))
}

func (b *bridge) say(ctx context.Context, req *mcp.CallToolRequest, in sayArgs) (*mcp.CallToolResult, any, error) {
	if _, err := b.conn.cmd("say " + in.Text); err != nil {
		return nil, nil, err
	}
	return textResult("said"), nil, nil
}

// listen blocks until the person speaks or the server's wait window ends.
func (b *bridge) listen(ctx context.Context, req *mcp.CallToolRequest, in any) (*mcp.CallToolResult, any, error) {
	res, err := b.conn.cmd("listen")
	if err != nil {
		return nil, nil, err
	}
	if msg, failed := strings.CutPrefix(res, "err "); failed {
		return nil, nil, fmt.Errorf("%s", msg)
	}
	p := proto.ParseReply(res)
	if !p.HasNote {
		return textResult("The person has not said anything yet and the wait window ran " +
			"out. Call listen again to keep waiting for them."), nil, nil
	}
	return textResult("THE PERSON SAYS: " + p.Note + "\nAnswer with say, then act on it."), nil, nil
}
