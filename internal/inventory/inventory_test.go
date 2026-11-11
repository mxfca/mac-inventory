package inventory_test

import (
	"reflect"
	"testing"

	"github.com/mxfca/mac-inventory/internal/apps"
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
	})

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
