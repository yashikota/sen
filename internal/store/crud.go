package store

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/yashikota/sen/internal/domain"
)

func (s *Store) Workspace() (Workspace, error) {
	var ws Workspace
	err := s.snapshot(func(m *mem) error {
		ws = workspaceFrom(m)
		return nil
	})
	return ws, err
}

func workspaceFrom(m *mem) Workspace {
	return Workspace{
		Name:             m.Workspace.Name,
		GHCRRef:          m.Workspace.GHCRRef,
		Timezone:         m.Workspace.Timezone,
		LastPushedAt:     m.Workspace.LastPushedAt,
		LastPushedDigest: m.Workspace.LastPushedDigest,
		UpdatedAt:        m.Workspace.UpdatedAt,
	}
}

func (s *Store) UpdateWorkspace(name, ghcrRef, timezone *string) (Workspace, error) {
	var ws Workspace
	err := s.mutate(func(m *mem) error {
		if name != nil {
			if strings.TrimSpace(*name) == "" {
				return validationf("name required")
			}
			m.Workspace.Name = strings.TrimSpace(*name)
		}
		if ghcrRef != nil {
			m.Workspace.GHCRRef = strings.TrimSpace(*ghcrRef)
		}
		if timezone != nil {
			if strings.TrimSpace(*timezone) == "" {
				return validationf("timezone required")
			}
			m.Workspace.Timezone = strings.TrimSpace(*timezone)
		}
		m.bump(domain.Now())
		ws = workspaceFrom(m)
		return nil
	})
	return ws, err
}

func (s *Store) MarkPushed(digest string) error {
	return s.mutate(func(m *mem) error {
		now := domain.Now()
		m.Workspace.LastPushedAt = &now
		m.Workspace.LastPushedDigest = &digest
		m.Workspace.UpdatedAt = now
		return nil
	})
}

func (s *Store) HasUserContent() (bool, error) {
	var n int
	err := s.snapshot(func(m *mem) error {
		n = len(m.Issues) + len(m.Projects) + len(m.Cycles) + len(m.Pages)
		for _, cs := range m.Comments {
			n += len(cs)
		}
		return nil
	})
	return n > 0, err
}

func (s *Store) Dirty() (bool, error) {
	ws, err := s.Workspace()
	if err != nil {
		return false, err
	}
	has, err := s.HasUserContent()
	if err != nil {
		return false, err
	}
	return domain.IsDirty(ws.UpdatedAt, ws.LastPushedAt, has), nil
}

func (s *Store) ListLabels() ([]Label, error) {
	var out []Label
	err := s.snapshot(func(m *mem) error {
		out = append([]Label{}, m.Labels...)
		return nil
	})
	return out, err
}

func (s *Store) CreateLabel(name, color string) (Label, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Label{}, validationf("name required")
	}
	if !validColor(color) {
		return Label{}, validationf("color must be #RRGGBB")
	}
	var out Label
	err := s.mutate(func(m *mem) error {
		if _, ok := labelByName(m, name); ok {
			return errf(ErrConflict, "label name")
		}
		out = Label{ID: m.nextID(), Name: name, Color: color}
		m.Labels = append(m.Labels, out)
		m.bump(domain.Now())
		return nil
	})
	return out, err
}

func (s *Store) ListProjects() ([]Project, error) {
	var out []Project
	err := s.snapshot(func(m *mem) error {
		out = make([]Project, len(m.Projects))
		copy(out, m.Projects)
		for i := range out {
			out[i].Progress = projectProgress(m, out[i].ID)
		}
		return nil
	})
	return out, err
}

func (s *Store) GetProject(slug string) (Project, error) {
	var p Project
	err := s.snapshot(func(m *mem) error {
		got, ok := projectBySlug(m, slug)
		if !ok {
			return ErrNotFound
		}
		got.Progress = projectProgress(m, got.ID)
		p = got
		return nil
	})
	return p, err
}

func (s *Store) CreateProject(name, slug, description, status string, start, target *string) (Project, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return Project{}, validationf("name required")
	}
	if !domain.ValidSlug(slug) {
		return Project{}, validationf("invalid slug")
	}
	if status == "" {
		status = "planned"
	}
	if !domain.ValidProjectStatus(status) {
		return Project{}, validationf("invalid status")
	}
	now := domain.Now()
	var out Project
	err := s.mutate(func(m *mem) error {
		if _, ok := projectBySlug(m, slug); ok {
			return errf(ErrConflict, "slug")
		}
		out = Project{
			ID: m.nextID(), Name: name, Slug: slug, Description: description, Status: status,
			StartDate: start, TargetDate: target, CreatedAt: now, UpdatedAt: now,
		}
		m.Projects = append(m.Projects, out)
		addActivity(m, "project", out.ID, "created", map[string]any{"slug": slug}, now)
		m.bump(now)
		return nil
	})
	return out, err
}

func (s *Store) UpdateProject(slug string, name, description, status *string, start, target **string) (Project, error) {
	var out Project
	err := s.mutate(func(m *mem) error {
		i := indexProject(m, slug)
		if i < 0 {
			return ErrNotFound
		}
		p := m.Projects[i]
		if name != nil {
			if strings.TrimSpace(*name) == "" {
				return validationf("name required")
			}
			p.Name = strings.TrimSpace(*name)
		}
		if description != nil {
			p.Description = *description
		}
		if status != nil {
			if !domain.ValidProjectStatus(*status) {
				return validationf("invalid status")
			}
			p.Status = *status
		}
		if start != nil {
			p.StartDate = *start
		}
		if target != nil {
			p.TargetDate = *target
		}
		now := domain.Now()
		p.UpdatedAt = now
		m.Projects[i] = p
		m.bump(now)
		p.Progress = projectProgress(m, p.ID)
		out = p
		return nil
	})
	return out, err
}

func (s *Store) DeleteProject(slug string) error {
	return s.mutate(func(m *mem) error {
		i := indexProject(m, slug)
		if i < 0 {
			return ErrNotFound
		}
		id := m.Projects[i].ID
		m.Projects = append(m.Projects[:i], m.Projects[i+1:]...)
		for j := range m.Issues {
			if m.Issues[j].ProjectID != nil && *m.Issues[j].ProjectID == id {
				m.Issues[j].ProjectID = nil
				m.Issues[j].ProjectSlug = nil
			}
		}
		m.bump(domain.Now())
		return nil
	})
}

func (s *Store) ListCycles() ([]Cycle, error) {
	var out []Cycle
	err := s.snapshot(func(m *mem) error {
		out = append([]Cycle{}, m.Cycles...)
		return nil
	})
	return out, err
}

func (s *Store) GetCycle(number int) (Cycle, error) {
	var c Cycle
	err := s.snapshot(func(m *mem) error {
		got, ok := cycleByNumber(m, number)
		if !ok {
			return ErrNotFound
		}
		c = got
		return nil
	})
	return c, err
}

func (s *Store) CreateCycle(startsAt, endsAt, status string) (Cycle, error) {
	if err := parseTime(startsAt); err != nil {
		return Cycle{}, err
	}
	if err := parseTime(endsAt); err != nil {
		return Cycle{}, err
	}
	if status == "" {
		status = "upcoming"
	}
	if !domain.ValidCycleStatus(status) {
		return Cycle{}, validationf("invalid status")
	}
	now := domain.Now()
	var out Cycle
	err := s.mutate(func(m *mem) error {
		next := 1
		for _, c := range m.Cycles {
			if c.Number >= next {
				next = c.Number + 1
			}
		}
		if next == 1 && status == "upcoming" {
			status = "active"
		}
		ensureSingleActive(m, 0, status)
		out = Cycle{ID: m.nextID(), Number: next, StartsAt: startsAt, EndsAt: endsAt, Status: status, CreatedAt: now, UpdatedAt: now}
		m.Cycles = append(m.Cycles, out)
		addActivity(m, "cycle", out.ID, "created", map[string]any{"number": next}, now)
		m.bump(now)
		return nil
	})
	return out, err
}

func (s *Store) UpdateCycle(number int, startsAt, endsAt, status *string) (Cycle, error) {
	var out Cycle
	err := s.mutate(func(m *mem) error {
		i := indexCycle(m, number)
		if i < 0 {
			return ErrNotFound
		}
		c := m.Cycles[i]
		if startsAt != nil {
			if err := parseTime(*startsAt); err != nil {
				return err
			}
			c.StartsAt = *startsAt
		}
		if endsAt != nil {
			if err := parseTime(*endsAt); err != nil {
				return err
			}
			c.EndsAt = *endsAt
		}
		if status != nil {
			if !domain.ValidCycleStatus(*status) {
				return validationf("invalid status")
			}
			c.Status = *status
		}
		now := domain.Now()
		ensureSingleActive(m, c.ID, c.Status)
		c.UpdatedAt = now
		m.Cycles[i] = c
		m.bump(now)
		out = c
		return nil
	})
	return out, err
}

func ensureSingleActive(m *mem, id int64, status string) {
	if status != "active" {
		return
	}
	now := domain.Now()
	for i := range m.Cycles {
		if m.Cycles[i].Status == "active" && m.Cycles[i].ID != id {
			m.Cycles[i].Status = "completed"
			m.Cycles[i].UpdatedAt = now
		}
	}
}

func (s *Store) ListIssues(f IssueFilter) ([]Issue, error) {
	var out []Issue
	err := s.snapshot(func(m *mem) error {
		for _, iss := range m.Issues {
			if f.Status != "" && iss.Status != f.Status {
				continue
			}
			if f.ProjectSlug != "" && (iss.ProjectSlug == nil || *iss.ProjectSlug != f.ProjectSlug) {
				if iss.ProjectID != nil {
					if p, ok := projectByID(m, *iss.ProjectID); !ok || p.Slug != f.ProjectSlug {
						continue
					}
				} else {
					continue
				}
			}
			if f.CycleNumber != 0 {
				ok := iss.CycleNumber != nil && *iss.CycleNumber == f.CycleNumber
				if !ok && iss.CycleID != nil {
					if c, found := cycleByID(m, *iss.CycleID); found && c.Number == f.CycleNumber {
						ok = true
					}
				}
				if !ok {
					continue
				}
			}
			out = append(out, iss)
		}
		if out == nil {
			out = []Issue{}
		}
		return nil
	})
	return out, err
}

func (s *Store) GetIssue(identifier string) (Issue, error) {
	var iss Issue
	err := s.snapshot(func(m *mem) error {
		got, ok := issueByIdent(m, identifier)
		if !ok {
			return ErrNotFound
		}
		iss = got
		return nil
	})
	return iss, err
}

func (s *Store) CreateIssue(in CreateIssueInput) (Issue, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return Issue{}, validationf("title required")
	}
	if in.Status == "" {
		in.Status = "backlog"
	}
	if !domain.ValidIssueStatus(in.Status) {
		return Issue{}, validationf("invalid status")
	}
	if !domain.ValidPriority(in.Priority) {
		return Issue{}, validationf("invalid priority")
	}
	now := domain.Now()
	var out Issue
	err := s.mutate(func(m *mem) error {
		m.Workspace.IssueCounter++
		n := m.Workspace.IssueCounter
		ident := domain.Identifier(n)
		sort := 1.0
		for _, iss := range m.Issues {
			if iss.SortOrder >= sort {
				sort = iss.SortOrder + 1
			}
		}
		out = Issue{
			ID: int64(n), Number: n, Identifier: ident, Title: in.Title, Body: in.Body,
			Status: in.Status, Priority: in.Priority, ProjectID: in.ProjectID, CycleID: in.CycleID,
			DueDate: in.DueDate, SortOrder: sort, CreatedAt: now, UpdatedAt: now,
			CompletedAt: completedAt(in.Status, now, nil), Labels: []Label{},
		}
		if in.ProjectID != nil {
			if p, ok := projectByID(m, *in.ProjectID); ok {
				slug := p.Slug
				out.ProjectSlug = &slug
			}
		}
		if in.CycleID != nil {
			if c, ok := cycleByID(m, *in.CycleID); ok {
				num := c.Number
				out.CycleNumber = &num
			}
		}
		for _, id := range in.LabelIDs {
			if l, ok := labelByID(m, id); ok {
				out.Labels = append(out.Labels, l)
			}
		}
		m.Issues = append(m.Issues, out)
		m.Comments[ident] = []Comment{}
		addActivity(m, "issue", out.ID, "created", map[string]any{"identifier": ident}, now)
		m.bump(now)
		return nil
	})
	return out, err
}

func (s *Store) UpdateIssue(identifier string, in PatchIssueInput) (Issue, error) {
	var out Issue
	err := s.mutate(func(m *mem) error {
		i := indexIssue(m, identifier)
		if i < 0 {
			return ErrNotFound
		}
		iss := m.Issues[i]
		oldStatus := iss.Status
		if in.Title != nil {
			if strings.TrimSpace(*in.Title) == "" {
				return validationf("title required")
			}
			iss.Title = strings.TrimSpace(*in.Title)
		}
		if in.Body != nil {
			iss.Body = *in.Body
		}
		if in.Status != nil {
			if !domain.ValidIssueStatus(*in.Status) {
				return validationf("invalid status")
			}
			iss.Status = *in.Status
		}
		if in.Priority != nil {
			if !domain.ValidPriority(*in.Priority) {
				return validationf("invalid priority")
			}
			iss.Priority = *in.Priority
		}
		if in.ProjectID != nil {
			iss.ProjectID = *in.ProjectID
			iss.ProjectSlug = nil
			if iss.ProjectID != nil {
				if p, ok := projectByID(m, *iss.ProjectID); ok {
					slug := p.Slug
					iss.ProjectSlug = &slug
				}
			}
		}
		if in.CycleID != nil {
			iss.CycleID = *in.CycleID
			iss.CycleNumber = nil
			if iss.CycleID != nil {
				if c, ok := cycleByID(m, *iss.CycleID); ok {
					n := c.Number
					iss.CycleNumber = &n
				}
			}
		}
		if in.DueDate != nil {
			iss.DueDate = *in.DueDate
		}
		if in.SortOrder != nil {
			iss.SortOrder = *in.SortOrder
		}
		now := domain.Now()
		iss.CompletedAt = completedAt(iss.Status, now, iss.CompletedAt)
		iss.UpdatedAt = now
		if in.LabelIDs != nil {
			iss.Labels = []Label{}
			for _, id := range *in.LabelIDs {
				if l, ok := labelByID(m, id); ok {
					iss.Labels = append(iss.Labels, l)
				}
			}
		}
		m.Issues[i] = iss
		if in.Status != nil && *in.Status != oldStatus {
			addActivity(m, "issue", iss.ID, "status_changed", map[string]any{"from": oldStatus, "to": iss.Status}, now)
		}
		m.bump(now)
		out = iss
		return nil
	})
	return out, err
}

func (s *Store) DeleteIssue(identifier string) error {
	return s.mutate(func(m *mem) error {
		i := indexIssue(m, identifier)
		if i < 0 {
			return ErrNotFound
		}
		m.Issues = append(m.Issues[:i], m.Issues[i+1:]...)
		delete(m.Comments, identifier)
		m.bump(domain.Now())
		return nil
	})
}

func (s *Store) ListComments(identifier string) ([]Comment, error) {
	var out []Comment
	err := s.snapshot(func(m *mem) error {
		if _, ok := issueByIdent(m, identifier); !ok {
			return ErrNotFound
		}
		out = append([]Comment{}, m.Comments[identifier]...)
		if out == nil {
			out = []Comment{}
		}
		return nil
	})
	return out, err
}

func (s *Store) AddComment(identifier, body string) (Comment, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Comment{}, validationf("body required")
	}
	var out Comment
	err := s.mutate(func(m *mem) error {
		iss, ok := issueByIdent(m, identifier)
		if !ok {
			return ErrNotFound
		}
		now := domain.Now()
		m.commentSeq[identifier]++
		out = Comment{ID: m.commentSeq[identifier], IssueID: iss.ID, Body: body, CreatedAt: now}
		m.Comments[identifier] = append(m.Comments[identifier], out)
		addActivity(m, "issue", iss.ID, "commented", map[string]any{"commentId": out.ID}, now)
		m.bump(now)
		return nil
	})
	return out, err
}

func (s *Store) ListActivities(identifier string) ([]Activity, error) {
	var out []Activity
	err := s.snapshot(func(m *mem) error {
		iss, ok := issueByIdent(m, identifier)
		if !ok {
			return ErrNotFound
		}
		out = []Activity{}
		for i := len(m.Activities) - 1; i >= 0; i-- {
			a := m.Activities[i]
			if a.EntityType == "issue" && a.EntityID == iss.ID {
				out = append(out, a)
			}
		}
		return nil
	})
	return out, err
}

func (s *Store) ListPages() ([]Page, error) {
	var out []Page
	err := s.snapshot(func(m *mem) error {
		out = append([]Page{}, m.Pages...)
		if out == nil {
			out = []Page{}
		}
		return nil
	})
	return out, err
}

func (s *Store) GetPage(slug string) (Page, error) {
	var p Page
	err := s.snapshot(func(m *mem) error {
		got, ok := pageBySlug(m, slug)
		if !ok {
			return ErrNotFound
		}
		p = got
		return nil
	})
	return p, err
}

func (s *Store) CreatePage(title, slug, body, status string, parentID, projectID *int64, date *string, tags []string) (Page, error) {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	if title == "" {
		return Page{}, validationf("title required")
	}
	if !domain.ValidSlug(slug) {
		return Page{}, validationf("invalid slug")
	}
	if status == "" {
		status = "proposed"
	}
	if !domain.ValidPageStatus(status) {
		return Page{}, validationf("invalid status")
	}
	if tags == nil {
		tags = []string{}
	}
	now := domain.Now()
	var out Page
	err := s.mutate(func(m *mem) error {
		if _, ok := pageBySlug(m, slug); ok {
			return errf(ErrConflict, "slug")
		}
		if err := checkPageParent(m, 0, parentID); err != nil {
			return err
		}
		out = Page{
			ID: m.nextID(), Title: title, Slug: slug, Body: body, ParentID: parentID, ProjectID: projectID,
			Status: status, Date: date, Tags: tags, CreatedAt: now, UpdatedAt: now,
		}
		if parentID != nil {
			if parent, ok := pageByID(m, *parentID); ok {
				ps := parent.Slug
				out.ParentSlug = &ps
			}
		}
		if projectID != nil {
			if proj, ok := projectByID(m, *projectID); ok {
				ps := proj.Slug
				out.ProjectSlug = &ps
			}
		}
		m.Pages = append(m.Pages, out)
		addActivity(m, "page", out.ID, "created", map[string]any{"slug": slug}, now)
		m.bump(now)
		return nil
	})
	return out, err
}

func (s *Store) UpdatePage(slug string, title, body, status *string, parentID, projectID **int64, date **string, tags *[]string) (Page, error) {
	var out Page
	err := s.mutate(func(m *mem) error {
		i := indexPage(m, slug)
		if i < 0 {
			return ErrNotFound
		}
		p := m.Pages[i]
		if title != nil {
			if strings.TrimSpace(*title) == "" {
				return validationf("title required")
			}
			p.Title = strings.TrimSpace(*title)
		}
		if body != nil {
			p.Body = *body
		}
		if status != nil {
			if !domain.ValidPageStatus(*status) {
				return validationf("invalid status")
			}
			p.Status = *status
		}
		if parentID != nil {
			if err := checkPageParent(m, p.ID, *parentID); err != nil {
				return err
			}
			p.ParentID = *parentID
			p.ParentSlug = nil
			if p.ParentID != nil {
				if parent, ok := pageByID(m, *p.ParentID); ok {
					ps := parent.Slug
					p.ParentSlug = &ps
				}
			}
		}
		if projectID != nil {
			p.ProjectID = *projectID
			p.ProjectSlug = nil
			if p.ProjectID != nil {
				if proj, ok := projectByID(m, *p.ProjectID); ok {
					ps := proj.Slug
					p.ProjectSlug = &ps
				}
			}
		}
		if date != nil {
			p.Date = *date
		}
		if tags != nil {
			p.Tags = *tags
		}
		now := domain.Now()
		p.UpdatedAt = now
		m.Pages[i] = p
		m.bump(now)
		out = p
		return nil
	})
	return out, err
}

func (s *Store) DeletePage(slug string) error {
	return s.mutate(func(m *mem) error {
		i := indexPage(m, slug)
		if i < 0 {
			return ErrNotFound
		}
		m.Pages = append(m.Pages[:i], m.Pages[i+1:]...)
		m.bump(domain.Now())
		return nil
	})
}

func checkPageParent(m *mem, pageID int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	if *parentID == pageID && pageID != 0 {
		return validationf("page cannot be its own parent")
	}
	cur := *parentID
	seen := map[int64]struct{}{pageID: {}}
	for cur != 0 {
		if _, ok := seen[cur]; ok {
			return validationf("page parent cycle")
		}
		seen[cur] = struct{}{}
		p, ok := pageByID(m, cur)
		if !ok {
			return validationf("parent not found")
		}
		if p.ParentID == nil {
			return nil
		}
		cur = *p.ParentID
	}
	return nil
}

func (s *Store) Search(q string) ([]SearchHit, error) {
	q = strings.ToLower(strings.TrimSpace(q))
	var hits []SearchHit
	err := s.snapshot(func(m *mem) error {
		if q == "" {
			hits = []SearchHit{}
			return nil
		}
		for _, iss := range m.Issues {
			if strings.Contains(strings.ToLower(iss.Title), q) || strings.Contains(strings.ToLower(iss.Identifier), q) {
				hits = append(hits, SearchHit{Kind: "issue", ID: iss.Identifier, Title: iss.Title})
			}
		}
		for _, p := range m.Projects {
			if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Slug), q) {
				hits = append(hits, SearchHit{Kind: "project", ID: p.Slug, Title: p.Name})
			}
		}
		for _, p := range m.Pages {
			if strings.Contains(strings.ToLower(p.Title), q) || strings.Contains(strings.ToLower(p.Slug), q) {
				hits = append(hits, SearchHit{Kind: "page", ID: p.Slug, Title: p.Title})
			}
		}
		if hits == nil {
			hits = []SearchHit{}
		}
		return nil
	})
	return hits, err
}

func (s *Store) Counts() (issues, pages int, err error) {
	err = s.snapshot(func(m *mem) error {
		issues = len(m.Issues)
		pages = len(m.Pages)
		return nil
	})
	return
}

func addActivity(m *mem, entityType string, entityID int64, action string, payload map[string]any, now string) {
	raw, _ := json.Marshal(payload)
	id := int64(len(m.Activities) + 1)
	m.Activities = append(m.Activities, Activity{
		ID: id, EntityType: entityType, EntityID: entityID, Action: action,
		Payload: json.RawMessage(raw), CreatedAt: now,
	})
}

func projectProgress(m *mem, id int64) float64 {
	total, done := 0, 0
	for _, iss := range m.Issues {
		if iss.ProjectID != nil && *iss.ProjectID == id {
			total++
			if iss.Status == "done" || iss.Status == "canceled" {
				done++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(done) / float64(total)
}

func issueByIdent(m *mem, ident string) (Issue, bool) {
	for _, iss := range m.Issues {
		if iss.Identifier == ident {
			return iss, true
		}
	}
	return Issue{}, false
}

func indexIssue(m *mem, ident string) int {
	for i, iss := range m.Issues {
		if iss.Identifier == ident {
			return i
		}
	}
	return -1
}

func indexProject(m *mem, slug string) int {
	for i, p := range m.Projects {
		if p.Slug == slug {
			return i
		}
	}
	return -1
}

func indexCycle(m *mem, number int) int {
	for i, c := range m.Cycles {
		if c.Number == number {
			return i
		}
	}
	return -1
}

func indexPage(m *mem, slug string) int {
	for i, p := range m.Pages {
		if p.Slug == slug {
			return i
		}
	}
	return -1
}

func completedAt(status, now string, current *string) *string {
	if status == "done" || status == "canceled" {
		if current != nil {
			return current
		}
		return &now
	}
	return nil
}

func parseTime(s string) error {
	if s == "" {
		return validationf("timestamp required")
	}
	if _, err := time.Parse(time.RFC3339, s); err != nil {
		return validationf("invalid RFC3339 time %q", s)
	}
	return nil
}

func validColor(c string) bool {
	if len(c) != 7 || c[0] != '#' {
		return false
	}
	for _, r := range c[1:] {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
