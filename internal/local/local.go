package local

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/hub"
	"github.com/kimjungminn24/pixel-duet/internal/web"
)

const (
	DefaultHTTP    = "127.0.0.1:8080"
	DefaultSize    = 32
	DefaultMaxWait = 100 * time.Second
)

// Servers is a hub and a viewer running in this process.
type Servers struct {
	Addr     string
	HTTPAddr string
	errs     chan error
}

func Start(addr, httpAddr string, size int, maxWait time.Duration, sprites string) (*Servers, error) {
	hubLn, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("canvas server: %w", err)
	}
	webLn, err := net.Listen("tcp", httpAddr)
	if err != nil {
		hubLn.Close()
		return nil, fmt.Errorf("browser viewer: %w", err)
	}
	l := &Servers{Addr: hubLn.Addr().String(), HTTPAddr: webLn.Addr().String(), errs: make(chan error, 2)}
	go func() { l.errs <- hub.New(size, maxWait).Serve(hubLn) }()
	go func() { l.errs <- http.Serve(webLn, web.Handler(l.Addr, sprites)) }()
	return l, nil
}

// url replaces an unspecified host ([::], 0.0.0.0) with localhost.
func (l *Servers) URL() string {
	host, port, err := net.SplitHostPort(l.HTTPAddr)
	if err != nil {
		return "http://" + l.HTTPAddr
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func Run(addr, httpAddr string, size int, maxWait time.Duration, sprites string, browser bool, version string) error {
	l, err := Start(addr, httpAddr, size, maxWait, sprites)
	if err != nil {
		return err
	}
	fmt.Printf("pixelduet %s\n\n", version)
	fmt.Printf("  canvas server   %s\n", l.Addr)
	fmt.Printf("  browser viewer  %s\n", l.URL())
	fmt.Printf("  let Claude in   claude mcp add pixelduet -- npx -y pixelduet bridge\n")
	fmt.Printf("  stop            ctrl-c\n")
	if browser {
		openBrowser(l.URL())
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case <-stop:
		return nil
	case err := <-l.errs:
		return err
	}
}

// Ensure starts a hub and viewer when nothing answers at addr; nil means one was already running.
func Ensure(addr, httpAddr, sprites string) (*Servers, error) {
	if c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond); err == nil {
		c.Close()
		return nil, nil
	}
	l, err := Start(addr, httpAddr, DefaultSize, DefaultMaxWait, sprites)
	if err != nil {
		return nil, err
	}
	log.Printf("no canvas server at %s; started one, browser viewer at %s", addr, l.URL())
	return l, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
