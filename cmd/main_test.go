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

// noHome leaves HOME empty so os.UserHomeDir fails, and puts a fake osascript on PATH
// that records its arguments; the returned file holds what the notification would show.
func noHome(t *testing.T) (shown string) {
	t.Helper()
	bin := t.TempDir()
	shown = filepath.Join(t.TempDir(), "osascript-args")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > '" + shown + "'\n"
	if err := os.WriteFile(filepath.Join(bin, "osascript"), []byte(script), 0o700); err != nil { //nolint:gosec // test stub must be executable
		t.Fatal(err)
	}
	t.Setenv("HOME", "")
	t.Setenv("PATH", bin)
	return shown
}

func TestRunWarnsWhenHomeFolderNotFound(t *testing.T) {
	shown := noHome(t)

	code, stderr := runWith("--show-logs")

	if code != 1 || !strings.Contains(stderr, "home folder not found") {
		t.Errorf("exit %d, stderr %q; want 1 and a home error in the log", code, stderr)
	}
	args, err := os.ReadFile(shown) //nolint:gosec // path built by the test
	if err != nil {
		t.Fatalf("no notification shown: %v", err)
	}
	if want := `display notification "home folder not found: $HOME is not defined"`; !strings.Contains(string(args), want) {
		t.Errorf("notification %q; want it to contain %q", args, want)
	}
}

func TestRunWithPathStillWarnsWhenHomeFolderNotFound(t *testing.T) {
	shown := noHome(t)

	dest := t.TempDir()

	code, _ := runWith("--path", dest)

	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	if _, err := os.Stat(shown); err != nil {
		t.Errorf("no notification shown: %v", err)
	}
	if entries, err := os.ReadDir(dest); err != nil || len(entries) != 0 {
		t.Errorf("destination has %d entries, %v; want it untouched", len(entries), err)
	}
}

func TestRunLogsWhenHomeWarningCannotBeShown(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("PATH", t.TempDir())

	code, stderr := runWith("--show-logs")

	if code != 1 || !strings.Contains(stderr, "showing notification") {
		t.Errorf("exit %d, stderr %q; want 1 and a notification error in the log", code, stderr)
	}
}
