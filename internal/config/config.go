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
	return writeFile(home, dir)
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

// writeFile atomically replaces home/.mac-inventory with dir and a line break:
// readers see the old file or the new one, never a partial one.
// The temp file is created in home so the rename stays on one filesystem, and it is born 0600.
func writeFile(home, dir string) (err error) {
	tmp, err := os.CreateTemp(home, FileName+"-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if _, err = tmp.WriteString(dir + "\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(home, FileName))
}
