package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFolder(t *testing.T) {
	folder := t.TempDir()
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		dir     string
		wantErr bool
	}{
		"existing folder": {folder, false},
		"empty":           {"", true},
		"relative path":   {"OneDrive/Inventory", true},
		"missing folder":  {filepath.Join(folder, "missing"), true},
		"not a folder":    {file, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateFolder(c.dir); (err != nil) != c.wantErr {
				t.Errorf("validateFolder(%q) = %v, wantErr %v", c.dir, err, c.wantErr)
			}
		})
	}
}
