package inventory_test

import (
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
	}, nil)

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
	}, []brew.Package{
		{Name: "gh", Version: "2.103.3"},
		{Name: "firefox", Version: "159.0.1", Apps: []string{"Firefox.app"}},
		{Name: "Bruno", Version: "4.3.0", Apps: []string{"Bruno.app"}},
		{Name: "creality-print", Version: "7.3.0", Apps: []string{"Creality Print.app"}},
	})

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
