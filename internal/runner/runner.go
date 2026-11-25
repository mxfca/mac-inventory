// Package runner runs external commands; the other packages receive it as an interface.
package runner

import (
	"context"
	"os/exec"
)

// Runner runs a command and returns its standard output.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Exec runs commands with os/exec.
type Exec struct{}

// Run runs name with args and returns its standard output; cancelling ctx kills the command.
func (Exec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output() // #nosec G204 -- callers pass fixed commands
}
