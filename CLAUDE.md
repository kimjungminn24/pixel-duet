# pixelduet

A pixel canvas shared between a model and a person. The server owns the
canvas; the browser viewer and the drawing clients all just connect.

The drawing rules a model follows are not in this file. They live in
internal/bridge/rules.go and are handed to it in the bridge's
initialize reply, because CLAUDE.md only loads inside a checkout while
the people who use pixelduet are in their own projects. Edit them there.

## Running it

    go run .              canvas server + browser viewer, one process
    go run . bridge       the tool bridge, over stdio

A session in this checkout is already connected: Claude Code reads
.mcp.json, which runs the bridge, and the bridge starts the servers
itself if none are up.

## Tests

    go test ./...
    node --test test/*.test.mjs

Most Go tests start a real hub and talk to it over TCP, so they cover
the line protocol and not only the functions under it. None needs a
browser. The node tests cover the viewer's shape math, which mirrors
shapes.go, and its translations.

## Layout

    main.go               the subcommands
    internal/canvas/      the grid, DB32 palette, shapes, finishing passes, PNG, sprite files
    internal/proto/       the line protocol: readers, writers, reply parsing
    internal/hub/         the server: state under one lock, one goroutine per client
    internal/local/       hub and viewer in one process; used by start and bridge
    internal/web/         HTTP and SSE for the viewer; static/ is the page itself
    internal/bridge/      the MCP server, its tools, and the drawing rules
    test/                 node tests for the viewer's pure functions
    npm/                  the wrapper that ships prebuilt binaries

Dependencies point one way: canvas <- proto <- hub <- local, and web
and bridge sit beside hub on top of proto. main.go is the only package
that imports everything.

The viewer's scripts are ES modules under internal/web/static/js.
state.js holds the one shared state object; api.js is the only file
that talks HTTP; geometry.js is pure functions of coordinates; i18n.js
holds every word the page shows; the rest each own one part of the
page and export an init function that main.js calls.

## Conventions

Comments say why, not what, and stay to one or two lines. No asides set
off with dashes, no first person.

A new command touches proto first, then hub, then whichever clients
should offer it. The hub never learns that browsers or models exist.

Everything that reads from the hub goes through proto.ReadMessage, so a
new push type is added there once and every client skips it correctly.

The viewer's lineCells and shapeCells mirror LinePoints and
EllipsePoints in canvas/shapes.go. Change one and change the other, or
a shape the person drags and the same shape the model draws stop
agreeing.
