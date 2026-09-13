# pixelduet

A pixel canvas a model and a person share. Claude draws it stroke by
stroke while you watch in the browser, and at any moment you can pause
it, take the brush, change things, leave a note, and hand the brush
back.

This package downloads the prebuilt binary for your machine and runs
it. Node 18 or newer is required.

    npx -y pixelduet                # canvas server + browser viewer, opens the viewer

Let Claude Code in, once:

    claude mcp add pixelduet -- npx -y pixelduet bridge

If nothing is running when a session starts, the bridge starts the
servers itself, so that one line is the whole setup.

Source and docs are in the repository.
