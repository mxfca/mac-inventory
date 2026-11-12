// Package config reads and writes the destination folder in ~/.mac-inventory.
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

// Write saves dir, with the spaces at its ends trimmed, as the destination folder in home/.mac-inventory.
// A relative path or a folder that does not exist is refused and the file is left untouched.
func Write(home, dir string) error {
	dir = strings.TrimSpace(dir)
	if err := validateFolder(dir); err != nil {
		return err
	}
	path := filepath.Join(home, FileName)
	if err := os.WriteFile(path, []byte(dir+"\n"), 0o600); err != nil {
		return err
	}
	// WriteFile only applies the mode when it creates the file.
	return os.Chmod(path, 0o600)
}

// validateFolder checks that dir is an absolute path to an existing folder.
func validateFolder(dir string) error {
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("%q is not an absolute path", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("destination folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a folder", dir)
	}
	return nil
}
