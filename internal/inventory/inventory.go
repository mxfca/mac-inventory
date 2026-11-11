// Package inventory groups the collected applications and Homebrew packages by source.
package inventory

import (
	"slices"
	"strings"

	"github.com/mxfca/mac-inventory/internal/apps"
	"github.com/mxfca/mac-inventory/internal/brew"
)

// Item is one entry in a section of the inventory: an application, formula or cask.
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

// New groups the applications found in /Applications and the Homebrew packages by source.
// An application installed by a cask is listed only under brew.
func New(found []apps.App, pkgs []brew.Package) Inventory {
	var inv Inventory
	fromCask := map[string]bool{}
	for _, p := range pkgs {
		inv.Brew = append(inv.Brew, Item{Name: p.Name, Version: p.Version})
		for _, bundle := range p.Apps {
			fromCask[bundle] = true
		}
	}
	for _, a := range found {
		item := Item{Name: a.Name, Version: a.Version}
		switch {
		case a.AppStore:
			inv.AppStore = append(inv.AppStore, item)
		case !fromCask[a.Bundle]:
			inv.Other = append(inv.Other, item)
		}
	}
	sortByName(inv.Brew)
	sortByName(inv.AppStore)
	sortByName(inv.Other)
	return inv
}

func sortByName(items []Item) {
	slices.SortStableFunc(items, func(a, b Item) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
}
