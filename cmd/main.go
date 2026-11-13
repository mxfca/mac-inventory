// Command mac-inventory writes the list of applications installed on this Mac
// into the destination folder set in ~/.mac-inventory.
package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/mxfca/mac-inventory/internal/apps"
	"github.com/mxfca/mac-inventory/internal/brew"
	"github.com/mxfca/mac-inventory/internal/config"
	"github.com/mxfca/mac-inventory/internal/inventory"
	"github.com/mxfca/mac-inventory/internal/notify"
	"github.com/mxfca/mac-inventory/internal/render"
	"github.com/mxfca/mac-inventory/internal/runner"
	"github.com/mxfca/mac-inventory/internal/store"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("mac-inventory", flag.ContinueOnError)
	flags.SetOutput(stderr)
	showLogs := flags.Bool("show-logs", false, "print logs to stderr")
	path := flags.String("path", "", "save `folder` as the destination and collect into it")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() > 0 {
		flags.Usage()
		return 2
	}

	log := slog.New(slog.DiscardHandler)
	if *showLogs {
		log = slog.New(slog.NewTextHandler(stderr, nil))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Error("home folder not found", "err", err)
		warn(log, "home folder not found: "+notify.Cause(err))
		return 1
	}
	if pathGiven(flags, "path") {
		if err := config.Write(home, *path); err != nil {
			_, _ = fmt.Fprintln(stderr, "mac-inventory: --path:", err)
			return 1
		}
		log.Info("destination saved", "dir", *path)
	}
	dir, err := config.Read(home)
	if err != nil {
		log.Error("reading destination", "err", err)
		warn(log, "destination not configured: set the folder in ~/.mac-inventory")
		return 1
	}
	log.Info("destination", "dir", dir)

	collectedAt := time.Now()
	found, appsErr := apps.Collect(runner.Exec{})
	if appsErr != nil {
		log.Error("collecting applications", "err", appsErr)
	}
	pkgs, brewErr := brew.Collect(runner.Exec{})
	if brewErr != nil {
		log.Error("collecting brew packages", "err", brewErr)
	}
	inv := inventory.New(found, appsErr, pkgs, brewErr)
	log.Info("collected", "brew", len(inv.Brew), "app_store", len(inv.AppStore), "other", len(inv.Other))

	if err := store.Write(dir, render.Markdown(inv, collectedAt), len(inv.Failures()) == 0, collectedAt); err != nil {
		log.Error("writing inventory", "err", err)
		warn(log, "could not write the inventory: "+notify.Cause(err))
		return 1
	}
	log.Info("inventory written", "file", store.FileName)
	return 0
}

// pathGiven tells --path "" (refused later by config) from no --path at all.
func pathGiven(flags *flag.FlagSet, name string) bool {
	given := false
	flags.Visit(func(f *flag.Flag) { given = given || f.Name == name })
	return given
}

// warn shows the notification; if it fails there is nothing left to do but log it.
func warn(log *slog.Logger, cause string) {
	if err := notify.Show(runner.Exec{}, "mac-inventory", cause); err != nil {
		log.Error("showing notification", "err", err)
	}
}
