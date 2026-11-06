package web

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kimjungminn24/pixel-duet/internal/canvas"
	"github.com/kimjungminn24/pixel-duet/internal/proto"
)

//go:embed static
var webFiles embed.FS

const cmdTimeout = 10 * time.Second

const saveBodyMax = 8192

// cmdBodyMax: the page splits long strokes to stay under it.
const cmdBodyMax = 4096

func init() {
	// ES modules need a JavaScript type; the Windows registry can say text/plain for .js.
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
}

func Run(addr, httpAddr, sprites string) error {
	log.Printf("browser viewer on http://%s (canvas server %s)", httpAddr, addr)
	return http.ListenAndServe(httpAddr, Handler(addr, sprites))
}

func Handler(addr, sprites string) http.Handler {
	pages, err := fs.Sub(webFiles, "static")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(pages))
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		serveEvents(w, r, addr)
	})
	mux.HandleFunc("GET /snapshot", func(w http.ResponseWriter, r *http.Request) {
		serveSnapshot(w, addr)
	})
	mux.HandleFunc("POST /cmd", func(w http.ResponseWriter, r *http.Request) {
		serveCmd(w, r, addr)
	})
	mux.HandleFunc("POST /save", func(w http.ResponseWriter, r *http.Request) {
		serveSave(w, r, addr, sprites)
	})
	mux.HandleFunc("GET /gallery", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, canvas.AbsPath(sprites))
	})
	return http.NewCrossOriginProtection().Handler(mux)
}

// dialHub opens a one-shot connection with a deadline; view commands are never held.
func dialHub(addr string) (net.Conn, *bufio.Reader, error) {
	conn, err := net.DialTimeout("tcp", addr, cmdTimeout)
	if err != nil {
		return nil, nil, err
	}
	conn.SetDeadline(time.Now().Add(cmdTimeout))
	return conn, bufio.NewReader(conn), nil
}

type (
	stateEvent struct {
		Type   string `json:"type"`
		W      int    `json:"w"`
		H      int    `json:"h"`
		Paused bool   `json:"paused"`
		Grid   string `json:"grid"`
	}
	logEvent struct {
		Type    string       `json:"type"`
		Entries []logMessage `json:"entries"`
	}
	logMessage struct {
		Who  string `json:"who"`
		Text string `json:"text"`
	}
	paletteEvent struct {
		Type   string   `json:"type"`
		Colors []string `json:"colors"` // "#rrggbb" per slot; slot 0 empty
	}
	speedEvent struct {
		Type string `json:"type"`
		MS   int64  `json:"ms"`
	}
)

func eventOf(m proto.Message) any {
	switch m.Kind {
	case proto.MsgState:
		return stateEvent{"state", m.Canvas.W, m.Canvas.H, m.Paused, m.Canvas.String()}
	case proto.MsgLog:
		entries := make([]logMessage, len(m.Entries))
		for i, e := range m.Entries {
			entries[i] = logMessage{e.Who, e.Text}
		}
		return logEvent{"log", entries}
	case proto.MsgPalette:
		colors := make([]string, len(m.Palette))
		for i := 1; i < len(m.Palette); i++ {
			colors[i] = "#" + m.Palette[i].Hex()
		}
		return paletteEvent{"palette", colors}
	case proto.MsgSpeed:
		return speedEvent{"speed", m.Delay.Milliseconds()}
	}
	return nil
}

// serveEvents holds one SSE stream, backed by one hub connection, per tab.
func serveEvents(w http.ResponseWriter, r *http.Request, addr string) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "hello view\n")
	go func() { <-r.Context().Done(); conn.Close() }()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	br := bufio.NewReader(conn)
	for {
		m, err := proto.ReadMessage(br)
		if err != nil {
			return
		}
		ev := eventOf(m)
		if ev == nil {
			continue
		}
		data, err := json.Marshal(ev)
		if err != nil {
			return
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return
		}
		fl.Flush()
	}
}

func serveSnapshot(w http.ResponseWriter, addr string) {
	conn, br, err := dialHub(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()
	c, _, err := proto.AwaitState(br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(struct {
		W    int    `json:"w"`
		H    int    `json:"h"`
		Grid string `json:"grid"`
	}{c.W, c.H, c.String()})
}

type saveRequest struct {
	Name   string `json:"name"`
	Folder string `json:"folder"` // empty for the server's gallery
}

// readSaveRequest accepts JSON, or a plain-text body holding only the name.
func readSaveRequest(r *http.Request, body []byte) (saveRequest, error) {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/json" {
		return saveRequest{Name: strings.TrimSpace(string(body))}, nil
	}
	var req saveRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("invalid save request")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Folder = strings.TrimSpace(req.Folder)
	return req, nil
}

func serveSave(w http.ResponseWriter, r *http.Request, addr, sprites string) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, saveBodyMax))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req, err := readSaveRequest(r, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Folder != "" {
		sprites = req.Folder
	}
	if _, err := canvas.SpritePath(sprites, req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, br, err := dialHub(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()
	c, _, err := proto.AwaitState(br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	path, err := canvas.SaveSprite(sprites, req.Name, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%dx%d saved to %s", c.W, c.H, path)
}

// serveCmd forwards one line as a view client.
func serveCmd(w http.ResponseWriter, r *http.Request, addr string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, cmdBodyMax))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	line := strings.TrimSpace(string(body))
	if line == "" {
		http.Error(w, "empty command", http.StatusBadRequest)
		return
	}
	conn, br, err := dialHub(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "hello view\n%s\n", line)
	if _, err := proto.AwaitResult(br); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	res, err := proto.AwaitResult(br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	io.WriteString(w, res)
}
