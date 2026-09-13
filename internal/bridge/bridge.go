package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

type bridge struct {
	viewer  string // viewer URL when this process started the servers
	conn    *hubConn
	sprites string
}

// Run serves the MCP tools over stdio. viewer is the browser URL to mention, or empty.
func Run(addr, sprites, viewer, version string) error {
	b := &bridge{conn: &hubConn{addr: addr}, sprites: sprites, viewer: viewer}
	s := mcp.NewServer(&mcp.Implementation{Name: "pixelduet", Version: version},
		&mcp.ServerOptions{Instructions: Rules})
	b.register(s)
	return s.Run(context.Background(), &mcp.StdioTransport{})
}

func (b *bridge) register(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "get_canvas", Description: "The canvas as a picture, with its " +
		"size, whether it is paused, and how many pixels differ from their mirror image. Name a " +
		"region with x,y,w,h to fill the same picture with just that part: a 14 wide face comes " +
		"back at 36 screen pixels per canvas pixel instead of 8, which is the only way to judge " +
		"an eye."}, b.getCanvas)
	mcp.AddTool(s, &mcp.Tool{Name: "read_pixels", Description: "The canvas as digits, for exact " +
		"coordinates: a picture cannot tell you whether an eye is in column 11 or 12, this can. " +
		"Name a region with x,y,w,h; none means the whole canvas. A wide read comes back as " +
		"run-length rows unless you ask otherwise."}, b.readPixels)
	mcp.AddTool(s, &mcp.Tool{Name: "draw_pixels", Description: "Paint one stroke of a single color, " +
		"pixel by pixel. mirror also paints the reflection across the vertical center. If the " +
		"canvas is paused the call waits, and its reply says what the person changed and said."}, b.drawPixels)
	mcp.AddTool(s, &mcp.Tool{Name: "draw_line", Description: "Paint a straight line of one color " +
		"from (x0,y0) to (x1,y1), one pixel thick with no gaps. mirror also paints the reflection " +
		"across the vertical center."}, b.drawLine)
	mcp.AddTool(s, &mcp.Tool{Name: "draw_rect", Description: "Paint a rectangle given two opposite " +
		"corners, outline or filled."}, b.drawRect)
	mcp.AddTool(s, &mcp.Tool{Name: "draw_ellipse", Description: "Paint an ellipse given its center " +
		"and radii, outline or filled. Centers take halves, so cx 15.5 is the middle of a 32 wide " +
		"canvas, and rx=ry=3 gives a 7 wide circle."}, b.drawEllipse)
	mcp.AddTool(s, &mcp.Tool{Name: "fill_area", Description: "Flood fill from one pixel, replacing " +
		"the connected patch of its color."}, b.fillArea)
	mcp.AddTool(s, &mcp.Tool{Name: "shade", Description: "Darken everything painted on the side " +
		"facing away from the light, one step down the palette (green to dark green, blue to navy, " +
		"orange to brown). Boundaries between two different fills are left alone, so a belly stays " +
		"a belly."}, b.shade)
	mcp.AddTool(s, &mcp.Tool{Name: "outline", Description: "Ring everything painted with a one " +
		"pixel border, or take the ring off with off. color 1 is the classic black border; no " +
		"color at all gives each shape the darker shade of its own fill. It outlines against " +
		"transparency, so it runs before any background fill, and calling it twice rings the " +
		"first ring."}, b.outline)
	mcp.AddTool(s, &mcp.Tool{Name: "highlight", Description: "Lighten everything painted on the " +
		"side facing the light, one step up the palette (green to yellow, navy to blue, grey to " +
		"silver). The opposite of shade; give it the same light."}, b.highlight)
	mcp.AddTool(s, &mcp.Tool{Name: "dither", Description: "Checker two colors inside a strip: " +
		"every other pixel of from becomes to. It makes a tone between two palette colors."}, b.dither)
	mcp.AddTool(s, &mcp.Tool{Name: "list_sprites", Description: "List the gallery: pictures saved " +
		"from earlier sessions because the person liked them."}, b.listSprites)
	mcp.AddTool(s, &mcp.Tool{Name: "view_sprite", Description: "One saved sprite, as a picture and " +
		"as digits."}, b.viewSprite)
	mcp.AddTool(s, &mcp.Tool{Name: "save_sprite", Description: "Save the canvas, or a region of it, " +
		"to the gallery under a name."}, b.saveSprite)
	mcp.AddTool(s, &mcp.Tool{Name: "stamp", Description: "Place a saved sprite with its top-left " +
		"corner at (x,y). Its empty pixels leave what is under them alone."}, b.stamp)
	mcp.AddTool(s, &mcp.Tool{Name: "say", Description: "Show one short line in the log panel beside " +
		"the canvas. What the person types appears in the same place."}, b.say)
	mcp.AddTool(s, &mcp.Tool{Name: "listen", Description: "Wait for the person to say something in " +
		"the chat beside the canvas. It comes back empty if the wait window runs out."}, b.listen)
	mcp.AddTool(s, &mcp.Tool{Name: "clear_canvas", Description: "Wipe the whole canvas back to " +
		"transparent. Never use it to erase what the person drew."}, b.clearCanvas)
}

// reportDraw turns a reply line into the tool result the model reads.
func (b *bridge) reportDraw(res string) (*mcp.CallToolResult, any, error) {
	if res == "paused" {
		return textResult("The canvas is paused and the wait window ran out. " +
			"Nothing was drawn. Call the tool again to keep waiting."), nil, nil
	}
	if msg, failed := strings.CutPrefix(res, "err "); failed {
		return nil, nil, fmt.Errorf("%s", msg)
	}
	p := proto.ParseReply(res)
	switch {
	case p.Edits > 0:
		msg := "THE PERSON PAUSED YOU AND TOOK THE BRUSH: " + res + "\n" +
			"Their edits sit inside box= (x0,y0-x1,y1) in the picture below. Honor what they " +
			"said and their edits - do not paint over them - then continue drawing. For the " +
			"exact pixels, read_pixels that box.\n"
		c, _, pal, err := b.conn.canvas()
		if err != nil {
			return textResult(msg), nil, nil
		}
		return pictureResult(msg, c, pal), nil, nil
	case p.HasNote:
		return textResult("THE PERSON SAYS: " + p.Note + "\nThe stroke landed (" + res + "). " +
			"Answer with say, then act on what they said."), nil, nil
	}
	return textResult(res), nil, nil
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func pictureResult(text string, c *canvas.Canvas, pal []canvas.RGB) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: text},
		&mcp.ImageContent{Data: canvas.RenderPNG(c, pal, canvas.PNGScale(c)), MIMEType: "image/png"},
	}}
}
