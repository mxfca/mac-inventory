// Package store writes the inventory file into the destination folder.
package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// FileName is the inventory file inside the destination folder.
const FileName = "inventario-mac.md"

const (
	backupPrefix = "inventario-mac-"
	backupSuffix = ".md"
	backupLayout = "2006-01-02-150405"
)

// Write replaces the inventory in dir with content, all at once: a reader sees
// either the old file or the new one, never half of it. It does not create dir.
//
// A partial collection first copies the previous inventory to a backup named
// after at, and fails without touching the inventory if the copy fails. A
// complete collection deletes every backup once the inventory is written.
func Write(dir string, content []byte, complete bool, at time.Time) error {
	if !complete {
		if err := backUp(dir, at); err != nil {
			return err
		}
	}
	if err := replace(dir, content); err != nil {
		return err
	}
	if complete {
		deleteBackups(dir)
	}
	return nil
}

// backUp copies the current inventory to a backup; no inventory means nothing to copy.
func backUp(dir string, at time.Time) error {
	previous, err := os.ReadFile(filepath.Join(dir, FileName)) // #nosec G304 -- dir comes from the config file
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	name := backupPrefix + at.Format(backupLayout) + backupSuffix
	return os.WriteFile(filepath.Join(dir, name), previous, 0o600) // #nosec G703 -- dir comes from the config file
}

// deleteBackups is best effort: the inventory is already written, and a backup
// left behind goes away on the next complete collection.
func deleteBackups(dir string) {
	backups, _ := filepath.Glob(filepath.Join(dir, backupPrefix+"*"+backupSuffix))
	for _, backup := range backups {
		_ = os.Remove(backup)
	}
}

func replace(dir string, content []byte) (err error) {
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
