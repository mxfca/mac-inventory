package apps_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/mxfca/mac-inventory/internal/apps"
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

func TestCollectKeepsApplicationsFolderAndClassifiesSource(t *testing.T) {
	out, err := os.ReadFile("testdata/system_profiler.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &fakeRunner{out: out}

	got, err := apps.Collect(r)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if r.name != "system_profiler" || !reflect.DeepEqual(r.args, []string{"SPApplicationsDataType", "-json"}) {
		t.Errorf("ran %s %v", r.name, r.args)
	}
	want := []apps.App{
		{Name: "Claude", Version: "5.1231.0", Bundle: "Claude.app"},
		{Name: "WhatsApp", Version: "26.44.72", Bundle: "WhatsApp.app", AppStore: true},
		{Name: "Firefox", Version: "159.0.1", Bundle: "Firefox.app"},
		{Name: "Keynote", Version: "15.2.1", Bundle: "Keynote Creator Studio.app", AppStore: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestCollectDropsAppleAppsInApplicationsFolder(t *testing.T) {
	r := &fakeRunner{out: []byte(`{"SPApplicationsDataType": [
		{"_name": "Safari", "obtained_from": "apple", "path": "/Applications/Safari.app", "version": "26.0"},
		{"_name": "Zoom", "obtained_from": "unknown", "path": "/Applications/zoom.us.app"}
	]}`)}

	got, err := apps.Collect(r)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	want := []apps.App{{Name: "Zoom", Bundle: "zoom.us.app"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestCollectFailsWhenCommandFails(t *testing.T) {
	r := &fakeRunner{err: errors.New("exit status 1")}

	_, err := apps.Collect(r)

	if err == nil || err.Error() != "exit status 1" {
		t.Errorf("got %v, want the bare command error (inventory adds the system_profiler label)", err)
	}
}

func TestCollectFailsOnUnreadableJSON(t *testing.T) {
	r := &fakeRunner{out: []byte("not json")}

	if _, err := apps.Collect(r); err == nil {
		t.Error("got nil error for unreadable JSON")
	}
}
