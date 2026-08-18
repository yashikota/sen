package syncer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashikota/sen/internal/store"
)

func TestExportWritesFrontmatter(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.CreatePage("Keep files", "keep-files", "Body here.\n", "proposed", nil, nil, nil, []string{"adr"}); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: st, Version: "test"}
	dir := t.TempDir()
	man, err := svc.Export(dir)
	if err != nil {
		t.Fatal(err)
	}
	if man.Pages != 1 {
		t.Fatalf("pages %d", man.Pages)
	}
	b, err := os.ReadFile(filepath.Join(dir, "pages", "keep-files.md"))
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)
	if !strings.Contains(md, "Keep files") || !strings.Contains(md, "+++\n") {
		t.Fatalf("frontmatter missing: %s", md)
	}
	if _, err := os.Stat(filepath.Join(dir, "workspace.toml")); err != nil {
		t.Fatal(err)
	}
}

func TestPushDoesNotMarkOnFailure(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ref := "ghcr.io/example/sen"
	if _, err := st.UpdateWorkspace(nil, &ref, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateIssue(store.CreateIssueInput{Title: "keep"}); err != nil {
		t.Fatal(err)
	}
	reg := NewMemory()
	reg.FailPush = errors.New("network down")
	svc := &Service{Store: st, Registry: reg, Version: "test"}
	if _, _, err := svc.Push(context.Background()); err == nil {
		t.Fatal("expected push error")
	}
	ws, err := st.Workspace()
	if err != nil {
		t.Fatal(err)
	}
	if ws.LastPushedDigest != nil {
		t.Fatalf("digest written on failure: %v", *ws.LastPushedDigest)
	}
}

func TestPullRefusesDirty(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ref := "ghcr.io/example/sen"
	if _, err := st.UpdateWorkspace(nil, &ref, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateIssue(store.CreateIssueInput{Title: "local"}); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: st, Registry: NewMemory(), Version: "test"}
	err = svc.Pull(context.Background(), "latest")
	if !errors.Is(err, ErrDirty) {
		t.Fatalf("got %v", err)
	}
}

func TestPushPullRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	ref := "ghcr.io/example/sen"
	if _, err := st.UpdateWorkspace(nil, &ref, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateIssue(store.CreateIssueInput{Title: "shipped"}); err != nil {
		t.Fatal(err)
	}
	reg := NewMemory()
	svc := &Service{Store: st, Registry: reg, Version: "test"}
	if _, _, err := svc.Push(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	otherDir := t.TempDir()
	st2, err := store.Open(otherDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st2.UpdateWorkspace(nil, &ref, nil); err != nil {
		t.Fatal(err)
	}
	svc2 := &Service{Store: st2, Registry: reg, Version: "test"}
	if err := svc2.Pull(context.Background(), "latest"); err != nil {
		t.Fatal(err)
	}
	if err := st2.Close(); err != nil {
		t.Fatal(err)
	}
	st3, err := store.Open(otherDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st3.Close() })
	iss, err := st3.GetIssue("SEN-1")
	if err != nil {
		t.Fatal(err)
	}
	if iss.Title != "shipped" {
		t.Fatalf("title %q", iss.Title)
	}
}

func TestPullRejectsLegacySQLiteArtifact(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ref := "ghcr.io/example/sen"
	if _, err := st.UpdateWorkspace(nil, &ref, nil); err != nil {
		t.Fatal(err)
	}
	art := t.TempDir()
	if err := os.WriteFile(filepath.Join(art, "workspace.sqlite"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg := NewMemory()
	if _, err := reg.Push(context.Background(), ref, "latest", art); err != nil {
		t.Fatal(err)
	}
	svc := &Service{Store: st, Registry: reg, Version: "test"}
	err = svc.Pull(context.Background(), "latest")
	if !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("got %v", err)
	}
}
