package domain

import (
	"strings"
	"testing"
	"time"
)

func TestIdentifierRoundTrip(t *testing.T) {
	t.Parallel()
	id := Identifier(12)
	if id != "SEN-12" {
		t.Fatalf("got %q", id)
	}
	n, ok := ParseIdentifier(id)
	if !ok || n != 12 {
		t.Fatalf("parse %q: %d %v", id, n, ok)
	}
	if _, ok := ParseIdentifier("SEN-0"); ok {
		t.Fatal("SEN-0 should be rejected")
	}
	if _, ok := ParseIdentifier("sen-1"); ok {
		t.Fatal("lowercase prefix should be rejected")
	}
}

func TestIsDirty(t *testing.T) {
	t.Parallel()
	pushed := "2026-08-18T00:00:00Z"
	if IsDirty("2026-08-18T00:00:00Z", nil, false) {
		t.Fatal("fresh init without user content is not dirty")
	}
	if !IsDirty("2026-08-18T00:00:00Z", nil, true) {
		t.Fatal("unpushed user content is dirty")
	}
	if IsDirty("2026-08-18T00:00:00Z", &pushed, true) {
		t.Fatal("equal timestamps are not dirty")
	}
	if !IsDirty("2026-08-18T00:00:01Z", &pushed, true) {
		t.Fatal("later updated_at is dirty")
	}
}

func TestUniqueSlug(t *testing.T) {
	t.Parallel()
	used := map[string]struct{}{}
	if g := UniqueSlug(used, "adr"); g != "adr" {
		t.Fatalf("got %q", g)
	}
	if g := UniqueSlug(used, "adr"); g != "adr-2" {
		t.Fatalf("got %q", g)
	}
	if g := UniqueSlug(used, "adr"); g != "adr-3" {
		t.Fatalf("got %q", g)
	}
}

func TestRenderPageMarkdown(t *testing.T) {
	t.Parallel()
	date := "2026-08-18"
	project := "sen"
	md := RenderPageMarkdown(Page{
		Title:       "SQLite as source of truth",
		Slug:        "sqlite-source",
		Body:        "We keep SQLite as the only source of truth.\n",
		Status:      "proposed",
		Date:        &date,
		Tags:        []string{"adr"},
		ProjectSlug: &project,
	})
	if !strings.HasPrefix(md, "---\n") {
		t.Fatal("missing frontmatter start")
	}
	for _, want := range []string{
		"title: SQLite as source of truth\n",
		"slug: sqlite-source\n",
		"status: proposed\n",
		"date: 2026-08-18\n",
		"tags: [adr]\n",
		"project: sen\n",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("missing %q in %s", want, md)
		}
	}
	if strings.Contains(md, "id:") {
		t.Fatal("internal ids must not appear")
	}
}

func TestValidSlug(t *testing.T) {
	t.Parallel()
	if !ValidSlug("sqlite-source") {
		t.Fatal("expected valid")
	}
	if ValidSlug("-x") || ValidSlug("X") || ValidSlug("") {
		t.Fatal("expected invalid")
	}
}

func TestValidators(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"backlog", "todo", "in_progress", "done", "canceled"} {
		if !ValidIssueStatus(s) {
			t.Fatalf("issue status %q", s)
		}
	}
	if ValidIssueStatus("ready") || ValidIssueStatus("") {
		t.Fatal("invalid issue status accepted")
	}
	if !ValidPriority(0) || !ValidPriority(4) || ValidPriority(-1) || ValidPriority(5) {
		t.Fatal("priority bounds")
	}
	for _, s := range []string{"planned", "started", "completed", "canceled"} {
		if !ValidProjectStatus(s) {
			t.Fatalf("project status %q", s)
		}
	}
	if ValidProjectStatus("active") {
		t.Fatal("project must not use cycle statuses")
	}
	for _, s := range []string{"upcoming", "active", "completed"} {
		if !ValidCycleStatus(s) {
			t.Fatalf("cycle status %q", s)
		}
	}
	if ValidCycleStatus("planned") {
		t.Fatal("cycle must not use project statuses")
	}
	for _, s := range []string{"proposed", "accepted", "deprecated", "superseded"} {
		if !ValidPageStatus(s) {
			t.Fatalf("page status %q", s)
		}
	}
	if ValidPageStatus("draft") {
		t.Fatal("draft is not a page status")
	}
	if !ValidViewDisplay("list") || !ValidViewDisplay("board") || ValidViewDisplay("table") {
		t.Fatal("view display")
	}
}

func TestParseIdentifierRejectsJunk(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"", "SEN-", "SEN-abc", "ISSUE-1", "SEN-1a", "SEN--1"} {
		if _, ok := ParseIdentifier(id); ok {
			t.Fatalf("accepted %q", id)
		}
	}
	n, ok := ParseIdentifier("SEN-01")
	if !ok || n != 1 {
		t.Fatalf("SEN-01: %d %v", n, ok)
	}
}

func TestIsDirtyEmptyLastPush(t *testing.T) {
	t.Parallel()
	empty := ""
	if IsDirty("2026-08-18T00:00:00Z", &empty, false) {
		t.Fatal("empty last push without content is not dirty")
	}
	if !IsDirty("2026-08-18T00:00:00Z", &empty, true) {
		t.Fatal("empty last push with content is dirty")
	}
}

func TestUniqueSlugSkipsOccupiedSuffix(t *testing.T) {
	t.Parallel()
	used := map[string]struct{}{"adr": {}, "adr-2": {}}
	if g := UniqueSlug(used, "adr"); g != "adr-3" {
		t.Fatalf("got %q", g)
	}
}

func TestRenderPageMarkdownQuotesAndParent(t *testing.T) {
	t.Parallel()
	parent := "root-adr"
	md := RenderPageMarkdown(Page{
		Title:      "Foo: bar",
		Slug:       "foo-bar",
		Body:       "line",
		Status:     "accepted",
		Tags:       nil,
		ParentSlug: &parent,
	})
	if !strings.Contains(md, `title: "Foo: bar"`) {
		t.Fatalf("quoted title missing: %s", md)
	}
	if !strings.Contains(md, "parent: root-adr\n") {
		t.Fatalf("parent missing: %s", md)
	}
	if !strings.HasSuffix(md, "line\n") {
		t.Fatalf("body newline: %q", md)
	}
}

func TestNowIsRFC3339UTC(t *testing.T) {
	t.Parallel()
	got := Now()
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Location() != time.UTC {
		t.Fatalf("location %v", parsed.Location())
	}
}
