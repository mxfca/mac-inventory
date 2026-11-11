// Package store writes the inventory file into the destination folder.
package store

import (
	"os"
	"path/filepath"
)

// FileName is the inventory file inside the destination folder.
const FileName = "inventario-mac.md"

// Write replaces the inventory in dir with content, all at once: a reader sees
// either the old file or the new one, never half of it. It does not create dir.
func Write(dir string, content []byte) (err error) {
	tmp, err := os.CreateTemp(dir, ".inventario-mac-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if _, err = tmp.Write(content); err != nil {
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
	return os.Rename(tmp.Name(), filepath.Join(dir, FileName))
}
