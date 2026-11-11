package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mxfca/mac-inventory/internal/store"
)

func TestWriteCreatesInventory(t *testing.T) {
	dir := t.TempDir()

	if err := store.Write(dir, []byte("novo")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertOnlyInventory(t, dir, "novo")
}

func TestWriteOverwritesInventory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "inventario-mac.md"), []byte("antigo, mais longo"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(dir, []byte("novo")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	assertOnlyInventory(t, dir, "novo")
}

func TestWriteFailsWithoutCreatingMissingFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "OneDrive")

	if err := store.Write(dir, []byte("novo")); err == nil {
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

	if err := store.Write(dir, []byte("novo")); err == nil {
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
