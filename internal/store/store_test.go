package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestIssueNumberingDoesNotReuse(t *testing.T) {
	s := openTest(t)
	a, err := s.CreateIssue(CreateIssueInput{Title: "one"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateIssue(CreateIssueInput{Title: "two"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Identifier != "SEN-1" || b.Identifier != "SEN-2" {
		t.Fatalf("got %s %s", a.Identifier, b.Identifier)
	}
	if err := s.DeleteIssue("SEN-1"); err != nil {
		t.Fatal(err)
	}
	c, err := s.CreateIssue(CreateIssueInput{Title: "three"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Identifier != "SEN-3" {
		t.Fatalf("reused number: %s", c.Identifier)
	}
}

func TestSingleActiveCycle(t *testing.T) {
	s := openTest(t)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	c1, err := s.CreateCycle(start, end, "active")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := s.CreateCycle(start, end, "active")
	if err != nil {
		t.Fatal(err)
	}
	got1, err := s.GetCycle(c1.Number)
	if err != nil {
		t.Fatal(err)
	}
	if got1.Status != "completed" {
		t.Fatalf("previous active want completed, got %s", got1.Status)
	}
	if c2.Status != "active" {
		t.Fatalf("new cycle want active, got %s", c2.Status)
	}
	all, err := s.ListCycles()
	if err != nil {
		t.Fatal(err)
	}
	active := 0
	for _, c := range all {
		if c.Status == "active" {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("active count %d", active)
	}
}

func TestDirtyWithoutUserContentAllowsPull(t *testing.T) {
	s := openTest(t)
	dirty, err := s.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("fresh workspace should not be dirty")
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "x"}); err != nil {
		t.Fatal(err)
	}
	dirty, err = s.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("unpushed issue should be dirty")
	}
	if err := s.MarkPushed("sha256:abc"); err != nil {
		t.Fatal(err)
	}
	dirty, err = s.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("just pushed should not be dirty")
	}
}

func TestCreateIssueValidation(t *testing.T) {
	s := openTest(t)
	_, err := s.CreateIssue(CreateIssueInput{Title: " "})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want validation, got %v", err)
	}
}

func TestPageRejectsCycle(t *testing.T) {
	s := openTest(t)
	p, err := s.CreatePage("Root", "root", "", "proposed", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CreatePage("Child", "child", "", "proposed", &p.ID, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	parentID := &child.ID
	_, err = s.UpdatePage("root", nil, nil, nil, &parentID, nil, nil, nil)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want cycle validation, got %v", err)
	}
}

func TestIssueParentRejectsCycle(t *testing.T) {
	s := openTest(t)
	a, err := s.CreateIssue(CreateIssueInput{Title: "parent"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateIssue(CreateIssueInput{Title: "child", ParentID: &a.ID})
	if err != nil {
		t.Fatal(err)
	}
	if b.ParentIdentifier == nil || *b.ParentIdentifier != "SEN-1" {
		t.Fatalf("parent identifier %#v", b.ParentIdentifier)
	}
	parentID := &b.ID
	_, err = s.UpdateIssue("SEN-1", PatchIssueInput{ParentID: &parentID})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want cycle validation, got %v", err)
	}
}

func TestIssueParentWritesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	parent, err := s.CreateIssue(CreateIssueInput{Title: "root"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := s.CreateIssue(CreateIssueInput{Title: "leaf", ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "issues", child.Identifier+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "parent =") || !strings.Contains(string(raw), "SEN-1") {
		t.Fatalf("frontmatter: %s", raw)
	}
}

func TestViewFiltersIssues(t *testing.T) {
	s := openTest(t)
	if _, err := s.CreateIssue(CreateIssueInput{Title: "open", Status: "todo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "closed", Status: "done"}); err != nil {
		t.Fatal(err)
	}
	st := "todo"
	v, err := s.CreateView(CreateViewInput{Name: "Open", Slug: "open", Display: "list", Status: &st})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.ListIssues(v.Filter())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "open" {
		t.Fatalf("got %#v", got)
	}
}

func TestIssueWritesMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	iss, err := s.CreateIssue(CreateIssueInput{Title: "file native"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "issues", iss.Identifier+".md"))
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)
	if !strings.Contains(md, "file native") || !strings.HasPrefix(md, "+++\n") {
		t.Fatalf("issue markdown: %s", md)
	}
}

func TestReloadSeesDiskEdits(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	iss, err := s.CreateIssue(CreateIssueInput{Title: "original"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "issues", iss.Identifier+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(b), "original", "edited by agent", 1)
	if edited == string(b) {
		t.Fatal("title not found in file")
	}
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetIssue(iss.Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "edited by agent" {
		t.Fatalf("title %q", got.Title)
	}
}

func TestDiagnosticsDanglingProject(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	iss, err := s.CreateIssue(CreateIssueInput{Title: "loose"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "issues", iss.Identifier+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(b), "+++\n", "+++\nproject = \"missing\"\n", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetIssue(iss.Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "loose" {
		t.Fatalf("title %q", got.Title)
	}
	diags, err := s.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range diags {
		if d.Code == "dangling_project" {
			found = true
		}
	}
	if !found {
		t.Fatalf("want dangling_project, got %#v", diags)
	}
}

func TestOpenRejectsLegacyYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte("name: sen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Open(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "workspace.yaml") {
		t.Fatal(err)
	}
}
