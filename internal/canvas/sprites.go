package canvas

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	DefaultSprites = "sprites"
	spriteExt      = ".txt"
)

var spriteNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

type SpriteInfo struct {
	Name string
	W, H int
}

// SpritePath rejects names that could escape dir.
func SpritePath(dir, name string) (string, error) {
	if !spriteNameRe.MatchString(name) {
		return "", fmt.Errorf("sprite name %q: use letters, digits, - and _ only", name)
	}
	return filepath.Join(dir, name+spriteExt), nil
}

func AbsPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// ListSprites treats a missing folder as empty.
func ListSprites(dir string) ([]SpriteInfo, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []SpriteInfo
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), spriteExt)
		if e.IsDir() || !ok || !spriteNameRe.MatchString(name) {
			continue
		}
		c, err := LoadSprite(dir, name)
		if err != nil {
			continue
		}
		out = append(out, SpriteInfo{name, c.W, c.H})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func LoadSprite(dir, name string) (*Canvas, error) {
	path, err := SpritePath(dir, name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no sprite named %q", name)
	}
	return ParseGrid(strings.ReplaceAll(string(data), "\r\n", "\n"))
}

// SaveSprite writes dir/name.txt and returns its absolute path.
func SaveSprite(dir, name string, c *Canvas) (string, error) {
	path, err := SpritePath(dir, name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(c.String()), 0o644); err != nil {
		return "", err
	}
	return AbsPath(path), nil
}

// StampStrokes places s at (x,y); empty pixels are skipped.
func StampStrokes(s *Canvas, x, y int) []Stroke {
	var set strokeSet
	for yy := 0; yy < s.H; yy++ {
		for xx := 0; xx < s.W; xx++ {
			if col := s.At(xx, yy); col != 0 {
				set.add(col, x+xx, y+yy)
			}
		}
	}
	return set.strokes()
}
