package res

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"glyphrenderer/fs"
)

var (
	shaderCache = map[string]*ebiten.Shader{}
	shaderMu    sync.Mutex
)

// LoadShader loads and compiles a shader from assets/shaders/, caching it in memory.
// Example: sh := res.LoadShader("glyph.kage")
func LoadShader(name string) *ebiten.Shader {
	shaderMu.Lock()
	defer shaderMu.Unlock()

	// Return cached shader if already loaded
	if sh, ok := shaderCache[name]; ok {
		return sh
	}

	base := fs.ProjectRoot()
	path := filepath.Join(base, "assets", "shaders", name)

	// Read shader source
	src, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read shader %q: %v", path, err)
	}

	// Compile shader
	sh, err := ebiten.NewShader(src)
	if err != nil {
		log.Fatalf("failed to compile shader %q: %v", path, err)
	}

	shaderCache[name] = sh
	return sh
}
