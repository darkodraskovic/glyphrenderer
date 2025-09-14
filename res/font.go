package res

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"glyphrenderer/fs"
)

var (
	fontCache = map[string]*text.GoTextFaceSource{}
	fontMu    sync.Mutex
)

// LoadFontSource loads a TTF font file into a cached GoTextFaceSource.
// Example: src := res.LoadFontSource("press-start-2p.ttf")
func LoadFontSource(name string) *text.GoTextFaceSource {
	fontMu.Lock()
	defer fontMu.Unlock()

	if src, ok := fontCache[name]; ok {
		return src
	}

	base := fs.ProjectRoot()
	path := filepath.Join(base, "assets", "fonts", name)

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open font %q: %v", path, err)
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		log.Fatalf("failed to parse font %q: %v", path, err)
	}

	fontCache[name] = src
	return src
}

// LoadFace loads a font by filename and size, returning a ready-to-use GoTextFace.
func LoadFace(filename string, size float64) *text.GoTextFace {
	src := LoadFontSource(filename)
	return &text.GoTextFace{
		Source: src,
		Size:   size,
	}
}
