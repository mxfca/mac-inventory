// Command mac-inventory writes the list of applications installed on this Mac
// into the destination folder set in ~/.mac-inventory.
package main

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/mxfca/mac-inventory/internal/apps"
	"github.com/mxfca/mac-inventory/internal/brew"
	"github.com/mxfca/mac-inventory/internal/config"
	"github.com/mxfca/mac-inventory/internal/inventory"
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
		return 1
	}
	dir, err := config.Read(home)
	if err != nil {
		log.Error("reading destination", "err", err)
		return 1
	}
	log.Info("destination", "dir", dir)

	collectedAt := time.Now()
	found, err := apps.Collect(runner.Exec{})
	if err != nil {
		log.Error("collecting applications", "err", err)
		return 1
	}
	pkgs, err := brew.Collect(runner.Exec{})
	if err != nil {
		log.Error("collecting brew packages", "err", err)
		return 1
	}
	inv := inventory.New(found, pkgs)
	log.Info("collected", "brew", len(inv.Brew), "app_store", len(inv.AppStore), "other", len(inv.Other))

	if err := store.Write(dir, render.Markdown(inv, collectedAt)); err != nil {
		log.Error("writing inventory", "err", err)
		return 1
	}
	log.Info("inventory written", "file", store.FileName)
	return 0
}
