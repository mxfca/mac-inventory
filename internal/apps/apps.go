// Package apps lists the applications in /Applications from system_profiler.
package apps

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mxfca/mac-inventory/internal/runner"
)

// App is an application found in /Applications.
type App struct {
	Name     string
	Version  string // empty when system_profiler does not report one
	Bundle   string // file name in /Applications, like "Firefox.app"
	AppStore bool
}

type report struct {
	Apps []struct {
		Name         string `json:"_name"`
		Version      string `json:"version"`
		ObtainedFrom string `json:"obtained_from"`
		Path         string `json:"path"`
	} `json:"SPApplicationsDataType"`
}

// Collect returns the applications in /Applications, without the ones that ship with macOS.
func Collect(r runner.Runner) ([]App, error) {
	out, err := r.Run("system_profiler", "SPApplicationsDataType", "-json")
	if err != nil {
		return nil, err
	}
	var rep report
	if err := json.Unmarshal(out, &rep); err != nil {
		return nil, fmt.Errorf("unreadable JSON: %w", err)
	}

	var apps []App
	for _, a := range rep.Apps {
		if !strings.HasPrefix(a.Path, "/Applications/") || a.ObtainedFrom == "apple" {
			continue
		}
		apps = append(apps, App{Name: a.Name, Version: a.Version, Bundle: filepath.Base(a.Path), AppStore: a.ObtainedFrom == "mac_app_store"})
	}
	return apps, nil
}
