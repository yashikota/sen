package domain

import (
	"strings"
	"testing"
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
