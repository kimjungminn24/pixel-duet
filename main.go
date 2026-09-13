package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kimjungminn24/pixel-duet/internal/bridge"
	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/hub"
	"github.com/kimjungminn24/pixel-duet/internal/local"
	"github.com/kimjungminn24/pixel-duet/internal/web"
)

const usage = `pixelduet - a pixel canvas a model and a person share

  pixelduet                  start the canvas server and browser viewer
  pixelduet bridge           run the tool bridge over stdio, for Claude Code,
                             Codex, or any MCP client; starts the servers
                             itself if none is running
  pixelduet serve            the canvas server on its own
  pixelduet web              the browser viewer on its own
  pixelduet rules            print the drawing rules a model is handed
  pixelduet version          print the version
`

// version is set at build time: -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	// A leading flag means start: `pixelduet -no-browser`.
	cmd, args := "start", os.Args[1:]
	switch {
	case len(args) == 0:
	case args[0] == "-h" || args[0] == "--help" || args[0] == "help":
		fmt.Print(usage)
		return
	case !strings.HasPrefix(args[0], "-"):
		cmd, args = args[0], args[1:]
	}

	var err error
	switch cmd {
	case "start":
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		addr := fs.String("addr", hub.DefaultAddr, "address for the canvas server")
		httpAddr := fs.String("http", local.DefaultHTTP, "address for the browser viewer")
		size := fs.Int("size", local.DefaultSize, "canvas width and height in pixels")
		wait := fs.Duration("max-wait", local.DefaultMaxWait, "how long a paused drawing command waits before giving up")
		sprites := fs.String("sprites", canvas.DefaultSprites, "folder for saved sprites")
		noBrowser := fs.Bool("no-browser", false, "do not open the viewer in a browser")
		fs.Parse(args)
		err = local.Run(*addr, *httpAddr, *size, *wait, *sprites, !*noBrowser, version)
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		addr := fs.String("addr", hub.DefaultAddr, "address to listen on")
		size := fs.Int("size", local.DefaultSize, "canvas width and height in pixels")
		wait := fs.Duration("max-wait", local.DefaultMaxWait, "how long a paused drawing command waits before giving up")
		fs.Parse(args)
		err = hub.Run(*addr, *size, *wait)
	case "web":
		fs := flag.NewFlagSet("web", flag.ExitOnError)
		addr := fs.String("addr", hub.DefaultAddr, "pixelduet server address")
		httpAddr := fs.String("http", local.DefaultHTTP, "address to serve the browser viewer on")
		sprites := fs.String("sprites", canvas.DefaultSprites, "folder the save button writes sprites to")
		fs.Parse(args)
		err = web.Run(*addr, *httpAddr, *sprites)
	case "bridge":
		fs := flag.NewFlagSet("bridge", flag.ExitOnError)
		addr := fs.String("addr", hub.DefaultAddr, "pixelduet server address")
		httpAddr := fs.String("http", local.DefaultHTTP, "address for the browser viewer, if this has to start the servers")
		sprites := fs.String("sprites", canvas.DefaultSprites, "folder of saved sprites the model can look at")
		fs.Parse(args)
		err = runBridge(*addr, *httpAddr, *sprites)
	case "rules":
		fmt.Println(bridge.Rules)
	case "version":
		fmt.Println(version)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "pixelduet:", err)
		os.Exit(1)
	}
}

// runBridge starts the servers in this process when none answer, then serves the tools.
func runBridge(addr, httpAddr, sprites string) error {
	viewer := ""
	if l, err := local.Ensure(addr, httpAddr, sprites); err != nil {
		return err
	} else if l != nil {
		viewer = l.URL()
	}
	return bridge.Run(addr, sprites, viewer, version)
}
