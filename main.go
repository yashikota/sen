package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	urfavecli "github.com/urfave/cli/v3"

	"github.com/yashikota/sen/internal/cli"
)

var Version string

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})))

	err := cli.Run(context.Background(), os.Args, os.Stdout, os.Stderr, getVersion())
	if err == nil {
		return
	}
	if ec, ok := err.(urfavecli.ExitCoder); ok {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(ec.ExitCode())
	}
	slog.Error("command failed", "err", err)
	os.Exit(1)
}

func getVersion() string {
	if Version != "" {
		return Version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
		if v, ok := getVCSBuildVersion(info); ok {
			return v
		}
	}

	return "(unset)"
}

func getVCSBuildVersion(info *debug.BuildInfo) (string, bool) {
	var (
		revision string
		dirty    string
	)

	for _, v := range info.Settings {
		switch v.Key {
		case "vcs.revision":
			revision = v.Value
		case "vcs.modified":
			if v.Value == "true" {
				dirty = " (dirty)"
			}
		}
	}

	if revision == "" {
		return "", false
	}

	return revision + dirty, true
}
