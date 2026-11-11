package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mxfca/mac-inventory/internal/config"
)

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, ".mac-inventory"), []byte(content), 0o600); err != nil {
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
