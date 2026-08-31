package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateIssueDefaultsAndValidation(t *testing.T) {
	s := openTest(t)
	iss, err := s.CreateIssue(CreateIssueInput{Title: "untitled"})
	if err != nil {
		t.Fatal(err)
	}
	if iss.Status != "backlog" {
		t.Fatalf("default status %q", iss.Status)
	}
	if iss.Priority != 0 {
		t.Fatalf("default priority %d", iss.Priority)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "bad", Status: "ready"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid status: %v", err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "bad", Priority: 5}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid priority: %v", err)
	}
	if _, err := s.GetIssue("SEN-99"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing issue: %v", err)
	}
	if err := s.DeleteIssue("SEN-99"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
}

func TestCreateLabelValidation(t *testing.T) {
	s := openTest(t)
	if _, err := s.CreateLabel(" ", "#aabbcc"); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := s.CreateLabel("Ok", "red"); !errors.Is(err, ErrValidation) {
		t.Fatalf("bad color: %v", err)
	}
	if _, err := s.CreateLabel("Ok", "#gg0000"); !errors.Is(err, ErrValidation) {
		t.Fatalf("non-hex color: %v", err)
	}
	if _, err := s.CreateLabel("Bug", "#aabbcc"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate seeded name: %v", err)
	}
	got, err := s.CreateLabel("Harbor", "#6B9BD1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Harbor" || got.Color != "#6B9BD1" {
		t.Fatalf("%#v", got)
	}
}

func TestCreateProjectDefaultsAndConflict(t *testing.T) {
	s := openTest(t)
	p, err := s.CreateProject("Harbor", "harbor", "", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "planned" {
		t.Fatalf("default status %q", p.Status)
	}
	if _, err := s.CreateProject("Again", "harbor", "", "planned", nil, nil); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate slug: %v", err)
	}
	if _, err := s.CreateProject("Bad", "Harbor", "", "", nil, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("uppercase slug: %v", err)
	}
	if _, err := s.CreateProject(" ", "ok", "", "", nil, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := s.GetProject("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing project: %v", err)
	}
}

func TestProjectProgressCountsDoneAndCanceled(t *testing.T) {
	s := openTest(t)
	p, err := s.CreateProject("Dock", "dock", "", "started", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "open", Status: "todo", ProjectID: &p.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "shipped", Status: "done", ProjectID: &p.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "wont", Status: "canceled", ProjectID: &p.ID}); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetProject("dock")
	if err != nil {
		t.Fatal(err)
	}
	if got.Progress != 2.0/3.0 {
		t.Fatalf("progress %v", got.Progress)
	}
}

func TestSearchEmptyAndByIdentifier(t *testing.T) {
	s := openTest(t)
	hits, err := s.Search("")
	if err != nil {
		t.Fatal(err)
	}
	if hits == nil || len(hits) != 0 {
		t.Fatalf("empty query %#v", hits)
	}
	iss, err := s.CreateIssue(CreateIssueInput{Title: "Find me"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err = s.Search("sen-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Kind != "issue" || hits[0].ID != iss.Identifier {
		t.Fatalf("identifier search %#v", hits)
	}
	hits, err = s.Search("no-such-thing")
	if err != nil {
		t.Fatal(err)
	}
	if hits == nil || len(hits) != 0 {
		t.Fatalf("no hits %#v", hits)
	}
}

func TestUpdateWorkspaceRejectsEmptyName(t *testing.T) {
	s := openTest(t)
	empty := "  "
	if _, err := s.UpdateWorkspace(&empty, nil, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: %v", err)
	}
	tz := ""
	if _, err := s.UpdateWorkspace(nil, nil, &tz); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty timezone: %v", err)
	}
}

func TestCreateViewValidation(t *testing.T) {
	s := openTest(t)
	if _, err := s.CreateView(CreateViewInput{Name: "", Slug: "open"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty name: %v", err)
	}
	if _, err := s.CreateView(CreateViewInput{Name: "Open", Slug: "OPEN"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid slug: %v", err)
	}
	if _, err := s.CreateView(CreateViewInput{Name: "Open", Slug: "open", Display: "table"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid display: %v", err)
	}
	v, err := s.CreateView(CreateViewInput{Name: "Open", Slug: "open"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Display != "list" {
		t.Fatalf("default display %q", v.Display)
	}
	if _, err := s.CreateView(CreateViewInput{Name: "Again", Slug: "open"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate slug: %v", err)
	}
}

func TestListIssuesCycleFilter(t *testing.T) {
	s := openTest(t)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	c, err := s.CreateCycle(start, end, "active")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "in cycle", CycleID: &c.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateIssue(CreateIssueInput{Title: "out"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListIssues(IssueFilter{CycleNumber: c.Number})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Title != "in cycle" {
		t.Fatalf("%#v", got)
	}
}

func TestFirstUpcomingCycleBecomesActive(t *testing.T) {
	s := openTest(t)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	c, err := s.CreateCycle(start, end, "upcoming")
	if err != nil {
		t.Fatal(err)
	}
	if c.Number != 1 || c.Status != "active" {
		t.Fatalf("first cycle %#v", c)
	}
	if _, err := s.CreateCycle(start, end, "planned"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid cycle status: %v", err)
	}
}

func TestCommentsEmptyAndAdd(t *testing.T) {
	s := openTest(t)
	iss, err := s.CreateIssue(CreateIssueInput{Title: "note me"})
	if err != nil {
		t.Fatal(err)
	}
	cs, err := s.ListComments(iss.Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if cs == nil || len(cs) != 0 {
		t.Fatalf("empty comments %#v", cs)
	}
	if _, err := s.AddComment(iss.Identifier, "  "); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty body: %v", err)
	}
	c, err := s.AddComment(iss.Identifier, "remember this")
	if err != nil {
		t.Fatal(err)
	}
	if c.Body != "remember this" || c.IssueID != iss.ID {
		t.Fatalf("%#v", c)
	}
}

func TestCompletedAtSetOnDone(t *testing.T) {
	s := openTest(t)
	iss, err := s.CreateIssue(CreateIssueInput{Title: "ship", Status: "done"})
	if err != nil {
		t.Fatal(err)
	}
	if iss.CompletedAt == nil {
		t.Fatal("done issue needs completedAt")
	}
	todo := "todo"
	updated, err := s.UpdateIssue(iss.Identifier, PatchIssueInput{Status: &todo})
	if err != nil {
		t.Fatal(err)
	}
	if updated.CompletedAt != nil {
		t.Fatalf("reopened still completed %#v", updated.CompletedAt)
	}
}

func TestYAMLFrontmatterRejected(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	iss, err := s.CreateIssue(CreateIssueInput{Title: "yaml"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "issues", iss.Identifier+".md")
	if err := os.WriteFile(path, []byte("---\ntitle: yaml\n---\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetIssue(iss.Identifier); err == nil {
		t.Fatal("yaml frontmatter should fail")
	}
}

func TestHasUserContentIgnoresSeededLabels(t *testing.T) {
	s := openTest(t)
	has, err := s.HasUserContent()
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("seeded labels are not user content")
	}
	if _, err := s.CreatePage("ADR", "adr", "", "proposed", nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	has, err = s.HasUserContent()
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("page is user content")
	}
}

func TestDeleteProjectUnassignsIssues(t *testing.T) {
	s := openTest(t)
	p, err := s.CreateProject("Dock", "dock", "", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	iss, err := s.CreateIssue(CreateIssueInput{Title: "tied", ProjectID: &p.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteProject("dock"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetIssue(iss.Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != nil || got.ProjectSlug != nil {
		t.Fatalf("still assigned %#v %#v", got.ProjectID, got.ProjectSlug)
	}
}

func TestListPagesEmptySlice(t *testing.T) {
	s := openTest(t)
	pages, err := s.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	if pages == nil {
		t.Fatal("want empty slice, not nil")
	}
}

func TestCreatePageDefaultsAndInvalidSlug(t *testing.T) {
	s := openTest(t)
	p, err := s.CreatePage("ADR", "adr-1", "", "", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "proposed" {
		t.Fatalf("default status %q", p.Status)
	}
	if p.Tags == nil {
		t.Fatal("tags should be empty slice")
	}
	if _, err := s.CreatePage("Bad", "ADR", "", "", nil, nil, nil, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid slug: %v", err)
	}
}
