package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLivedInWorkspaceTreeAndView(t *testing.T) {
	s := openTest(t)
	epic, err := s.CreateIssue(CreateIssueInput{Title: "Ship v1", Status: "todo", Priority: 1})
	if err != nil {
		t.Fatal(err)
	}
	feat, err := s.CreateIssue(CreateIssueInput{Title: "CLI", Status: "todo", ParentID: &epic.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "Flags", Status: "todo", ParentID: &feat.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "Done already", Status: "done"}); err != nil {
		t.Fatal(err)
	}

	all, err := s.ListIssues(IssueFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 4 {
		t.Fatalf("want 4 issues, got %d", len(all))
	}
	if all[0].Identifier != "SEN-1" || all[0].Depth != 0 {
		t.Fatalf("epic %#v", all[0])
	}
	if all[1].Identifier != "SEN-2" || all[1].Depth != 1 {
		t.Fatalf("feature %#v", all[1])
	}
	if all[2].Identifier != "SEN-3" || all[2].Depth != 2 {
		t.Fatalf("task %#v", all[2])
	}
	if all[3].Identifier != "SEN-4" || all[3].Depth != 0 {
		t.Fatalf("unrelated %#v", all[3])
	}

	st := "todo"
	pri := 1
	v, err := s.CreateView(CreateViewInput{Name: "Hot", Slug: "hot", Display: "list", Status: &st, Priority: &pri})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.ListIssues(v.Filter())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "Ship v1" {
		t.Fatalf("hot view %#v", got)
	}

	filtered, err := s.ListIssues(IssueFilter{Status: "todo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 3 {
		t.Fatalf("todo filter %#v", filtered)
	}
	if filtered[0].Depth != 0 || filtered[1].Depth != 1 || filtered[2].Depth != 2 {
		t.Fatalf("filtered depths %d %d %d", filtered[0].Depth, filtered[1].Depth, filtered[2].Depth)
	}
}

func TestDeleteParentOrphansDirectChildren(t *testing.T) {
	s := openTest(t)
	root, err := s.CreateIssue(CreateIssueInput{Title: "root"})
	if err != nil {
		t.Fatal(err)
	}
	mid, err := s.CreateIssue(CreateIssueInput{Title: "mid", ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "leaf", ParentID: &mid.ID}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteIssue("SEN-1"); err != nil {
		t.Fatal(err)
	}
	midGot, err := s.GetIssue("SEN-2")
	if err != nil {
		t.Fatal(err)
	}
	if midGot.ParentID != nil {
		t.Fatalf("mid still has parent %#v", midGot.ParentID)
	}
	leaf, err := s.GetIssue("SEN-3")
	if err != nil {
		t.Fatal(err)
	}
	if leaf.ParentIdentifier == nil || *leaf.ParentIdentifier != "SEN-2" {
		t.Fatalf("leaf parent %#v", leaf.ParentIdentifier)
	}
}

func TestAgentEditsCRLFAndDanglingParent(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	iss, err := s.CreateIssue(CreateIssueInput{Title: "from ui"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "issues", iss.Identifier+".md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	crlf := strings.ReplaceAll(string(raw), "\n", "\r\n")
	crlf = strings.Replace(crlf, "from ui", "edited on windows", 1)
	crlf = strings.Replace(crlf, "+++\r\n", "+++\r\nparent = \"SEN-99\"\r\n", 1)
	if err := os.WriteFile(path, []byte(crlf), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetIssue(iss.Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "edited on windows" {
		t.Fatalf("title %q", got.Title)
	}
	if got.ParentID != nil {
		t.Fatalf("dangling parent resolved to id %d", *got.ParentID)
	}
	diags, err := s.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range diags {
		if d.Code == "dangling_parent" {
			found = true
		}
	}
	if !found {
		t.Fatalf("want dangling_parent, got %#v", diags)
	}
}

func TestFilteredChildWithoutParentIsRoot(t *testing.T) {
	s := openTest(t)
	parent, err := s.CreateIssue(CreateIssueInput{Title: "epic", Status: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "still open", Status: "todo", ParentID: &parent.ID}); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListIssues(IssueFilter{Status: "todo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "still open" || got[0].Depth != 0 {
		t.Fatalf("orphan-in-filter %#v", got)
	}
}

func TestViewFileAndSearchAndDirty(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if _, err := s.CreateView(CreateViewInput{Name: "Bugs", Slug: "bugs", Display: "board"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "views", "bugs.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "name") || !strings.Contains(string(raw), "Bugs") {
		t.Fatalf("view file: %s", raw)
	}
	hits, err := s.Search("bug")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hits {
		if h.Kind == "view" && h.ID == "bugs" {
			found = true
		}
	}
	if !found {
		t.Fatalf("search %#v", hits)
	}
	dirty, err := s.Dirty()
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("view-only workspace should be dirty")
	}
}

func TestLabelFilterRequiresAll(t *testing.T) {
	s := openTest(t)
	labels, err := s.ListLabels()
	if err != nil || len(labels) < 2 {
		t.Fatalf("seeded labels: %v %#v", err, labels)
	}
	a := labels[0]
	b := labels[1]
	if _, err := s.CreateIssue(CreateIssueInput{Title: "both", LabelIDs: []int64{a.ID, b.ID}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "one", LabelIDs: []int64{a.ID}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListIssues(IssueFilter{Labels: []string{a.Name, b.Name}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "both" {
		t.Fatalf("AND labels %#v", got)
	}
}

func TestUnknownParentOnCreate(t *testing.T) {
	s := openTest(t)
	missing := int64(99)
	_, err := s.CreateIssue(CreateIssueInput{Title: "child", ParentID: &missing})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want validation, got %v", err)
	}
}
