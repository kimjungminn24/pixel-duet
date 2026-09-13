# pixelduet

A pixel canvas a model and a person share. Claude draws it stroke by
stroke while you watch in the browser, and at any moment you can pause
it, take the brush, change things, leave a note, and hand the brush
back. The reply to its next stroke tells it what you did. Or
just talk to it: the box under the log is a chat, and it listens.

## Install

No Go needed. The npm package downloads the prebuilt binary for your
machine (Linux, macOS, Windows; Node 18+):

    npx -y pixelduet version

Or take a binary from the Releases page, or build from source with
`go build -o pixelduet .` in a checkout.

## Quick start

One command starts the canvas server and the browser viewer, and opens
the viewer:

    npx -y pixelduet

Then let Claude Code in, once:

    claude mcp add pixelduet -- npx -y pixelduet bridge

That line is enough on its own: when a session starts and nothing is
running, the bridge starts the servers itself and tells the model
where the viewer is.

Working in a checkout of this repository, `go run .` stands in for
`npx -y pixelduet`, and Claude Code picks the bridge up from
`.mcp.json` automatically.

## Codex, and other MCP clients

The bridge is a plain MCP server over stdio, so nothing about it is
Claude's. Codex CLI takes the same line:

    codex mcp add pixelduet -- npx -y pixelduet bridge

or by hand in `~/.codex/config.toml`:

    [mcp_servers.pixelduet]
    command = "npx"
    args = ["-y", "pixelduet", "bridge"]

On Windows, where `npx` is a script and not an executable, wrap it:
`command = "cmd"`, `args = ["/c", "npx", "-y", "pixelduet", "bridge"]`.

Two things vary by client, and both are worth checking once:

- The drawing rules ride in the server's `instructions`, which some
  clients drop. `pixelduet rules` prints the same text, so
  `npx -y pixelduet rules >> AGENTS.md` puts it back.
- `get_canvas` answers with a PNG. A client that does not pass image
  results through leaves the model working blind, on `read_pixels`
  digits alone. It can still draw; it just cannot judge what it drew.

No MCP at all? The server speaks plain lines over TCP, so anything that
can open a socket can draw. The protocol is written down in internal/proto/proto.go.

## The handoff

The whole point is the pause loop:

1. The model draws, one stroke per call, narrating into the log.
2. You hit pause. Its next stroke is HELD: the server simply does not
   answer yet.
3. You edit pixels, pick colors, leave a note. Your clicks never wait.
4. You hit resume. The held stroke lands, and its reply says:

       ok changed=1 waited=12s edits=5 note=make the eyes bigger

5. The model reads the canvas, honors your edits, and keeps going.

There is no push channel to a model mid-thought. The trick is that the
answer it is already waiting for becomes the messenger.

## Talking to it

You do not have to pause to be heard. Type into the box under the log
at any time:

- While the model is painting, your words ride along on the reply to
  its next stroke (`ok changed=3 note=more blue in the sky`).
- When it has nothing to paint, it calls `listen`, and that call simply
  does not return until you say something. So "what should I draw
  next?" in the log means it is sitting there waiting for you.

Notes queue up, so type three lines and it gets all three, oldest first.
The model answers in the same log with `say`.

## The viewer

Pen, eraser, line, rectangle, circle, filled rectangle and circle, fill
and color-picker tools, undo (ctrl+z), an optional pixel grid, a palette
editor (edit any of the 32 slots, and every pixel painted with that slot
changes with it), a canvas size picker (32, 64 or 128, which starts over
blank), and a speed slider that slows the model down to watch it work.
The chat box under the log accepts any language. The page itself is in
Korean or English: it follows the browser's language, and the button at
the top right switches it.

The top bar has play and stop buttons for AI drawing. Magnifier buttons
resize the palette swatches. Canvas zoom uses magnifier buttons and a
slider, with a fit-to-window button. Select the move tool (H), or hold
Space and drag, to pan an enlarged canvas. AI stroke delay adds up to
three seconds between strokes; zero adds no extra wait.

The circle tools keep width and height equal while dragging, including
near canvas edges. O selects a circle; P selects a filled circle.

Choose a save folder with the native directory picker in Chrome or Edge,
then enter a file name to save a `.txt` sprite directly to that folder.
The choice lasts for the current page session. Existing files require
overwrite confirmation. The reset arrow switches back to the server's
configured gallery directory, which is also the default in browsers
without a directory picker. Choosing an export folder does not change
the MCP bridge's gallery directory.

## What the model gets

The bridge gives a model more than a bare socket would, because drawing
blind is how pictures come out boxy:

- `get_canvas` returns a rendered picture of the canvas, so it can look
  at what it drew and fix it; `read_pixels` returns a region as digits
  when it needs exact coordinates, or as run-length rows when those are
  shorter, which keeps a wide read cheap however big the canvas is.
- When you edit during a pause, its next reply says how many pixels
  changed and the box they are in, so it knows where to look.
- `draw_line`, `draw_rect` and `draw_ellipse` (outline or filled) let
  it think in shapes, and every drawing tool takes `mirror` to paint
  the reflection across the center line, so half a face becomes a face.
- `shade`, `highlight`, `outline` and `dither` are the finishing
  passes: darken the side away from a light you choose, lighten the
  side facing it, ring everything painted (black, or each color's own
  darker shade), checker two colors for a tone the palette lacks.
- A gallery. `sprites/` holds grids saved because someone liked them
  (the browser's save button, or `save_sprite`). The model lists and
  looks at them before it draws, as reference pictures, and can
  `stamp` one onto the canvas. It starts with one frog; it grows.
- `get_canvas` also reports how many pixels differ from their mirror
  image, so a face meant to look at you is caught being lopsided.
- The pixel-art house rules, handed to the model when it connects (see
  internal/bridge/rules.go): plan the proportions in numbers first, silhouette, check
  the picture after every step, ask the person before going on, one
  light source, three fixes before calling it done. They travel with the
  server, so they apply in any project and not only in a checkout.

## How it fits together

    pixelduet serve            owns the one true canvas
      |- pixelduet web         a bridge: browser <-> HTTP/SSE <-> lines
      `- pixelduet bridge      a bridge: an MCP client <-> tools <-> lines

`pixelduet` on its own runs serve and web together in one process, and
the bridge does the same when it finds no server to talk to.

The server speaks a line protocol you can read with your eyes (see
internal/proto/proto.go). The browser viewer is plain HTML, CSS and ES modules under
internal/web/static/, embedded into the binary with no build step. Every viewer and every bridge is just another client of it,
which is why adding the browser and the bridge never touched the server.

The canvas itself is a grid of palette indices, one character per pixel
(`0` empty, `1`-`w` colors), the same format on the wire, in save
files, and in what the model reads.

## Tests

    go test ./...
    node --test test/*.test.mjs

The Go tests drive a real hub over TCP; the node test covers the
viewer's shape math. See [CONTRIBUTING.md](CONTRIBUTING.md) for the
checks a pull request runs.

## License

MIT. See [LICENSE](LICENSE).

