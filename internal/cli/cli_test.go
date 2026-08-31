package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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

func TestRunAcceptsGoRunBinaryPath(t *testing.T) {
	var stdout bytes.Buffer
	err := cli.Run(context.Background(), []string{"/tmp/go-build/exe/sen", "version"}, &stdout, ioDiscard{}, "9.9.9")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "9.9.9" {
		t.Fatalf("version = %q, want 9.9.9", stdout.String())
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

func TestInitCheckAndDirtyStatus(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEN_HOME", dir)

	var stdout bytes.Buffer
	if err := cli.Run(context.Background(), []string{"sen", "init"}, &stdout, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), dir) {
		t.Fatalf("init output %q", stdout.String())
	}

	stdout.Reset()
	if err := cli.Run(context.Background(), []string{"sen", "check"}, &stdout, ioDiscard{}, "test"); err != nil {
		t.Fatalf("fresh check: %v %s", err, stdout.String())
	}
	if strings.TrimSpace(stdout.String()) != "ok" {
		t.Fatalf("check %q", stdout.String())
	}

	stdout.Reset()
	if err := cli.Run(context.Background(), []string{"sen", "status"}, &stdout, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "dirty\tfalse") {
		t.Fatalf("status %q", stdout.String())
	}

	issue := filepath.Join(dir, "issues")
	if err := os.MkdirAll(issue, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "+++\ntitle = 'loose'\nstatus = 'todo'\npriority = 0\nproject = \"missing\"\ncreated = '2026-01-01T00:00:00Z'\nupdated = '2026-01-01T00:00:00Z'\n+++\n\n"
	if err := os.WriteFile(filepath.Join(issue, "SEN-1.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	err := cli.Run(context.Background(), []string{"sen", "check"}, &stdout, ioDiscard{}, "test")
	if err == nil {
		t.Fatal("expected check failure")
	}
	if !strings.Contains(stdout.String(), "unknown project") && !strings.Contains(stdout.String(), "missing") {
		t.Fatalf("check diagnostics %q err=%v", stdout.String(), err)
	}
}

func TestRunDefaultIsHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := cli.Run(context.Background(), []string{"sen"}, &stdout, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "init") {
		t.Fatalf("default help: %s", stdout.String())
	}
}

func TestRunVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	if err := cli.Run(context.Background(), []string{"sen", "version"}, &stdout, ioDiscard{}, "4.5.6"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stdout.String()) != "4.5.6" {
		t.Fatalf("version %q", stdout.String())
	}
}

func TestInitAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEN_HOME", dir)
	if err := cli.Run(context.Background(), []string{"sen", "init"}, ioDiscard{}, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	err := cli.Run(context.Background(), []string{"sen", "init"}, ioDiscard{}, ioDiscard{}, "test")
	if err == nil {
		t.Fatal("expected already initialized")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("got %v", err)
	}
}

func TestStatusAndCheckWithoutWorkspace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEN_HOME", dir)
	err := cli.Run(context.Background(), []string{"sen", "status"}, ioDiscard{}, ioDiscard{}, "test")
	if err == nil {
		t.Fatal("expected missing workspace")
	}
	if !strings.Contains(err.Error(), "workspace not found") {
		t.Fatalf("status %v", err)
	}
	err = cli.Run(context.Background(), []string{"sen", "check"}, ioDiscard{}, ioDiscard{}, "test")
	if err == nil {
		t.Fatal("expected missing workspace")
	}
	if !strings.Contains(err.Error(), "run sen init") {
		t.Fatalf("check %v", err)
	}
}

func TestStatusDirtyAfterIssueFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEN_HOME", dir)
	if err := cli.Run(context.Background(), []string{"sen", "init"}, ioDiscard{}, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	issue := filepath.Join(dir, "issues")
	if err := os.MkdirAll(issue, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "+++\ntitle = 'real'\nstatus = 'todo'\npriority = 0\ncreated = '2026-01-01T00:00:00Z'\nupdated = '2026-01-01T00:00:00Z'\n+++\n\n"
	if err := os.WriteFile(filepath.Join(issue, "SEN-1.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := cli.Run(context.Background(), []string{"sen", "status"}, &stdout, ioDiscard{}, "test"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "dirty\ttrue") {
		t.Fatalf("status %q", stdout.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
