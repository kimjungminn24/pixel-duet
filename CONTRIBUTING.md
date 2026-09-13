# Contributing

Thanks for looking. pixelduet is small on purpose, and the easiest way
to keep it that way is to know where things go before adding to them.

## Before you send a change

    gofmt -l .                          # prints nothing
    go vet ./...
    go test ./...
    node --test test/*.test.mjs         # the viewer: shape math, translations

CI runs the same four steps on every pull request.

## Where things live

CLAUDE.md is the map of the code and the house style for comments. It
is short; read it first. The wire protocol is documented at the top of
internal/proto/proto.go, and a new command starts there.

The browser viewer is plain HTML, CSS and ES modules under
internal/web/static/, with no build step. It is embedded into the binary, so `go run .` serves
whatever is on disk.

The drawing rules a model receives are in internal/bridge/rules.go. They are prose, and
changing them is the most effective way to change how the model draws.

## Reporting a bug

Open an issue with the canvas size, the tool or command involved, and
what the log panel said. `pixelduet version` prints the build you are
on. If the model drew something wrong, the digit grid from read_pixels
is more useful than a screenshot.

## Releasing

    git tag v0.2.0
    git push origin v0.2.0

The release workflow runs the tests, builds binaries for six platforms,
creates the GitHub Release, and publishes the npm package through
Trusted Publishing. No secrets are involved: npmjs.com is configured
to accept release.yml from this repository as the package's publisher.
