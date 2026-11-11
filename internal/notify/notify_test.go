package notify_test

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"testing"

	"github.com/mxfca/mac-inventory/internal/notify"
)

type fakeRunner struct {
	err  error
	name string
	args []string
}

func (f *fakeRunner) Run(name string, args ...string) ([]byte, error) {
	f.name, f.args = name, args
	return nil, f.err
}

func TestShowAsksOsascriptForANotificationWithTitleAndCause(t *testing.T) {
	r := &fakeRunner{}

	if err := notify.Show(r, "mac-inventory", "folder not found"); err != nil {
		t.Fatalf("Show: %v", err)
	}

	want := []string{"-e", `display notification "folder not found" with title "mac-inventory"`}
	if r.name != "osascript" || !reflect.DeepEqual(r.args, want) {
		t.Errorf("ran %s %q", r.name, r.args)
	}
}

func TestShowEscapesQuotesAndBackslashes(t *testing.T) {
	r := &fakeRunner{}

	if err := notify.Show(r, "mac-inventory", `open "a\b": denied`); err != nil {
		t.Fatalf("Show: %v", err)
	}

	want := []string{"-e", `display notification "open \"a\\b\": denied" with title "mac-inventory"`}
	if !reflect.DeepEqual(r.args, want) {
		t.Errorf("args %q", r.args)
	}
}

func TestShowReturnsTheCommandError(t *testing.T) {
	r := &fakeRunner{err: errors.New("boom")}

	if err := notify.Show(r, "t", "c"); err == nil {
		t.Fatal("Show returned nil")
	}
}

func TestCauseDropsTheFilePathFromAFileError(t *testing.T) {
	err := fmt.Errorf("writing: %w", &fs.PathError{Op: "open", Path: "/long/path/inventory.md", Err: errors.New("no such file or directory")})

	if got := notify.Cause(err); got != "no such file or directory" {
		t.Errorf("Cause = %q", got)
	}
}

func TestCauseKeepsOtherErrorsAsTheyAre(t *testing.T) {
	if got := notify.Cause(errors.New("boom")); got != "boom" {
		t.Errorf("Cause = %q", got)
	}
}
