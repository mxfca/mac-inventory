// Package runner runs external commands; the other packages receive it as an interface.
package runner

import "os/exec"

// Runner runs a command and returns its standard output.
type Runner interface {
	Run(name string, args ...string) ([]byte, error)
}

// Exec runs commands with os/exec.
type Exec struct{}

// Run runs name with args and returns its standard output.
func (Exec) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output() // #nosec G204 -- callers pass fixed commands
}
