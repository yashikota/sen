package httpapi

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/yashikota/sen/internal/store"
)

type Server struct {
	store *store.Store
	dist  fs.FS
	mux   *http.ServeMux
}

func New(st *store.Store, dist fs.FS) *Server {
	s := &Server{store: st, dist: dist, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/workspace", s.getWorkspace)
	s.mux.HandleFunc("PATCH /api/workspace", s.patchWorkspace)
	s.mux.HandleFunc("GET /api/labels", s.listLabels)
	s.mux.HandleFunc("POST /api/labels", s.createLabel)
	s.mux.HandleFunc("GET /api/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/projects", s.createProject)
	s.mux.HandleFunc("GET /api/projects/{slug}", s.getProject)
	s.mux.HandleFunc("PATCH /api/projects/{slug}", s.patchProject)
	s.mux.HandleFunc("DELETE /api/projects/{slug}", s.deleteProject)
	s.mux.HandleFunc("GET /api/cycles", s.listCycles)
	s.mux.HandleFunc("POST /api/cycles", s.createCycle)
	s.mux.HandleFunc("GET /api/cycles/{number}", s.getCycle)
	s.mux.HandleFunc("PATCH /api/cycles/{number}", s.patchCycle)
	s.mux.HandleFunc("GET /api/views", s.listViews)
	s.mux.HandleFunc("POST /api/views", s.createView)
	s.mux.HandleFunc("GET /api/views/{slug}", s.getView)
	s.mux.HandleFunc("PATCH /api/views/{slug}", s.patchView)
	s.mux.HandleFunc("DELETE /api/views/{slug}", s.deleteView)
	s.mux.HandleFunc("GET /api/issues", s.listIssues)
	s.mux.HandleFunc("POST /api/issues", s.createIssue)
	s.mux.HandleFunc("GET /api/issues/{id}", s.getIssue)
	s.mux.HandleFunc("PATCH /api/issues/{id}", s.patchIssue)
	s.mux.HandleFunc("DELETE /api/issues/{id}", s.deleteIssue)
	s.mux.HandleFunc("GET /api/issues/{id}/comments", s.listComments)
	s.mux.HandleFunc("POST /api/issues/{id}/comments", s.addComment)
	s.mux.HandleFunc("GET /api/issues/{id}/activities", s.listActivities)
	s.mux.HandleFunc("GET /api/pages", s.listPages)
	s.mux.HandleFunc("POST /api/pages", s.createPage)
	s.mux.HandleFunc("GET /api/pages/{slug}", s.getPage)
	s.mux.HandleFunc("PATCH /api/pages/{slug}", s.patchPage)
	s.mux.HandleFunc("DELETE /api/pages/{slug}", s.deletePage)
	s.mux.HandleFunc("GET /api/search", s.search)
	s.mux.HandleFunc("GET /api/commands", s.commands)
	s.mux.HandleFunc("GET /api/diagnostics", s.listDiagnostics)
	s.mux.Handle("/", http.HandlerFunc(s.spa))
}

func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	if s.dist == nil || strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" {
		p = "index.html"
	}
	info, err := fs.Stat(s.dist, p)
	if err != nil || info.IsDir() {
		p = "index.html"
	}
	http.ServeFileFS(w, r, s.dist, p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, store.ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, store.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		slog.Error("api error", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (s *Server) writeWorkspace(w http.ResponseWriter, ws store.Workspace) {
	dirty, err := s.store.Dirty()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		store.Workspace
		Dirty bool `json:"dirty"`
	}{Workspace: ws, Dirty: dirty})
}

func (s *Server) getWorkspace(w http.ResponseWriter, _ *http.Request) {
	ws, err := s.store.Workspace()
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeWorkspace(w, ws)
}

func (s *Server) listDiagnostics(w http.ResponseWriter, _ *http.Request) {
	diags, err := s.store.Diagnostics()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diags)
}

func (s *Server) patchWorkspace(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name     *string `json:"name"`
		GHCRRef  *string `json:"ghcrRef"`
		Timezone *string `json:"timezone"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	ws, err := s.store.UpdateWorkspace(in.Name, in.GHCRRef, in.Timezone)
	if err != nil {
		writeError(w, err)
		return
	}
	s.writeWorkspace(w, ws)
}

func (s *Server) listLabels(w http.ResponseWriter, _ *http.Request) {
	out, err := s.store.ListLabels()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createLabel(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreateLabel(in.Name, in.Color)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listProjects(w http.ResponseWriter, _ *http.Request) {
	out, err := s.store.ListProjects()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		StartDate   *string `json:"startDate"`
		TargetDate  *string `json:"targetDate"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreateProject(in.Name, in.Slug, in.Description, in.Status, in.StartDate, in.TargetDate)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.GetProject(r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) patchProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		StartDate   *string `json:"startDate"`
		TargetDate  *string `json:"targetDate"`
		ClearStart  bool    `json:"clearStartDate"`
		ClearTarget bool    `json:"clearTargetDate"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	var start, target **string
	if in.ClearStart {
		var nils *string
		start = &nils
	} else if in.StartDate != nil {
		start = &in.StartDate
	}
	if in.ClearTarget {
		var nils *string
		target = &nils
	} else if in.TargetDate != nil {
		target = &in.TargetDate
	}
	out, err := s.store.UpdateProject(r.PathValue("slug"), in.Name, in.Description, in.Status, start, target)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteProject(r.PathValue("slug")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listCycles(w http.ResponseWriter, _ *http.Request) {
	out, err := s.store.ListCycles()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createCycle(w http.ResponseWriter, r *http.Request) {
	var in struct {
		StartsAt string `json:"startsAt"`
		EndsAt   string `json:"endsAt"`
		Status   string `json:"status"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreateCycle(in.StartsAt, in.EndsAt, in.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) getCycle(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("number"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cycle number"})
		return
	}
	out, err := s.store.GetCycle(n)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) patchCycle(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("number"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cycle number"})
		return
	}
	var in struct {
		StartsAt *string `json:"startsAt"`
		EndsAt   *string `json:"endsAt"`
		Status   *string `json:"status"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.UpdateCycle(n, in.StartsAt, in.EndsAt, in.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func viewInput(r *http.Request) (store.CreateViewInput, error) {
	var in struct {
		Name     string   `json:"name"`
		Slug     string   `json:"slug"`
		Display  string   `json:"display"`
		Status   *string  `json:"status"`
		Project  *string  `json:"project"`
		Cycle    *int     `json:"cycle"`
		Labels   []string `json:"labels"`
		Priority *int     `json:"priority"`
	}
	if err := decodeJSON(r, &in); err != nil {
		return store.CreateViewInput{}, err
	}
	return store.CreateViewInput{
		Name: in.Name, Slug: in.Slug, Display: in.Display, Status: in.Status,
		Project: in.Project, Cycle: in.Cycle, Labels: in.Labels, Priority: in.Priority,
	}, nil
}

func (s *Server) listViews(w http.ResponseWriter, _ *http.Request) {
	out, err := s.store.ListViews()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createView(w http.ResponseWriter, r *http.Request) {
	in, err := viewInput(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreateView(in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) getView(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.GetView(r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) patchView(w http.ResponseWriter, r *http.Request) {
	in, err := viewInput(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.UpdateView(r.PathValue("slug"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteView(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteView(r.PathValue("slug")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listIssues(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.IssueFilter{Status: q.Get("status"), ProjectSlug: q.Get("project")}
	if c := q.Get("cycle"); c != "" {
		n, err := strconv.Atoi(c)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cycle"})
			return
		}
		f.CycleNumber = n
	}
	if p := q.Get("priority"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid priority"})
			return
		}
		f.Priority = &n
	}
	if labels := q.Get("labels"); labels != "" {
		for _, name := range strings.Split(labels, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				f.Labels = append(f.Labels, name)
			}
		}
	}
	out, err := s.store.ListIssues(f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createIssue(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title     string  `json:"title"`
		Body      string  `json:"body"`
		Status    string  `json:"status"`
		Priority  int     `json:"priority"`
		ProjectID *int64  `json:"projectId"`
		CycleID   *int64  `json:"cycleId"`
		ParentID  *int64  `json:"parentId"`
		DueDate   *string `json:"dueDate"`
		LabelIDs  []int64 `json:"labelIds"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreateIssue(store.CreateIssueInput{
		Title: in.Title, Body: in.Body, Status: in.Status, Priority: in.Priority,
		ProjectID: in.ProjectID, CycleID: in.CycleID, ParentID: in.ParentID, DueDate: in.DueDate, LabelIDs: in.LabelIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) getIssue(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.GetIssue(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) patchIssue(w http.ResponseWriter, r *http.Request) {
	raw := map[string]json.RawMessage{}
	if err := decodeJSON(r, &raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	in := store.PatchIssueInput{}
	if v, ok := raw["title"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid title"})
			return
		}
		in.Title = &s
	}
	if v, ok := raw["body"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}
		in.Body = &s
	}
	if v, ok := raw["status"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
			return
		}
		in.Status = &s
	}
	if v, ok := raw["priority"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid priority"})
			return
		}
		in.Priority = &n
	}
	if v, ok := raw["projectId"]; ok {
		id, err := unmarshalOptInt64(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid projectId"})
			return
		}
		in.ProjectID = &id
	}
	if v, ok := raw["cycleId"]; ok {
		id, err := unmarshalOptInt64(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cycleId"})
			return
		}
		in.CycleID = &id
	}
	if v, ok := raw["parentId"]; ok {
		id, err := unmarshalOptInt64(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid parentId"})
			return
		}
		in.ParentID = &id
	}
	if v, ok := raw["dueDate"]; ok {
		d, err := unmarshalOptString(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid dueDate"})
			return
		}
		in.DueDate = &d
	}
	if v, ok := raw["labelIds"]; ok {
		var ids []int64
		if err := json.Unmarshal(v, &ids); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid labelIds"})
			return
		}
		in.LabelIDs = &ids
	}
	if v, ok := raw["sortOrder"]; ok {
		var n float64
		if err := json.Unmarshal(v, &n); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid sortOrder"})
			return
		}
		in.SortOrder = &n
	}
	out, err := s.store.UpdateIssue(r.PathValue("id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteIssue(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteIssue(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.ListComments(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) addComment(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.AddComment(r.PathValue("id"), in.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listActivities(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.ListActivities(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listPages(w http.ResponseWriter, _ *http.Request) {
	out, err := s.store.ListPages()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createPage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title     string   `json:"title"`
		Slug      string   `json:"slug"`
		Body      string   `json:"body"`
		Status    string   `json:"status"`
		ParentID  *int64   `json:"parentId"`
		ProjectID *int64   `json:"projectId"`
		Date      *string  `json:"date"`
		Tags      []string `json:"tags"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	out, err := s.store.CreatePage(in.Title, in.Slug, in.Body, in.Status, in.ParentID, in.ProjectID, in.Date, in.Tags)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) getPage(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.GetPage(r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) patchPage(w http.ResponseWriter, r *http.Request) {
	raw := map[string]json.RawMessage{}
	if err := decodeJSON(r, &raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	var title, body, status *string
	var parentID, projectID **int64
	var date **string
	var tags *[]string
	if v, ok := raw["title"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid title"})
			return
		}
		title = &s
	}
	if v, ok := raw["body"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}
		body = &s
	}
	if v, ok := raw["status"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
			return
		}
		status = &s
	}
	if v, ok := raw["parentId"]; ok {
		id, err := unmarshalOptInt64(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid parentId"})
			return
		}
		parentID = &id
	}
	if v, ok := raw["projectId"]; ok {
		id, err := unmarshalOptInt64(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid projectId"})
			return
		}
		projectID = &id
	}
	if v, ok := raw["date"]; ok {
		d, err := unmarshalOptString(v)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date"})
			return
		}
		date = &d
	}
	if v, ok := raw["tags"]; ok {
		var t []string
		if err := json.Unmarshal(v, &t); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tags"})
			return
		}
		tags = &t
	}
	out, err := s.store.UpdatePage(r.PathValue("slug"), title, body, status, parentID, projectID, date, tags)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deletePage(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeletePage(r.PathValue("slug")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.Search(r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) commands(w http.ResponseWriter, _ *http.Request) {
	cmds := []map[string]string{
		{"id": "new-issue", "title": "Create issue", "hint": "c"},
		{"id": "new-page", "title": "Create page", "hint": "p"},
		{"id": "new-view", "title": "Create view", "hint": ""},
		{"id": "goto-issues", "title": "Go to Issues", "hint": ""},
		{"id": "goto-board", "title": "Go to Board", "hint": ""},
		{"id": "goto-projects", "title": "Go to Projects", "hint": ""},
		{"id": "goto-cycles", "title": "Go to Cycles", "hint": ""},
		{"id": "goto-pages", "title": "Go to Pages", "hint": ""},
		{"id": "set-status-backlog", "title": "Set status: Backlog", "hint": "s"},
		{"id": "set-status-todo", "title": "Set status: Todo", "hint": "s"},
		{"id": "set-status-in_progress", "title": "Set status: In Progress", "hint": "s"},
		{"id": "set-status-done", "title": "Set status: Done", "hint": "s"},
		{"id": "set-status-canceled", "title": "Set status: Canceled", "hint": "s"},
	}
	cycles, err := s.store.ListCycles()
	if err == nil {
		for _, c := range cycles {
			cmds = append(cmds, map[string]string{
				"id":    "assign-cycle:" + strconv.FormatInt(c.ID, 10),
				"title": "Assign to Cycle " + strconv.Itoa(c.Number),
				"hint":  "",
			})
		}
		cmds = append(cmds, map[string]string{
			"id":    "assign-cycle:none",
			"title": "Remove from cycle",
			"hint":  "",
		})
	}
	writeJSON(w, http.StatusOK, cmds)
}

func unmarshalOptInt64(raw json.RawMessage) (*int64, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

func unmarshalOptString(raw json.RawMessage) (*string, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
