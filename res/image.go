package res

import (
	"image"
	"log"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"glyphrenderer/fs"
)

var (
	imageCache = map[string]*ebiten.Image{}
	imageMu    sync.Mutex
)

// LoadImage loads an image from assets/images/, caching it in memory.
// Example: img := assets.LoadImage("mop.png")
func LoadImage(name string) *ebiten.Image {
	imageMu.Lock()
	defer imageMu.Unlock()

	// Return cached image if already loaded
	if img, ok := imageCache[name]; ok {
		return img
	}

	// Build path relative to project root
	base := fs.ProjectRoot()
	path := filepath.Join(base, "assets", "images", name)

	// Load with ebitenutil (simple dev loader)
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Fatalf("failed to load image %q: %v", path, err)
	}

	imageCache[name] = img
	return img
}

// LoadSubImage loads a sub-region of an image from assets/images/.
// It uses the cached base image and returns a view into the requested rect.
func LoadSubImage(name string, rect image.Rectangle) *ebiten.Image {
	base := LoadImage(name) // whole image (cached)

	// Ensure rect is within bounds
	bounds := base.Bounds()
	if !rect.In(bounds) {
		log.Fatalf("subimage rect %v out of bounds %v for %q", rect, bounds, name)
	}

	return base.SubImage(rect).(*ebiten.Image)
}
