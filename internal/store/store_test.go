package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mxfca/mac-inventory/internal/store"
)

var at = time.Date(2026, 11, 15, 12, 0, 3, 0, time.Local)

func TestWriteCreatesInventory(t *testing.T) {
	dir := t.TempDir()

	if err := store.Write(dir, []byte("novo"), true, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertOnlyInventory(t, dir, "novo")
}

func TestWriteOverwritesInventory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inventario-mac.md"), []byte("antigo, mais longo"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("novo"), true, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertOnlyInventory(t, dir, "novo")
}

func TestWriteFailsWithoutCreatingMissingFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "OneDrive")

	if err := store.Write(dir, []byte("novo"), true, at); err == nil {
		t.Fatal("got nil error for a missing folder")
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("folder was created: %v", err)
	}
}

func TestWriteLeavesNoTemporaryFileWhenRenameFails(t *testing.T) {
	dir := t.TempDir()
	// A non-empty folder named like the inventory makes the rename fail.
	if err := os.MkdirAll(filepath.Join(dir, "inventario-mac.md", "x"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("novo"), true, at); err == nil {
		t.Fatal("got nil error when the rename fails")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("folder has %v, want no temporary file left", entries)
	}
}

func assertOnlyInventory(t *testing.T, dir, want string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "inventario-mac.md" {
		t.Fatalf("folder has %v, want only inventario-mac.md", entries)
	}
	got, err := os.ReadFile(filepath.Join(dir, "inventario-mac.md")) // #nosec G304 -- test temp folder
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("inventory is %q, want %q", got, want)
	}
}

func TestWritePartialBacksUpPreviousInventory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "inventario-mac.md", "anterior")

	if err := store.Write(dir, []byte("parcial"), false, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertFile(t, dir, "inventario-mac.md", "parcial")
	assertFile(t, dir, "inventario-mac-2026-11-15-120003.md", "anterior")
}

func TestWritePartialWithoutPreviousInventoryMakesNoBackup(t *testing.T) {
	dir := t.TempDir()

	if err := store.Write(dir, []byte("parcial"), false, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertOnlyInventory(t, dir, "parcial")
}

func TestWriteTwoPartialsKeepTwoBackups(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "inventario-mac.md", "completa")

	if err := store.Write(dir, []byte("parcial 1"), false, at); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := store.Write(dir, []byte("parcial 2"), false, at.Add(time.Hour)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertFile(t, dir, "inventario-mac.md", "parcial 2")
	assertFile(t, dir, "inventario-mac-2026-11-15-120003.md", "completa")
	assertFile(t, dir, "inventario-mac-2026-11-15-130003.md", "parcial 1")
}

func TestWriteCompleteDeletesBackupsAndKeepsOtherFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "inventario-mac.md", "parcial")
	writeFile(t, dir, "inventario-mac-2026-11-08-120003.md", "backup")
	writeFile(t, dir, "notas.md", "minhas notas")

	if err := store.Write(dir, []byte("completa"), true, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertFile(t, dir, "inventario-mac.md", "completa")
	assertFile(t, dir, "notas.md", "minhas notas")
	assertEntries(t, dir, 2)
}

func TestWriteFailsAndKeepsInventoryWhenBackupFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "inventario-mac.md", "anterior")
	// A folder named like the backup makes the backup fail.
	if err := os.Mkdir(filepath.Join(dir, "inventario-mac-2026-11-15-120003.md"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("parcial"), false, at); err == nil {
		t.Fatal("got nil error when the backup fails")
	}

	assertFile(t, dir, "inventario-mac.md", "anterior")
}

func TestWriteCompleteKeepsBackupsWhenWriteFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "inventario-mac-2026-11-08-120003.md", "backup")
	if err := os.MkdirAll(filepath.Join(dir, "inventario-mac.md", "x"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("completa"), true, at); err == nil {
		t.Fatal("got nil error when the write fails")
	}

	assertFile(t, dir, "inventario-mac-2026-11-08-120003.md", "backup")
}

func TestWriteCompleteSucceedsWhenBackupCannotBeDeleted(t *testing.T) {
	dir := t.TempDir()
	// A non-empty folder named like a backup cannot be removed.
	if err := os.MkdirAll(filepath.Join(dir, "inventario-mac-2026-11-08-120003.md", "x"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("completa"), true, at); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertFile(t, dir, "inventario-mac.md", "completa")
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, dir, name, want string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(dir, name)) // #nosec G304 -- test temp folder
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s is %q, want %q", name, got, want)
	}
}

func assertEntries(t *testing.T, dir string, want int) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != want {
		t.Errorf("folder has %v, want %d entries", entries, want)
	}
}
