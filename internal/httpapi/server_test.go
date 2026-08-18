package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
