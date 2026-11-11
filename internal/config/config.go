// Package config reads the destination folder from ~/.mac-inventory.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileName is the config file inside the home folder.
const FileName = ".mac-inventory"

// Read returns the destination folder written in home/.mac-inventory.
// A missing or empty file, or a relative path, is a configuration error.
func Read(home string) (string, error) {
	path := filepath.Join(home, FileName)
	content, err := os.ReadFile(path) // #nosec G304 -- fixed file name in the home folder
	if err != nil {
		return "", fmt.Errorf("destination not configured: %w", err)
	}
	dir := strings.TrimSpace(string(content))
	if dir == "" {
		return "", fmt.Errorf("destination not configured: %s is empty", path)
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("destination not configured: %q in %s is not an absolute path", dir, path)
	}
	return dir, nil
}
