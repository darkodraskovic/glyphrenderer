package fs

import (
	"log"
	"os"
	"path/filepath"
)

// ProjectRoot returns the root directory of the project
// by walking up one level from the executable.
func ProjectRoot() string {
	wd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			log.Fatal("project root not found")
		}
		wd = parent
	}
}
