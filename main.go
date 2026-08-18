package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/yashikota/sen/internal/cli"
)

var Version string

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})))

	err := cli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, getVersion())
	if err == nil {
		return
	}
	if errors.Is(err, cli.ErrUsage) {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
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
