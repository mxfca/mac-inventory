// Package install tests install.sh from the outside: it runs the script as a process
// with HOME in a temp folder and go/launchctl replaced by stubs on PATH.
package install

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const prompt = "Destination folder for the inventory"

// stub records its arguments in $STUB_LOG; the go stub also creates the -o output,
// and launchctl print fails so the script sees no previous agent.
const goStub = `#!/bin/sh
echo "go $*" >>"$STUB_LOG"
[ "$1" = build ] && : >"$3"
exit 0
`
const launchctlStub = `#!/bin/sh
echo "launchctl $*" >>"$STUB_LOG"
[ "$1" = print ] && exit 1
exit 0
`

type result struct {
	home, stdout, stderr, calls string
	exitCode                    int
}

func (r result) asked() bool { return strings.Contains(r.stdout, prompt) }
func (r result) installed() bool {
	return strings.Contains(r.calls, "go build") && strings.Contains(r.calls, "launchctl bootstrap")
}

func (r result) configFile() string { return filepath.Join(r.home, ".mac-inventory") }

// install runs install.sh with the given ~/.mac-inventory content (nil: no file) and answer on stdin;
// setup funcs prepare HOME further.
func install(t *testing.T, config *string, answer string, setup ...func(home string)) result {
	t.Helper()
	script, err := filepath.Abs("../../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	home, bin := t.TempDir(), t.TempDir()
	log := filepath.Join(t.TempDir(), "calls.log")
	for name, body := range map[string]string{"go": goStub, "launchctl": launchctlStub} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0o700); err != nil { //nolint:gosec // stubs must be executable
			t.Fatal(err)
		}
	}
	if config != nil {
		if err := os.WriteFile(filepath.Join(home, ".mac-inventory"), []byte(*config), 0o644); err != nil { //nolint:gosec // a loose mode proves the installer tightens it
			t.Fatal(err)
		}
	}

	for _, f := range setup {
		f(home)
	}

	cmd := exec.Command("/bin/sh", script) //nolint:gosec // script path is fixed by the repo layout
	cmd.Env = []string{"HOME=" + home, "PATH=" + bin + ":/usr/bin:/bin", "STUB_LOG=" + log}
	cmd.Stdin = strings.NewReader(answer + "\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	code := 0
	if exitErr := (&exec.ExitError{}); errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	calls, _ := os.ReadFile(log) //nolint:gosec // path built by this test
	return result{home, stdout.String(), stderr.String(), string(calls), code}
}

func str(s string) *string { return &s }

func TestKeepsValidConfigWithoutAsking(t *testing.T) {
	dest := t.TempDir()

	r := install(t, str(dest+"\n"), "")

	if r.exitCode != 0 || r.asked() || !r.installed() {
		t.Fatalf("exit %d, asked %v, installed %v; want 0, false, true\n%s", r.exitCode, r.asked(), r.installed(), r.stderr)
	}
	if got, _ := os.ReadFile(r.configFile()); string(got) != dest+"\n" {
		t.Errorf("config changed to %q", got)
	}
}

func TestFirstInstallAsksWritesPrivateConfigAndInstalls(t *testing.T) {
	dest := t.TempDir()

	r := install(t, nil, dest)

	if r.exitCode != 0 || !r.asked() || !r.installed() {
		t.Fatalf("exit %d, asked %v, installed %v; want 0, true, true\n%s", r.exitCode, r.asked(), r.installed(), r.stderr)
	}
	if strings.Contains(r.stdout, "invalid") {
		t.Errorf("stdout %q warns about a config that did not exist", r.stdout)
	}
	if got, _ := os.ReadFile(r.configFile()); string(got) != dest+"\n" {
		t.Errorf("config = %q, want %q", got, dest+"\n")
	}
	info, err := os.Stat(r.configFile())
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("config mode %v (err %v), want 0600", info.Mode().Perm(), err)
	}
}

func TestFirstInstallRefusesInvalidAnswer(t *testing.T) {
	for name, answer := range map[string]string{
		"relative":     "some/folder",
		"nonexistent":  "/no/such/folder",
		"empty answer": "",
	} {
		t.Run(name, func(t *testing.T) {
			r := install(t, nil, answer)

			if r.exitCode == 0 || r.calls != "" || r.stderr == "" {
				t.Errorf("exit %d, calls %q, stderr %q; want an error and no go/launchctl call", r.exitCode, r.calls, r.stderr)
			}
			if _, err := os.Stat(r.configFile()); err == nil {
				t.Error("config was written")
			}
		})
	}
}

func TestInvalidConfigWarnsAsksAndRepairs(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct{ config, reason string }{
		"empty":        {"", "empty"},
		"blank":        {" \n\t\n", "empty"},
		"relative":     {"some/folder\n", "relative"},
		"nonexistent":  {"/no/such/folder\n", "does not exist"},
		"not a folder": {file + "\n", "not a folder"},
	} {
		t.Run(name, func(t *testing.T) {
			dest := t.TempDir()

			r := install(t, str(tc.config), "  "+dest+"  ")

			if r.exitCode != 0 || !r.asked() || !r.installed() {
				t.Fatalf("exit %d, asked %v, installed %v; want 0, true, true\n%s", r.exitCode, r.asked(), r.installed(), r.stderr)
			}
			if !strings.Contains(r.stdout, tc.reason) {
				t.Errorf("stdout %q has no warning about %q", r.stdout, tc.reason)
			}
			if got, _ := os.ReadFile(r.configFile()); string(got) != dest+"\n" {
				t.Errorf("config = %q, want %q", got, dest+"\n")
			}
			info, err := os.Stat(r.configFile())
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Errorf("config mode %v (err %v), want 0600", info.Mode().Perm(), err)
			}
		})
	}
}

func TestInvalidConfigAndInvalidAnswerAbortsLeavingConfig(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for name, answer := range map[string]string{
		"relative":    "some/folder",
		"nonexistent": "/no/such/folder",
		"file":        file,
	} {
		t.Run(name, func(t *testing.T) {
			const old = "relative/old\n"

			r := install(t, str(old), answer)

			if r.exitCode == 0 || r.calls != "" || r.stderr == "" {
				t.Errorf("exit %d, calls %q, stderr %q; want an error and no go/launchctl call", r.exitCode, r.calls, r.stderr)
			}
			if got, _ := os.ReadFile(r.configFile()); string(got) != old {
				t.Errorf("config = %q, want it untouched (%q)", got, old)
			}
		})
	}
}

func TestConfigThatCannotBeWrittenAbortsWithoutInstalling(t *testing.T) {
	dest := t.TempDir()

	r := install(t, nil, dest, func(home string) {
		if err := os.Mkdir(filepath.Join(home, ".mac-inventory"), 0o700); err != nil {
			t.Fatal(err)
		}
	})

	if r.exitCode == 0 || r.calls != "" {
		t.Errorf("exit %d, calls %q; want an error and no go/launchctl call", r.exitCode, r.calls)
	}
}
