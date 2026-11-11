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

	// BrewFailure and AppsFailure hold the cause, on one line, when that source failed to
	// collect; empty otherwise. AppsFailure covers system_profiler (App Store and outro).
	BrewFailure string
	AppsFailure string
}

// New groups the applications found in /Applications and the Homebrew packages by source.
// An application installed by a cask is listed only under brew.
// A non-nil appsErr or brewErr marks that source as failed (its list is then empty).
func New(found []apps.App, appsErr error, pkgs []brew.Package, brewErr error) Inventory {
	inv := Inventory{BrewFailure: cause(brewErr), AppsFailure: cause(appsErr)}
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

// Failures lists one "<source>: <cause>" entry per source that failed, brew first.
func (inv Inventory) Failures() []string {
	var out []string
	if inv.BrewFailure != "" {
		out = append(out, "brew: "+inv.BrewFailure)
	}
	if inv.AppsFailure != "" {
		out = append(out, "system_profiler: "+inv.AppsFailure)
	}
	return out
}

func cause(err error) string {
	if err == nil {
		return ""
	}
	if msg := strings.Join(strings.Fields(err.Error()), " "); msg != "" {
		return msg
	}
	return "unknown error" // an empty cause would read as "no failure"
}
