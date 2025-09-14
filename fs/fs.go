package fs

import (
	"os"
	"path/filepath"
)

// ProjectRoot returns the root directory of the project
// by walking up one level from the executable.
func ProjectRoot() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return filepath.Join(exPath, "../..")
}
