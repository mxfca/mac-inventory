package brew_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mxfca/mac-inventory/internal/brew"
)

type fakeRunner struct {
	out  []byte
	err  error
	name string
	args []string
}

func (f *fakeRunner) Run(name string, args ...string) ([]byte, error) {
	f.name, f.args = name, args
	return f.out, f.err
}

func TestCollectKeepsRequestedFormulaeAndAllCasks(t *testing.T) {
	out, err := os.ReadFile("testdata/brew_info.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &fakeRunner{out: out}

	got, err := brew.Collect(r)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if r.name != "/opt/homebrew/bin/brew" || !reflect.DeepEqual(r.args, []string{"info", "--json=v2", "--installed"}) {
		t.Errorf("ran %s %v", r.name, r.args)
	}
	want := []brew.Package{
		{Name: "gh", Version: "2.103.3"},
		{Name: "pre-commit", Version: "4.7.0_2"},
		{Name: "renamed-example", Version: "1.0", Apps: []string{"Example.app"}},
		{Name: "bruno", Version: "4.3.0", Apps: []string{"Bruno.app"}},
		{Name: "claude", Version: "4.4212.0,cbfe3c56d49c49332ad5a9588303c9dbc70a5e7f", Apps: []string{"Claude.app"}},
		{Name: "codex", Version: "0.166.0"},
		{Name: "creality-print", Version: "7.3.0.6157", Apps: []string{"Creality Print.app"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestCollectFailsWhenCommandFails(t *testing.T) {
	r := &fakeRunner{err: errors.New("exit status 1")}

	_, err := brew.Collect(r)

	if err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Errorf("got %v, want the command error", err)
	}
}

func TestCollectFailsOnUnreadableJSON(t *testing.T) {
	r := &fakeRunner{out: []byte("not json")}

	if _, err := brew.Collect(r); err == nil {
		t.Error("got nil error for unreadable JSON")
	}
}
