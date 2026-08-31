package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yashikota/sen/internal/store"
)

func testAPI(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st, nil)
}

func doJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestListIssuesEmptyJSONArray(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/issues", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("empty list want [], got %s", rec.Body.String())
	}
}

func TestCreateIssueRequiresTitle(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code %d body %s", rec.Code, rec.Body.String())
	}
}

func TestGetMissingIssue(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/issues/SEN-99", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestScenarioIssueCommentPage(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"First"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create issue %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	id := issue["identifier"].(string)
	rec = doJSON(t, s, "POST", "/api/issues/"+id+"/comments", `{"body":"looks good"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("comment %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "PATCH", "/api/issues/"+id, `{"status":"in_progress"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "POST", "/api/pages", `{"title":"ADR 1","slug":"adr-1","body":"# Decision\n","status":"proposed"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("page %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/issues/"+id+"/activities", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("activities %d", rec.Code)
	}
	var acts []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &acts); err != nil {
		t.Fatal(err)
	}
	if len(acts) < 2 {
		t.Fatalf("want created+commented+status, got %d", len(acts))
	}
}

func TestIssueLabelsDueAndDiagnostics(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/diagnostics", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("empty diagnostics %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/labels", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("labels %d %s", rec.Code, rec.Body.String())
	}
	var labels []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &labels); err != nil {
		t.Fatal(err)
	}
	if len(labels) < 1 {
		t.Fatal("want seeded labels")
	}
	id := int64(labels[0]["id"].(float64))
	rec = doJSON(t, s, "POST", "/api/issues", `{"title":"tagged"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	ident := issue["identifier"].(string)
	body := `{"labelIds":[` + strconv.FormatInt(id, 10) + `],"dueDate":"2026-09-01"}`
	rec = doJSON(t, s, "PATCH", "/api/issues/"+ident, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	gotLabels, _ := issue["labels"].([]any)
	if len(gotLabels) != 1 {
		t.Fatalf("labels %#v", issue["labels"])
	}
	if issue["dueDate"] != "2026-09-01" {
		t.Fatalf("dueDate %#v", issue["dueDate"])
	}
}

func TestIssueParentAndListDepth(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"root"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create parent %d %s", rec.Code, rec.Body.String())
	}
	var parent map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parent); err != nil {
		t.Fatal(err)
	}
	parentID := int64(parent["id"].(float64))
	body := `{"title":"leaf","parentId":` + strconv.FormatInt(parentID, 10) + `}`
	rec = doJSON(t, s, "POST", "/api/issues", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create child %d %s", rec.Code, rec.Body.String())
	}
	var child map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &child); err != nil {
		t.Fatal(err)
	}
	if child["parentIdentifier"] != "SEN-1" {
		t.Fatalf("parentIdentifier %#v", child["parentIdentifier"])
	}
	rec = doJSON(t, s, "GET", "/api/issues", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var issues []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("want 2 issues, got %d", len(issues))
	}
	if issues[0]["identifier"] != "SEN-1" || issues[1]["identifier"] != "SEN-2" {
		t.Fatalf("tree order %#v %#v", issues[0]["identifier"], issues[1]["identifier"])
	}
	if issues[1]["depth"].(float64) != 1 {
		t.Fatalf("child depth %#v", issues[1]["depth"])
	}
}

func TestViewCRUDAndIssueFilter(t *testing.T) {
	s := testAPI(t)
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"open","status":"todo"}`); rec.Code != http.StatusCreated {
		t.Fatalf("todo issue %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"closed","status":"done"}`); rec.Code != http.StatusCreated {
		t.Fatalf("done issue %d %s", rec.Code, rec.Body.String())
	}
	rec := doJSON(t, s, "POST", "/api/views", `{"name":"Open","slug":"open","display":"list","status":"todo"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create view %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/views/open", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get view %d %s", rec.Code, rec.Body.String())
	}
	var view map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view["name"] != "Open" || view["status"] != "todo" {
		t.Fatalf("view %#v", view)
	}
	rec = doJSON(t, s, "GET", "/api/issues?status=todo", "")
	var issues []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0]["title"] != "open" {
		t.Fatalf("filtered %#v", issues)
	}
	rec = doJSON(t, s, "PATCH", "/api/views/open", `{"name":"Todos","display":"board"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch view %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/views", "")
	var views []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &views); err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || views[0]["name"] != "Todos" || views[0]["display"] != "board" {
		t.Fatalf("list views %#v", views)
	}
	rec = doJSON(t, s, "DELETE", "/api/views/open", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete view %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/views/open", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("deleted view %d", rec.Code)
	}
}

func TestListIssuesPriorityQuery(t *testing.T) {
	s := testAPI(t)
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"hot","priority":1}`); rec.Code != http.StatusCreated {
		t.Fatalf("hot %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"cold","priority":4}`); rec.Code != http.StatusCreated {
		t.Fatalf("cold %d %s", rec.Code, rec.Body.String())
	}
	rec := doJSON(t, s, "GET", "/api/issues?priority=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var issues []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0]["title"] != "hot" {
		t.Fatalf("got %#v", issues)
	}
}

func TestUnknownJSONRejected(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "POST", "/api/issues", `{"title":"x","assignee":"me"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown field %d %s", rec.Code, rec.Body.String())
	}
}

func TestLivedInWorkspaceHTTP(t *testing.T) {
	s := testAPI(t)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	rec := doJSON(t, s, "POST", "/api/projects", `{"name":"Atlas","slug":"atlas"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("project %d %s", rec.Code, rec.Body.String())
	}
	var project map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	rec = doJSON(t, s, "POST", "/api/cycles", `{"startsAt":"`+start+`","endsAt":"`+end+`","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("cycle %d %s", rec.Code, rec.Body.String())
	}
	var cycle map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &cycle); err != nil {
		t.Fatal(err)
	}
	rec = doJSON(t, s, "GET", "/api/labels", "")
	var labels []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &labels); err != nil {
		t.Fatal(err)
	}
	labelID := int64(labels[0]["id"].(float64))
	body := `{"title":"Ship","status":"todo","priority":1,"projectId":` + strconv.FormatInt(int64(project["id"].(float64)), 10) + `,"cycleId":` + strconv.FormatInt(int64(cycle["id"].(float64)), 10) + `,"labelIds":[` + strconv.FormatInt(labelID, 10) + `]}`
	rec = doJSON(t, s, "POST", "/api/issues", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("epic %d %s", rec.Code, rec.Body.String())
	}
	var epic map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &epic); err != nil {
		t.Fatal(err)
	}
	childBody := `{"title":"Docs","parentId":` + strconv.FormatInt(int64(epic["id"].(float64)), 10) + `,"status":"todo","projectId":` + strconv.FormatInt(int64(project["id"].(float64)), 10) + `}`
	rec = doJSON(t, s, "POST", "/api/issues", childBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("child %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, s, "POST", "/api/issues", `{"title":"Old","status":"done"}`); rec.Code != http.StatusCreated {
		t.Fatalf("done %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "POST", "/api/views", `{"name":"Atlas todo","slug":"atlas-todo","display":"list","status":"todo","project":"atlas"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("view %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/issues", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list all %d %s", rec.Code, rec.Body.String())
	}
	var issues []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 3 {
		t.Fatalf("all issues %#v", issues)
	}
	rec = doJSON(t, s, "GET", "/api/issues?status=todo&project=atlas", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("atlas todo %#v", issues)
	}
	if issues[0]["identifier"] != "SEN-1" || issues[1]["depth"].(float64) != 1 {
		t.Fatalf("tree %#v", issues)
	}
	rec = doJSON(t, s, "GET", "/api/issues?labels="+labels[0]["name"].(string), "")
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0]["title"] != "Ship" {
		t.Fatalf("label filter %#v", issues)
	}
	rec = doJSON(t, s, "GET", "/api/search?q=atlas", "")
	var hits []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &hits); err != nil {
		t.Fatal(err)
	}
	kinds := map[string]bool{}
	for _, h := range hits {
		kinds[h["kind"].(string)] = true
	}
	if !kinds["project"] || !kinds["view"] {
		t.Fatalf("search kinds %#v", hits)
	}
	rec = doJSON(t, s, "GET", "/api/views", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("views %d %s", rec.Code, rec.Body.String())
	}
}

func TestListViewsEmptyJSONArray(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/views", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("empty views want [], got %s", rec.Body.String())
	}
}

func TestCreateCycle(t *testing.T) {
	s := testAPI(t)
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"startsAt":"` + start + `","endsAt":"` + end + `"}`
	rec := doJSON(t, s, "POST", "/api/cycles", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
}

func TestLabelTimezoneSortOrderAndDeletes(t *testing.T) {
	s := testAPI(t)

	rec := doJSON(t, s, "PATCH", "/api/workspace", `{"timezone":"Asia/Tokyo"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("timezone %d %s", rec.Code, rec.Body.String())
	}
	var ws map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &ws); err != nil {
		t.Fatal(err)
	}
	if ws["timezone"] != "Asia/Tokyo" {
		t.Fatalf("timezone %#v", ws["timezone"])
	}

	rec = doJSON(t, s, "POST", "/api/labels", `{"name":"Harbor","color":"#6b9bd1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("label %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, s, "POST", "/api/projects", `{"name":"Dock","slug":"dock"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("project %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, s, "POST", "/api/issues", `{"title":"Move me","status":"todo"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("issue %d %s", rec.Code, rec.Body.String())
	}
	var issue map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	id := issue["identifier"].(string)
	rec = doJSON(t, s, "PATCH", "/api/issues/"+id, `{"status":"in_progress","sortOrder":4.5}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("sort %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &issue); err != nil {
		t.Fatal(err)
	}
	if issue["status"] != "in_progress" {
		t.Fatalf("status %#v", issue["status"])
	}
	if issue["sortOrder"] != 4.5 {
		t.Fatalf("sortOrder %#v", issue["sortOrder"])
	}
	rec = doJSON(t, s, "DELETE", "/api/issues/"+id, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete issue %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, s, "POST", "/api/views", `{"name":"Todo","slug":"todo","display":"list","status":"todo"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("view %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "DELETE", "/api/views/todo", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete view %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, s, "POST", "/api/pages", `{"title":"ADR","slug":"adr-1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("page %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "DELETE", "/api/pages/adr-1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete page %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, s, "DELETE", "/api/projects/dock", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete project %d %s", rec.Code, rec.Body.String())
	}
}

func TestWorkspaceDirtyAfterUserContent(t *testing.T) {
	s := testAPI(t)
	rec := doJSON(t, s, "GET", "/api/workspace", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d %s", rec.Code, rec.Body.String())
	}
	var ws map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &ws); err != nil {
		t.Fatal(err)
	}
	if ws["dirty"] != false {
		t.Fatalf("empty workspace dirty %#v", ws["dirty"])
	}
	rec = doJSON(t, s, "POST", "/api/issues", `{"title":"Work"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("issue %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, s, "GET", "/api/workspace", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &ws); err != nil {
		t.Fatal(err)
	}
	if ws["dirty"] != true {
		t.Fatalf("after create dirty %#v", ws["dirty"])
	}
}
