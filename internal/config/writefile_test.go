package config

import (
	"os"
	"path/filepath"
	"testing"
)

func entries(t *testing.T, home string) []string {
	t.Helper()
	list, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range list {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFileCreatesFileWith0600AndLineBreak(t *testing.T) {
	home := t.TempDir()
	if err := writeFile(home, "/some/dir"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, FileName)
	got, err := os.ReadFile(path) // #nosec G304 -- test file in a temp dir
	if err != nil || string(got) != "/some/dir\n" {
		t.Fatalf("content = %q, err %v", got, err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
	if names := entries(t, home); len(names) != 1 {
		t.Errorf("home has leftovers: %v", names)
	}
}

func TestWriteFileTightensLooseExistingMode(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, FileName)
	if err := os.WriteFile(path, []byte("/old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil { // #nosec G302 -- loose mode on purpose
		t.Fatal(err)
	}
	if err := writeFile(home, "/new"); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestWriteFileFailureKeepsOldFileAndLeavesNoTemp(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, FileName)
	// A folder in place of the config makes the final rename fail after the temp was written.
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(home, "/new"); err == nil {
		t.Fatal("expected error")
	}
	if names := entries(t, home); len(names) != 1 {
		t.Errorf("temp file left behind: %v", names)
	}
}

func TestWriteFileFailsOnMissingHome(t *testing.T) {
	if err := writeFile(filepath.Join(t.TempDir(), "missing"), "/new"); err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteFileOnReadOnlyHomeKeepsOldContent(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, FileName)
	if err := os.WriteFile(path, []byte("/old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0o500); err != nil { // #nosec G302 -- read-only home on purpose
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(home, 0o700) }) // #nosec G302 -- lets TempDir clean up
	if err := writeFile(home, "/new"); err == nil {
		t.Fatal("expected error")
	}
	got, err := os.ReadFile(path) // #nosec G304 -- test file in a temp dir
	if err != nil || string(got) != "/old\n" {
		t.Errorf("old content = %q, err %v", got, err)
	}
}
