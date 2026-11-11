// Package brew lists what Homebrew installed: formulae installed on request and all casks.
package brew

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mxfca/mac-inventory/internal/runner"
)

// path is where brew lives on Apple Silicon; launchd's PATH does not include it.
const path = "/opt/homebrew/bin/brew"

// Package is a formula or cask installed by Homebrew.
type Package struct {
	Name    string   // the name brew install takes
	Version string   // the installed version
	Apps    []string // .app bundles a cask installs, like "Firefox.app"
}

type artifact struct {
	App    []any  `json:"app"`    // bundle names, sometimes followed by an options object
	Target string `json:"target"` // where the bundle lands, when the cask renames it
}

type report struct {
	Formulae []struct {
		FullName  string `json:"full_name"`
		Installed []struct {
			Version            string `json:"version"`
			InstalledOnRequest bool   `json:"installed_on_request"`
		} `json:"installed"`
	} `json:"formulae"`
	Casks []struct {
		FullToken string     `json:"full_token"`
		Installed string     `json:"installed"`
		Artifacts []artifact `json:"artifacts"`
	} `json:"casks"`
}

// Collect returns the formulae installed on request and all casks, formulae first.
func Collect(r runner.Runner) ([]Package, error) {
	out, err := r.Run(path, "info", "--json=v2", "--installed")
	if err != nil {
		return nil, err
	}
	var rep report
	if err := json.Unmarshal(out, &rep); err != nil {
		return nil, fmt.Errorf("unreadable JSON: %w", err)
	}

	var pkgs []Package
	for _, f := range rep.Formulae {
		for _, inst := range f.Installed {
			if inst.InstalledOnRequest {
				pkgs = append(pkgs, Package{Name: f.FullName, Version: inst.Version})
				break
			}
		}
	}
	for _, c := range rep.Casks {
		pkgs = append(pkgs, Package{Name: c.FullToken, Version: c.Installed, Apps: caskApps(c.Artifacts)})
	}
	return pkgs, nil
}

func caskApps(artifacts []artifact) []string {
	var apps []string
	for _, a := range artifacts {
		if len(a.App) > 0 && a.Target != "" {
			apps = append(apps, filepath.Base(a.Target))
			continue
		}
		for _, v := range a.App {
			if name, ok := v.(string); ok {
				apps = append(apps, name)
			}
		}
	}
	return apps
}
