package runner_test

import (
	"testing"

	"github.com/mxfca/mac-inventory/internal/runner"
)

func TestExecReturnsStandardOutput(t *testing.T) {
	got, err := runner.Exec{}.Run("go", "version")

	if err != nil || len(got) == 0 {
		t.Errorf("got %q, %v; want the command output", got, err)
	}
}

func TestExecFailsOnMissingCommand(t *testing.T) {
	if _, err := (runner.Exec{}).Run("mac-inventory-no-such-command"); err == nil {
		t.Error("want an error for a missing command")
	}
}
