package inventory_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mxfca/mac-inventory/internal/apps"
	"github.com/mxfca/mac-inventory/internal/brew"
	"github.com/mxfca/mac-inventory/internal/inventory"
)

func TestNewSplitsBySourceAndSortsByNameIgnoringCase(t *testing.T) {
	got := inventory.New([]apps.App{
		{Name: "zoom.us", Version: "6.6.0"},
		{Name: "WhatsApp", Version: "26.44.72", AppStore: true},
		{Name: "Bruno", Version: "4.3.0"},
		{Name: "iMovie", Version: "10.4.4", AppStore: true},
		{Name: "Keynote", Version: "15.2.1", AppStore: true},
		{Name: "anki", Version: "25.0"},
	}, nil, nil, nil)

	want := inventory.Inventory{
		AppStore: []inventory.Item{
			{Name: "iMovie", Version: "10.4.4"},
			{Name: "Keynote", Version: "15.2.1"},
			{Name: "WhatsApp", Version: "26.44.72"},
		},
		Other: []inventory.Item{
			{Name: "anki", Version: "25.0"},
			{Name: "Bruno", Version: "4.3.0"},
			{Name: "zoom.us", Version: "6.6.0"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestNewListsBrewPackagesAndDropsCaskAppsFromOther(t *testing.T) {
	got := inventory.New([]apps.App{
		{Name: "Firefox", Version: "159.0.1", Bundle: "Firefox.app"},
		{Name: "Zoom", Version: "6.6.0", Bundle: "zoom.us.app"},
		{Name: "Creality Print", Version: "7.3.0", Bundle: "Creality Print.app"},
		{Name: "Keynote", Version: "15.2.1", Bundle: "Keynote.app", AppStore: true},
	}, nil, []brew.Package{
		{Name: "gh", Version: "2.103.3"},
		{Name: "firefox", Version: "159.0.1", Apps: []string{"Firefox.app"}},
		{Name: "Bruno", Version: "4.3.0", Apps: []string{"Bruno.app"}},
		{Name: "creality-print", Version: "7.3.0", Apps: []string{"Creality Print.app"}},
	}, nil)

	want := inventory.Inventory{
		Brew: []inventory.Item{
			{Name: "Bruno", Version: "4.3.0"},
			{Name: "creality-print", Version: "7.3.0"},
			{Name: "firefox", Version: "159.0.1"},
			{Name: "gh", Version: "2.103.3"},
		},
		AppStore: []inventory.Item{{Name: "Keynote", Version: "15.2.1"}},
		Other:    []inventory.Item{{Name: "Zoom", Version: "6.6.0"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestNewWhenBrewFailsKeepsCaskAppsInOther(t *testing.T) {
	got := inventory.New([]apps.App{
		{Name: "Firefox", Version: "159.0.1", Bundle: "Firefox.app"},
	}, nil, nil, errors.New("exit status 1"))

	want := inventory.Inventory{
		Other:       []inventory.Item{{Name: "Firefox", Version: "159.0.1"}},
		BrewFailure: "exit status 1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	if f := got.Failures(); !reflect.DeepEqual(f, []string{"brew: exit status 1"}) {
		t.Errorf("failures = %q", f)
	}
}

func TestNewWhenSystemProfilerFailsKeepsBrew(t *testing.T) {
	got := inventory.New(nil, errors.New("invalid JSON"), []brew.Package{{Name: "gh", Version: "2.103.3"}}, nil)

	want := inventory.Inventory{
		Brew:        []inventory.Item{{Name: "gh", Version: "2.103.3"}},
		AppsFailure: "invalid JSON",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
	if f := got.Failures(); !reflect.DeepEqual(f, []string{"system_profiler: invalid JSON"}) {
		t.Errorf("failures = %q", f)
	}
}

func TestNewWhenBothFailListsBothFailuresOnOneLineEach(t *testing.T) {
	got := inventory.New(nil, errors.New("timeout\nafter 5s"), nil, errors.New("exit status 1"))

	want := []string{"brew: exit status 1", "system_profiler: timeout after 5s"}
	if f := got.Failures(); !reflect.DeepEqual(f, want) {
		t.Errorf("failures = %q, want %q", f, want)
	}
}

func TestNewFailureWithEmptyMessageStillCountsAsFailed(t *testing.T) {
	got := inventory.New(nil, errors.New(""), nil, nil)

	if f := got.Failures(); !reflect.DeepEqual(f, []string{"system_profiler: unknown error"}) {
		t.Errorf("failures = %q", f)
	}
}
