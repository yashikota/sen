package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestInvalidIssueQueryParams(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/issues?cycle=abc", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cycle %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/issues?priority=high", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("priority %d %s", rec.Code, rec.Body.String())
	}
}

func TestCommandsAndSearch(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/commands", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("commands %d %s", rec.Code, rec.Body.String())
	}
	var cmds []map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &cmds); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, c := range cmds {
		ids[c["id"]] = true
	}
	if !ids["new-issue"] || !ids["new-view"] || !ids["assign-cycle:none"] {
		t.Fatalf("commands %#v", cmds)
	}

	rec = doJSON(t, s, "GET", "/api/search?q=", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("empty search %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"Findable"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/search?q=Findable", "")
	var hits []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &hits); err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0]["kind"] != "issue" {
		t.Fatalf("search %#v", hits)
	}
}

func TestMissingResourcesAndEmptyLists(t *testing.T) {
	s := testAPI(t)
	if rec := doJSON(t, s, "GET", "/api/pages/missing", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("page %d", rec.Code)
	}
	if rec := doJSON(t, s, "GET", "/api/projects/missing", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("project %d", rec.Code)
	}
	if rec := doJSON(t, s, "GET", "/api/cycles/9", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("cycle %d", rec.Code)
	}
	if rec := doJSON(t, s, "GET", "/api/projects", ""); rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("projects %s", rec.Body.String())
	}
	if rec := doJSON(t, s, "GET", "/api/cycles", ""); rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("cycles %s", rec.Body.String())
	}
	if rec := doJSON(t, s, "GET", "/api/pages", ""); rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("pages %s", rec.Body.String())
	}
}

func TestPatchInvalidStatusAndUnknownAPI(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"x"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	id := issue["identifier"].(string)
	rec = doJSON(t, s, "PATCH", "/api/issues/"+id, `{"status":"ready"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/nope", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown api %d", rec.Code)
	}
	rec = doJSON(t, s, "GET", "/issues", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("spa without dist %d", rec.Code)
	}
}

func TestCommentsEmptyJSONArray(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"x"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	id := issue["identifier"].(string)
	rec = doJSON(t, s, "GET", "/api/issues/"+id+"/comments", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("comments %d %s", rec.Code, rec.Body.String())
	}
}

func TestDuplicateLabelAndProjectConflict(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/labels", `{"name":"Bug","color":"#ffffff"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("label conflict %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, s, "POST", "/api/projects", `{"name":"A","slug":"dock"}`); rec.Code != http.StatusCreated {
		t.Fatalf("project %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "POST", "/api/projects", `{"name":"B","slug":"dock"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("project conflict %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreateIssueOmitsStatusDefaultsBacklog(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"from api"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	if issue["status"] != "backlog" {
		t.Fatalf("status %#v", issue["status"])
	}
}

func TestWorkspaceEmptyNameRejected(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "PATCH", "/api/workspace", `{"name":"  "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty name %d %s", rec.Code, rec.Body.String())
	}
}
