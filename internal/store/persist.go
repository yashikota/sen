package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yashikota/sen/internal/domain"
	"github.com/pelletier/go-toml/v2"
)

func errf(kind error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", kind, fmt.Sprintf(format, args...))
}

func (s *Store) snapshot(fn func(*mem) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := load(s.root)
	if err != nil {
		return err
	}
	if m.dirtyMeta {
		if err := save(s.root, m); err != nil {
			return err
		}
	}
	return fn(m)
}

func (s *Store) mutate(fn func(*mem) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := load(s.root)
	if err != nil {
		return err
	}
	if err := fn(m); err != nil {
		return err
	}
	return save(s.root, m)
}

func seed(root string) error {
	now := domain.Now()
	m := &mem{
		Workspace: workspaceFile{
			Name:         "sen",
			Timezone:     "Asia/Tokyo",
			IssueCounter: 0,
			NextID:       4,
			UpdatedAt:    now,
		},
		Labels: []Label{
			{ID: 1, Name: "Bug", Color: "#d4725a"},
			{ID: 2, Name: "Feature", Color: "#6b9bd1"},
			{ID: 3, Name: "Improvement", Color: "#c4a574"},
		},
		Comments:   map[string][]Comment{},
		commentSeq: map[string]int64{},
	}
	return save(root, m)
}

func load(root string) (*mem, error) {
	raw, err := os.ReadFile(filepath.Join(root, "workspace.toml"))
	if err != nil {
		return nil, err
	}
	m := &mem{Comments: map[string][]Comment{}, commentSeq: map[string]int64{}}
	if err := toml.Unmarshal(raw, &m.Workspace); err != nil {
		return nil, fmt.Errorf("workspace.toml: %w", err)
	}
	if m.Workspace.NextID < 1 {
		m.Workspace.NextID = 1
	}

	if b, err := os.ReadFile(filepath.Join(root, "labels.toml")); err == nil {
		var lf labelsFile
		if err := toml.Unmarshal(b, &lf); err != nil {
			return nil, fmt.Errorf("labels.toml: %w", err)
		}
		m.Labels = lf.Labels
	}
	if m.Labels == nil {
		m.Labels = []Label{}
	}
	for i := range m.Labels {
		if m.Labels[i].ID == 0 {
			m.Labels[i].ID = m.nextID()
		} else {
			m.observeID(m.Labels[i].ID)
		}
	}

	if err := readTOMLDir(filepath.Join(root, "projects"), func(name string, b []byte) error {
		var p Project
		if err := toml.Unmarshal(b, &p); err != nil {
			return err
		}
		stem := strings.TrimSuffix(name, ".toml")
		if p.Slug != "" && p.Slug != stem {
			m.diag("projects/"+name, "slug_mismatch", fmt.Sprintf("slug %q does not match filename", p.Slug))
		}
		p.Slug = stem
		if p.ID == 0 {
			p.ID = m.nextID()
		} else {
			m.observeID(p.ID)
		}
		m.Projects = append(m.Projects, p)
		return nil
	}); err != nil {
		return nil, err
	}

	if err := readTOMLDir(filepath.Join(root, "cycles"), func(name string, b []byte) error {
		var c Cycle
		if err := toml.Unmarshal(b, &c); err != nil {
			return err
		}
		stem := strings.TrimSuffix(name, ".toml")
		n, err := strconv.Atoi(stem)
		if err != nil || n < 1 {
			return fmt.Errorf("filename must be <n>.toml")
		}
		if c.Number != 0 && c.Number != n {
			m.diag("cycles/"+name, "number_mismatch", fmt.Sprintf("number %d does not match filename", c.Number))
		}
		c.Number = n
		if c.ID == 0 {
			c.ID = m.nextID()
		} else {
			m.observeID(c.ID)
		}
		m.Cycles = append(m.Cycles, c)
		return nil
	}); err != nil {
		return nil, err
	}

	issueDir := filepath.Join(root, "issues")
	ents, err := os.ReadDir(issueDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(issueDir, ent.Name()))
		if err != nil {
			return nil, err
		}
		iss, comments, err := parseIssueMarkdown(strings.TrimSuffix(ent.Name(), ".md"), string(b), m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ent.Name(), err)
		}
		m.Issues = append(m.Issues, iss)
		m.Comments[iss.Identifier] = comments
		var maxID int64
		for _, c := range comments {
			if c.ID > maxID {
				maxID = c.ID
			}
		}
		m.commentSeq[iss.Identifier] = maxID
		if iss.Number > m.Workspace.IssueCounter {
			m.Workspace.IssueCounter = iss.Number
			m.dirtyMeta = true
		}
	}

	pageDir := filepath.Join(root, "pages")
	pents, err := os.ReadDir(pageDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, ent := range pents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(pageDir, ent.Name()))
		if err != nil {
			return nil, err
		}
		stem := strings.TrimSuffix(ent.Name(), ".md")
		p, err := parsePageMarkdown(string(b), m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ent.Name(), err)
		}
		if p.Slug != "" && p.Slug != stem {
			m.diag("pages/"+ent.Name(), "slug_mismatch", fmt.Sprintf("slug %q does not match filename", p.Slug))
		}
		p.Slug = stem
		m.Pages = append(m.Pages, p)
	}

	if b, err := os.ReadFile(filepath.Join(root, "activities.jsonl")); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var a Activity
			if err := json.Unmarshal([]byte(line), &a); err != nil {
				return nil, fmt.Errorf("activities.jsonl: %w", err)
			}
			m.Activities = append(m.Activities, a)
		}
	}

	fillIssueRefs(m)
	fillPageRefs(m)
	diagnose(m)
	return m, nil
}

func save(root string, m *mem) error {
	if err := os.MkdirAll(filepath.Join(root, "issues"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "pages"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "projects"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "cycles"), 0o755); err != nil {
		return err
	}

	if err := writeTOML(filepath.Join(root, "workspace.toml"), m.Workspace); err != nil {
		return err
	}
	if err := writeTOML(filepath.Join(root, "labels.toml"), labelsFile{Labels: m.Labels}); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(root, "workspace.yaml"))
	_ = os.Remove(filepath.Join(root, "labels.yaml"))

	wantProjects := map[string]struct{}{}
	for _, p := range m.Projects {
		wantProjects[p.Slug+".toml"] = struct{}{}
		if err := writeTOML(filepath.Join(root, "projects", p.Slug+".toml"), p); err != nil {
			return err
		}
	}
	if err := pruneDir(filepath.Join(root, "projects"), ".toml", wantProjects); err != nil {
		return err
	}
	_ = pruneDir(filepath.Join(root, "projects"), ".yaml", map[string]struct{}{})

	wantCycles := map[string]struct{}{}
	for _, c := range m.Cycles {
		name := fmt.Sprintf("%d.toml", c.Number)
		wantCycles[name] = struct{}{}
		if err := writeTOML(filepath.Join(root, "cycles", name), c); err != nil {
			return err
		}
	}
	if err := pruneDir(filepath.Join(root, "cycles"), ".toml", wantCycles); err != nil {
		return err
	}
	_ = pruneDir(filepath.Join(root, "cycles"), ".yaml", map[string]struct{}{})

	wantIssues := map[string]struct{}{}
	for _, iss := range m.Issues {
		name := iss.Identifier + ".md"
		wantIssues[name] = struct{}{}
		md := renderIssueMarkdown(iss, m.Comments[iss.Identifier], m)
		if err := atomicWrite(filepath.Join(root, "issues", name), []byte(md)); err != nil {
			return err
		}
	}
	if err := pruneDir(filepath.Join(root, "issues"), ".md", wantIssues); err != nil {
		return err
	}

	wantPages := map[string]struct{}{}
	for _, p := range m.Pages {
		name := p.Slug + ".md"
		wantPages[name] = struct{}{}
		md := renderPageMarkdown(p, m)
		if err := atomicWrite(filepath.Join(root, "pages", name), []byte(md)); err != nil {
			return err
		}
	}
	if err := pruneDir(filepath.Join(root, "pages"), ".md", wantPages); err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, a := range m.Activities {
		b, err := json.Marshal(a)
		if err != nil {
			return err
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	if err := atomicWrite(filepath.Join(root, "activities.jsonl"), buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (s *Store) Snapshot(dest string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := load(s.root)
	if err != nil {
		return err
	}
	return save(dest, m)
}

func (s *Store) ReplaceFrom(src string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := load(src)
	if err != nil {
		return err
	}
	return save(s.root, m)
}

func readTOMLDir(dir string, fn func(name string, b []byte) error) error {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".toml") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, ent.Name()))
		if err != nil {
			return err
		}
		if err := fn(ent.Name(), b); err != nil {
			return fmt.Errorf("%s: %w", ent.Name(), err)
		}
	}
	return nil
}

func writeTOML(path string, v any) error {
	b, err := toml.Marshal(v)
	if err != nil {
		return err
	}
	return atomicWrite(path, b)
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func pruneDir(dir, ext string, keep map[string]struct{}) error {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ext) {
			continue
		}
		if _, ok := keep[ent.Name()]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, ent.Name())); err != nil {
			return err
		}
	}
	return nil
}

func fillIssueRefs(m *mem) {
	for i := range m.Issues {
		iss := &m.Issues[i]
		iss.ID = int64(iss.Number)
		iss.Identifier = domain.Identifier(iss.Number)
		if iss.Labels == nil {
			iss.Labels = []Label{}
		}
		if iss.ProjectSlug != nil {
			if p, ok := projectBySlug(m, *iss.ProjectSlug); ok {
				id := p.ID
				iss.ProjectID = &id
			}
		}
		if iss.CycleNumber != nil {
			if c, ok := cycleByNumber(m, *iss.CycleNumber); ok {
				id := c.ID
				iss.CycleID = &id
			}
		}
	}
}

func fillPageRefs(m *mem) {
	for i := range m.Pages {
		p := &m.Pages[i]
		if p.ProjectSlug != nil {
			if proj, ok := projectBySlug(m, *p.ProjectSlug); ok {
				id := proj.ID
				p.ProjectID = &id
			}
		}
		if p.ParentSlug != nil && *p.ParentSlug != "" {
			if parent, ok := pageBySlug(m, *p.ParentSlug); ok {
				id := parent.ID
				p.ParentID = &id
			}
		}
	}
}

func projectBySlug(m *mem, slug string) (Project, bool) {
	for _, p := range m.Projects {
		if p.Slug == slug {
			return p, true
		}
	}
	return Project{}, false
}

func projectByID(m *mem, id int64) (Project, bool) {
	for _, p := range m.Projects {
		if p.ID == id {
			return p, true
		}
	}
	return Project{}, false
}

func cycleByNumber(m *mem, n int) (Cycle, bool) {
	for _, c := range m.Cycles {
		if c.Number == n {
			return c, true
		}
	}
	return Cycle{}, false
}

func cycleByID(m *mem, id int64) (Cycle, bool) {
	for _, c := range m.Cycles {
		if c.ID == id {
			return c, true
		}
	}
	return Cycle{}, false
}

func pageBySlug(m *mem, slug string) (Page, bool) {
	for _, p := range m.Pages {
		if p.Slug == slug {
			return p, true
		}
	}
	return Page{}, false
}

func pageByID(m *mem, id int64) (Page, bool) {
	for _, p := range m.Pages {
		if p.ID == id {
			return p, true
		}
	}
	return Page{}, false
}

func labelByID(m *mem, id int64) (Label, bool) {
	for _, l := range m.Labels {
		if l.ID == id {
			return l, true
		}
	}
	return Label{}, false
}

func (m *mem) nextID() int64 {
	id := m.Workspace.NextID
	if id < 1 {
		id = 1
	}
	m.Workspace.NextID = id + 1
	m.dirtyMeta = true
	return id
}

func (m *mem) observeID(id int64) {
	if id >= m.Workspace.NextID {
		m.Workspace.NextID = id + 1
		m.dirtyMeta = true
	}
}

func (m *mem) bump(now string) {
	m.Workspace.UpdatedAt = now
}
