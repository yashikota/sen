package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	urfavecli "github.com/urfave/cli/v3"

	"github.com/yashikota/sen/internal/cli"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	err := cli.Run(context.Background(), []string{"sen", "help"}, &stdout, ioDiscard{}, "test")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "init") || !strings.Contains(out, "serve") {
		t.Fatalf("help output missing commands: %s", out)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	err := cli.Run(context.Background(), []string{"sen", "not-a-command"}, ioDiscard{}, ioDiscard{}, "test")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	ec, ok := err.(urfavecli.ExitCoder)
	if !ok {
		t.Fatalf("expected ExitCoder, got %T: %v", err, err)
	}
	if ec.ExitCode() != 3 {
		t.Fatalf("exit code = %d, want 3", ec.ExitCode())
	}
}

func TestRunVersionFlag(t *testing.T) {
	var stdout bytes.Buffer
	err := cli.Run(context.Background(), []string{"sen", "--version"}, &stdout, ioDiscard{}, "1.2.3")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", stdout.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
