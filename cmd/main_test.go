package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mxfca/mac-inventory/internal/config"
	"github.com/mxfca/mac-inventory/internal/store"
)

// sandbox isolates run from the real machine: HOME is a temp folder and PATH is empty,
// so system_profiler, brew and osascript are not found (no real data, no notifications).
func sandbox(t *testing.T) (home string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	return home
}

func runWith(args ...string) (int, string) {
	var stderr bytes.Buffer
	code := run(args, &stderr)
	return code, stderr.String()
}

func TestRunRejectsBadFlagsAndArguments(t *testing.T) {
	sandbox(t)
	for name, args := range map[string][]string{
		"unknown flag":     {"--nope"},
		"positional value": {"extra"},
	} {
		if code, _ := runWith(args...); code != 2 {
			t.Errorf("%s: exit %d, want 2", name, code)
		}
	}
}

func TestRunFailsWithoutDestination(t *testing.T) {
	sandbox(t)

	code, stderr := runWith("--show-logs")

	if code != 1 || !strings.Contains(stderr, "reading destination") {
		t.Errorf("exit %d, stderr %q; want 1 and a destination error", code, stderr)
	}
}

func TestRunRefusesInvalidPath(t *testing.T) {
	home := sandbox(t)

	code, stderr := runWith("--path", "")

	if code != 1 || !strings.Contains(stderr, "--path") {
		t.Errorf("exit %d, stderr %q; want 1 and a --path error", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(home, config.FileName)); err == nil {
		t.Error("config file was written for an invalid path")
	}
}

func TestRunSavesPathAndWritesInventory(t *testing.T) {
	home := sandbox(t)
	dest := t.TempDir()

	code, _ := runWith("--path", dest, "--show-logs")

	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if got, err := config.Read(home); err != nil || got != dest {
		t.Errorf("saved destination %q, %v; want %q", got, err, dest)
	}
	if _, err := os.Stat(filepath.Join(dest, store.FileName)); err != nil {
		t.Errorf("inventory not written: %v", err)
	}
}

func TestRunUsesConfiguredDestination(t *testing.T) {
	home := sandbox(t)
	dest := t.TempDir()
	if err := config.Write(home, dest); err != nil {
		t.Fatal(err)
	}

	if code, _ := runWith(); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(dest, store.FileName)); err != nil {
		t.Errorf("inventory not written: %v", err)
	}
}

func TestRunFailsWhenInventoryCannotBeWritten(t *testing.T) {
	home := sandbox(t)
	dest := t.TempDir()
	if err := config.Write(home, dest); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dest); err != nil {
		t.Fatal(err)
	}

	code, stderr := runWith("--show-logs")

	if code != 1 || !strings.Contains(stderr, "writing inventory") {
		t.Errorf("exit %d, stderr %q; want 1 and a write error", code, stderr)
	}
}
