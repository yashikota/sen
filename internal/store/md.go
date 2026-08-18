package store

import (
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/yashikota/sen/internal/domain"
)

type issueFM struct {
	Title     string      `toml:"title"`
	Status    string      `toml:"status"`
	Priority  int         `toml:"priority"`
	Project   *string     `toml:"project,omitempty"`
	Cycle     *int        `toml:"cycle,omitempty"`
	Labels    []string    `toml:"labels"`
	Due       *string     `toml:"due,omitempty"`
	Sort      float64     `toml:"sort"`
	Created   string      `toml:"created"`
	Updated   string      `toml:"updated"`
	Completed *string     `toml:"completed,omitempty"`
	Comments  []commentFM `toml:"comments,omitempty"`
}

type commentFM struct {
	ID      int64  `toml:"id"`
	Created string `toml:"created"`
	Body    string `toml:"body"`
}

type pageFM struct {
	ID      int64    `toml:"id"`
	Title   string   `toml:"title"`
	Slug    string   `toml:"slug"`
	Status  string   `toml:"status"`
	Date    *string  `toml:"date,omitempty"`
	Tags    []string `toml:"tags"`
	Project *string  `toml:"project,omitempty"`
	Parent  *string  `toml:"parent,omitempty"`
	Created string   `toml:"created"`
	Updated string   `toml:"updated"`
}

func splitFrontmatter(raw string) (block, body string, err error) {
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	if strings.HasPrefix(s, "---\n") {
		return "", "", fmt.Errorf("yaml frontmatter is not supported; use +++ TOML")
	}
	if !strings.HasPrefix(s, "+++\n") {
		return "", strings.TrimPrefix(s, "+++\n"), nil
	}
	rest := s[4:]
	idx := strings.Index(rest, "\n+++")
	if idx < 0 {
		return "", "", fmt.Errorf("unterminated frontmatter")
	}
	block = rest[:idx]
	body = strings.TrimPrefix(rest[idx+4:], "\n")
	return block, body, nil
}

func parseIssueMarkdown(ident string, raw string, m *mem) (Issue, []Comment, error) {
	n, ok := domain.ParseIdentifier(ident)
	if !ok {
		return Issue{}, nil, fmt.Errorf("invalid identifier %q", ident)
	}
	block, body, err := splitFrontmatter(raw)
	if err != nil {
		return Issue{}, nil, err
	}
	var fm issueFM
	if strings.TrimSpace(block) != "" {
		if err := toml.Unmarshal([]byte(block), &fm); err != nil {
			return Issue{}, nil, err
		}
	}
	if fm.Title == "" {
		fm.Title = ident
	}
	if fm.Status == "" {
		fm.Status = "backlog"
	}
	if fm.Created == "" {
		fm.Created = domain.Now()
	}
	if fm.Updated == "" {
		fm.Updated = fm.Created
	}
	iss := Issue{
		ID:          int64(n),
		Number:      n,
		Identifier:  ident,
		Title:       fm.Title,
		Body:        body,
		Status:      fm.Status,
		Priority:    fm.Priority,
		ProjectSlug: fm.Project,
		CycleNumber: fm.Cycle,
		DueDate:     fm.Due,
		SortOrder:   fm.Sort,
		CreatedAt:   fm.Created,
		UpdatedAt:   fm.Updated,
		CompletedAt: fm.Completed,
		Labels:      []Label{},
	}
	for _, name := range fm.Labels {
		if l, ok := labelByName(m, name); ok {
			iss.Labels = append(iss.Labels, l)
			continue
		}
		iss.Labels = append(iss.Labels, Label{Name: name})
	}
	comments := make([]Comment, 0, len(fm.Comments))
	for i, c := range fm.Comments {
		id := c.ID
		if id == 0 {
			id = int64(i + 1)
			m.dirtyMeta = true
		}
		comments = append(comments, Comment{ID: id, IssueID: int64(n), Body: c.Body, CreatedAt: c.Created})
	}
	return iss, comments, nil
}

func parsePageMarkdown(raw string, m *mem) (Page, error) {
	block, body, err := splitFrontmatter(raw)
	if err != nil {
		return Page{}, err
	}
	var fm pageFM
	if strings.TrimSpace(block) != "" {
		if err := toml.Unmarshal([]byte(block), &fm); err != nil {
			return Page{}, err
		}
	}
	if fm.Status == "" {
		fm.Status = "proposed"
	}
	if fm.Tags == nil {
		fm.Tags = []string{}
	}
	if fm.Created == "" {
		fm.Created = domain.Now()
	}
	if fm.Updated == "" {
		fm.Updated = fm.Created
	}
	p := Page{
		ID:          fm.ID,
		Title:       fm.Title,
		Slug:        fm.Slug,
		Body:        body,
		Status:      fm.Status,
		Date:        fm.Date,
		Tags:        fm.Tags,
		ProjectSlug: fm.Project,
		ParentSlug:  fm.Parent,
		CreatedAt:   fm.Created,
		UpdatedAt:   fm.Updated,
	}
	if p.ID == 0 {
		p.ID = m.nextID()
	} else {
		m.observeID(p.ID)
	}
	return p, nil
}

func renderIssueMarkdown(iss Issue, comments []Comment, m *mem) string {
	fm := issueFM{
		Title:     iss.Title,
		Status:    iss.Status,
		Priority:  iss.Priority,
		Due:       iss.DueDate,
		Sort:      iss.SortOrder,
		Created:   iss.CreatedAt,
		Updated:   iss.UpdatedAt,
		Completed: iss.CompletedAt,
		Labels:    make([]string, 0, len(iss.Labels)),
	}
	if iss.ProjectID != nil {
		if p, ok := projectByID(m, *iss.ProjectID); ok {
			slug := p.Slug
			fm.Project = &slug
		}
	} else if iss.ProjectSlug != nil {
		fm.Project = iss.ProjectSlug
	}
	if iss.CycleID != nil {
		if c, ok := cycleByID(m, *iss.CycleID); ok {
			n := c.Number
			fm.Cycle = &n
		}
	} else if iss.CycleNumber != nil {
		fm.Cycle = iss.CycleNumber
	}
	for _, l := range iss.Labels {
		fm.Labels = append(fm.Labels, l.Name)
	}
	for _, c := range comments {
		fm.Comments = append(fm.Comments, commentFM{ID: c.ID, Created: c.CreatedAt, Body: c.Body})
	}
	return marshalDoc(fm, iss.Body)
}

func renderPageMarkdown(p Page, m *mem) string {
	fm := pageFM{
		ID:      p.ID,
		Title:   p.Title,
		Slug:    p.Slug,
		Status:  p.Status,
		Date:    p.Date,
		Tags:    p.Tags,
		Created: p.CreatedAt,
		Updated: p.UpdatedAt,
	}
	if fm.Tags == nil {
		fm.Tags = []string{}
	}
	if p.ProjectID != nil {
		if proj, ok := projectByID(m, *p.ProjectID); ok {
			slug := proj.Slug
			fm.Project = &slug
		}
	} else if p.ProjectSlug != nil {
		fm.Project = p.ProjectSlug
	}
	if p.ParentID != nil {
		if parent, ok := pageByID(m, *p.ParentID); ok {
			slug := parent.Slug
			fm.Parent = &slug
		}
	} else if p.ParentSlug != nil {
		fm.Parent = p.ParentSlug
	}
	return marshalDoc(fm, p.Body)
}

func marshalDoc(fm any, body string) string {
	b, err := toml.Marshal(fm)
	if err != nil {
		b = []byte{}
	}
	var out strings.Builder
	out.WriteString("+++\n")
	out.Write(b)
	if len(b) > 0 && !strings.HasSuffix(string(b), "\n") {
		out.WriteByte('\n')
	}
	out.WriteString("+++\n\n")
	out.WriteString(body)
	if body != "" && !strings.HasSuffix(body, "\n") {
		out.WriteByte('\n')
	}
	return out.String()
}

func labelByName(m *mem, name string) (Label, bool) {
	for _, l := range m.Labels {
		if l.Name == name {
			return l, true
		}
	}
	return Label{}, false
}
