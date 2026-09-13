package local

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

func TestStartLocalServesTheCanvasAndTheViewer(t *testing.T) {
	l, err := Start("127.0.0.1:0", "127.0.0.1:0", 8, time.Second, t.TempDir())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res, err := http.Get(l.URL() + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || !strings.Contains(string(body), "pixelduet") {
		t.Fatalf("viewer answered %d with %d bytes", res.StatusCode, len(body))
	}

	conn, err := net.Dial("tcp", l.Addr)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	defer conn.Close()
	c, _, err := proto.AwaitState(bufio.NewReader(conn))
	if err != nil || c.W != 8 || c.H != 8 {
		t.Fatalf("hub greeted with %v, %v", c, err)
	}
}

func TestEnsureServerStartsOneOnlyWhenNoneAnswers(t *testing.T) {
	free := func() string {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		return ln.Addr().String()
	}
	addr := free()
	l, err := Ensure(addr, "127.0.0.1:0", t.TempDir())
	if err != nil || l == nil {
		t.Fatalf("first call = %v, %v; want a started server", l, err)
	}
	again, err := Ensure(addr, "127.0.0.1:0", t.TempDir())
	if err != nil || again != nil {
		t.Fatalf("second call = %v, %v; want nil for a server already running", again, err)
	}
}

func TestLocalURLNamesAHostABrowserCanReach(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"[::]:8080", "http://localhost:8080"},
		{"0.0.0.0:8080", "http://localhost:8080"},
	} {
		if got := (&Servers{HTTPAddr: tc.in}).URL(); got != tc.want {
			t.Fatalf("url of %q = %q, want %q", tc.in, got, tc.want)
		}
	}
}
