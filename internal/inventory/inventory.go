// Package inventory groups the collected applications by source.
package inventory

import (
	"slices"
	"strings"

	"github.com/mxfca/mac-inventory/internal/apps"
)

// Item is one application in a section of the inventory.
type Item struct {
	Name    string
	Version string // empty when unknown
}

// Inventory holds the applications of each source, sorted by name ignoring case.
type Inventory struct {
	Brew     []Item
	AppStore []Item
	Other    []Item
}

// New groups the applications found in /Applications by source.
func New(found []apps.App) Inventory {
	var inv Inventory
	for _, a := range found {
		item := Item{Name: a.Name, Version: a.Version}
		if a.AppStore {
			inv.AppStore = append(inv.AppStore, item)
		} else {
			inv.Other = append(inv.Other, item)
		}
	}
	sortByName(inv.AppStore)
	sortByName(inv.Other)
	return inv
}

func sortByName(items []Item) {
	slices.SortStableFunc(items, func(a, b Item) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
}
