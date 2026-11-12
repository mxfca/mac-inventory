package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mxfca/mac-inventory/internal/config"
)

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, config.FileName), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestReadTrimsSpacesAndLineBreaks(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "  /Users/maxx/OneDrive/Inventory \n")

	got, err := config.Read(home)

	if err != nil || got != "/Users/maxx/OneDrive/Inventory" {
		t.Errorf("got %q, %v; want the trimmed path", got, err)
	}
}

func TestReadFailsWithoutFile(t *testing.T) {
	if got, err := config.Read(t.TempDir()); err == nil {
		t.Errorf("got %q, want a configuration error", got)
	}
}

func TestReadFailsOnBadContent(t *testing.T) {
	cases := map[string]string{
		"empty file":    " \n",
		"relative path": "OneDrive/Inventory\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			writeConfig(t, home, content)

			if got, err := config.Read(home); err == nil {
				t.Errorf("got %q, want a configuration error", got)
			}
		})
	}
}

func readFile(t *testing.T, home string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(home, config.FileName)) // #nosec G304 -- temp home in a test
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestWriteCreatesFileReadableByRead(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()

	if err := config.Write(home, dir); err != nil {
		t.Fatal(err)
	}

	if got, err := config.Read(home); err != nil || got != dir {
		t.Errorf("Read = %q, %v; want %q", got, err, dir)
	}
}

func TestWriteReplacesExistingFile(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	writeConfig(t, home, "/old/folder\n")

	if err := config.Write(home, dir); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, home); got != dir+"\n" {
		t.Errorf("file = %q, want %q", got, dir+"\n")
	}
}

func TestWriteUsesPermission0600(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "/old/folder\n")
	if err := os.Chmod(filepath.Join(home, config.FileName), 0o644); err != nil { // #nosec G302 -- the test needs a loose mode to prove Write tightens it
		t.Fatal(err)
	}

	if err := config.Write(home, t.TempDir()); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(home, config.FileName))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, %v; want 0600", info.Mode().Perm(), err)
	}
}

func TestWriteRefusesBadFolderAndKeepsFile(t *testing.T) {
	notAFolder := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(notAFolder, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"relative path":  "OneDrive/Inventory",
		"missing folder": filepath.Join(t.TempDir(), "missing"),
		"not a folder":   notAFolder,
	}
	for name, dir := range cases {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			writeConfig(t, home, "/old/folder\n")

			if err := config.Write(home, dir); err == nil {
				t.Fatal("want an error")
			}

			if got := readFile(t, home); got != "/old/folder\n" {
				t.Errorf("file changed to %q", got)
			}
		})
	}
}

func TestWriteRefusesBadFolderWithoutCreatingFile(t *testing.T) {
	home := t.TempDir()

	if err := config.Write(home, "relative"); err == nil {
		t.Fatal("want an error")
	}

	if _, err := os.Stat(filepath.Join(home, config.FileName)); !os.IsNotExist(err) {
		t.Errorf("file should not exist, stat err = %v", err)
	}
}

func TestWriteTrimsSpacesAndReadReturnsThemTrimmed(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()

	if err := config.Write(home, "  "+dir+" \t"); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, home); got != dir+"\n" {
		t.Errorf("file = %q, want %q", got, dir+"\n")
	}
	if got, err := config.Read(home); err != nil || got != dir {
		t.Errorf("Read = %q, %v; want %q", got, err, dir)
	}
}

func TestWriteRefusesBlankPathAndKeepsFile(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "/old/folder\n")

	if err := config.Write(home, " \t "); err == nil {
		t.Fatal("want an error")
	}

	if got := readFile(t, home); got != "/old/folder\n" {
		t.Errorf("file changed to %q", got)
	}
}
