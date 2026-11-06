package web

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/hub"
)

type webResponse struct {
	header http.Header
	Code   int
	Body   bytes.Buffer
}

func (r *webResponse) Header() http.Header         { return r.header }
func (r *webResponse) WriteHeader(code int)        { r.Code = code }
func (r *webResponse) Write(p []byte) (int, error) { return r.Body.Write(p) }

func TestWebSaveFolder(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go hub.New(8, time.Second).Serve(ln)
	defaultFolder := t.TempDir()
	customFolder := filepath.Join(t.TempDir(), "custom folder", "pixels")
	handler := Handler(ln.Addr().String(), defaultFolder)
	t.Run("snapshot after command", func(t *testing.T) {
		command, _ := http.NewRequest("POST", "http://localhost/cmd", strings.NewReader("set 2 3 11"))
		command.Header.Set("Origin", "http://localhost")
		written := &webResponse{header: make(http.Header), Code: 200}
		handler.ServeHTTP(written, command)
		if written.Code != 200 {
			t.Fatalf("command: %d %s", written.Code, written.Body.String())
		}
		req, _ := http.NewRequest("GET", "http://localhost/snapshot", nil)
		res := &webResponse{header: make(http.Header), Code: 200}
		handler.ServeHTTP(res, req)
		var state struct {
			W, H int
			Grid string
		}
		if err := json.Unmarshal(res.Body.Bytes(), &state); err != nil {
			t.Fatal(err)
		}
		canvas, err := canvas.ParseGrid(state.Grid)
		if res.Code != 200 || state.W != 8 || state.H != 8 || err != nil || canvas.At(2, 3) != 11 {
			t.Fatalf("stale snapshot: %d %+v %v", res.Code, state, err)
		}
		if res.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("snapshot must not be cached")
		}
	})
	for _, tc := range []struct{ name, folder, contentType, body string }{
		{"custom", customFolder, "application/json", ""},
		{"default", defaultFolder, "application/json", `{"name":"default"}`},
		{"legacy", defaultFolder, "text/plain", "legacy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			if body == "" {
				data, _ := json.Marshal(map[string]string{"name": tc.name, "folder": tc.folder})
				body = string(data)
			}
			req, _ := http.NewRequest("POST", "http://localhost/save", strings.NewReader(body))
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Origin", "http://localhost")
			res := &webResponse{header: make(http.Header), Code: 200}
			handler.ServeHTTP(res, req)
			if res.Code != 200 {
				t.Fatalf("save: %d %s", res.Code, res.Body.String())
			}
			data, err := os.ReadFile(filepath.Join(tc.folder, tc.name+".txt"))
			if err != nil {
				t.Fatal(err)
			}
			canvas, err := canvas.ParseGrid(string(data))
			if err != nil || canvas.W != 8 || canvas.H != 8 {
				t.Fatalf("saved canvas: %v, %v", canvas, err)
			}
		})
	}
}

func TestWebRejectsInvalidSavesAndCrossOriginWrites(t *testing.T) {
	handler := Handler("127.0.0.1:0", t.TempDir())
	for _, tc := range []struct {
		path, body, origin string
		want               int
	}{
		{"/save", `{"name":"../escape"}`, "http://localhost", 400},
		{"/save", `{broken`, "http://localhost", 400},
		{"/save", strings.Repeat("x", 8193), "http://localhost", 400},
		{"/save", `{"name":"test"}`, "https://other.example", 403},
		{"/cmd", "clear", "https://other.example", 403},
	} {
		req, _ := http.NewRequest("POST", "http://localhost"+tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", tc.origin)
		res := &webResponse{header: make(http.Header), Code: 200}
		handler.ServeHTTP(res, req)
		if res.Code != tc.want {
			t.Errorf("%s origin %s: got %d, want %d", tc.path, tc.origin, res.Code, tc.want)
		}
	}
}
