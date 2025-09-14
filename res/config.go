package res

import (
	"image/color"
	"log"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"

	"glyphrenderer/fs"
)

type Config struct {
	LogicalW int
	LogicalH int
	Scale    int

	ClearColor []int
}

var (
	configCache = map[string]*Config{}
	configMu    sync.Mutex
)

func (c *Config) ClearColorRGBA() color.Color {
	if len(c.ClearColor) != 4 {
		// Fallback to opaque black if not defined properly
		return color.RGBA{0, 0, 0, 255}
	}
	return color.RGBA{
		R: uint8(c.ClearColor[0]),
		G: uint8(c.ClearColor[1]),
		B: uint8(c.ClearColor[2]),
		A: uint8(c.ClearColor[3]),
	}
}

// LoadConfig loads and caches a TOML config from a relative path,
// e.g. "game/pong/config.toml".
func LoadConfig(relPath string) *Config {
	configMu.Lock()
	defer configMu.Unlock()

	if cfg, ok := configCache[relPath]; ok {
		return cfg
	}

	base := fs.ProjectRoot()
	path := filepath.Join(base, relPath)

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		log.Fatalf("failed to load config %q: %v", path, err)
	}

	configCache[relPath] = &cfg
	return &cfg
}
