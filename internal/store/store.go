package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation")
	ErrConflict   = errors.New("conflict")
)

type Store struct {
	mu   sync.Mutex
	root string
}

type Workspace struct {
	Name             string  `json:"name"`
	GHCRRef          string  `json:"ghcrRef"`
	Timezone         string  `json:"timezone"`
	LastPushedAt     *string `json:"lastPushedAt"`
	LastPushedDigest *string `json:"lastPushedDigest"`
	UpdatedAt        string  `json:"updatedAt"`
}

type Label struct {
	ID    int64  `json:"id" toml:"id"`
	Name  string `json:"name" toml:"name"`
	Color string `json:"color" toml:"color"`
}

type Project struct {
	ID          int64   `json:"id" toml:"id"`
	Name        string  `json:"name" toml:"name"`
	Slug        string  `json:"slug" toml:"slug"`
	Description string  `json:"description" toml:"description"`
	Status      string  `json:"status" toml:"status"`
	StartDate   *string `json:"startDate" toml:"startDate,omitempty"`
	TargetDate  *string `json:"targetDate" toml:"targetDate,omitempty"`
	Progress    float64 `json:"progress" toml:"-"`
	CreatedAt   string  `json:"createdAt" toml:"createdAt"`
	UpdatedAt   string  `json:"updatedAt" toml:"updatedAt"`
}

type Cycle struct {
	ID        int64  `json:"id" toml:"id"`
	Number    int    `json:"number" toml:"number"`
	StartsAt  string `json:"startsAt" toml:"startsAt"`
	EndsAt    string `json:"endsAt" toml:"endsAt"`
	Status    string `json:"status" toml:"status"`
	CreatedAt string `json:"createdAt" toml:"createdAt"`
	UpdatedAt string `json:"updatedAt" toml:"updatedAt"`
}

type Issue struct {
	ID               int64   `json:"id"`
	Number           int     `json:"number"`
	Identifier       string  `json:"identifier"`
	Title            string  `json:"title"`
	Body             string  `json:"body"`
	Status           string  `json:"status"`
	Priority         int     `json:"priority"`
	ProjectID        *int64  `json:"projectId"`
	ProjectSlug      *string `json:"projectSlug,omitempty"`
	CycleID          *int64  `json:"cycleId"`
	CycleNumber      *int    `json:"cycleNumber,omitempty"`
	ParentID         *int64  `json:"parentId"`
	ParentIdentifier *string `json:"parentIdentifier,omitempty"`
	Depth            int     `json:"depth"`
	DueDate          *string `json:"dueDate"`
	SortOrder        float64 `json:"sortOrder"`
	Labels           []Label `json:"labels"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	CompletedAt      *string `json:"completedAt"`
}

type Comment struct {
	ID        int64  `json:"id"`
	IssueID   int64  `json:"issueId"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

type Activity struct {
	ID         int64           `json:"id"`
	EntityType string          `json:"entityType"`
	EntityID   int64           `json:"entityId"`
	Action     string          `json:"action"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  string          `json:"createdAt"`
}

type Page struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Body        string   `json:"body"`
	ParentID    *int64   `json:"parentId"`
	ParentSlug  *string  `json:"parentSlug,omitempty"`
	ProjectID   *int64   `json:"projectId"`
	ProjectSlug *string  `json:"projectSlug,omitempty"`
	Status      string   `json:"status"`
	Date        *string  `json:"date"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type View struct {
	ID        int64    `json:"id" toml:"id"`
	Name      string   `json:"name" toml:"name"`
	Slug      string   `json:"slug" toml:"slug"`
	Display   string   `json:"display" toml:"display"`
	Status    *string  `json:"status" toml:"status,omitempty"`
	Project   *string  `json:"project" toml:"project,omitempty"`
	Cycle     *int     `json:"cycle" toml:"cycle,omitempty"`
	Labels    []string `json:"labels" toml:"labels,omitempty"`
	Priority  *int     `json:"priority" toml:"priority,omitempty"`
	CreatedAt string   `json:"createdAt" toml:"createdAt"`
	UpdatedAt string   `json:"updatedAt" toml:"updatedAt"`
}

func (v View) Filter() IssueFilter {
	f := IssueFilter{Labels: v.Labels, Priority: v.Priority}
	if v.Status != nil {
		f.Status = *v.Status
	}
	if v.Project != nil {
		f.ProjectSlug = *v.Project
	}
	if v.Cycle != nil {
		f.CycleNumber = *v.Cycle
	}
	return f
}

type Diagnostic struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SearchHit struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Title string `json:"title"`
}

type IssueFilter struct {
	Status      string
	ProjectSlug string
	CycleNumber int
	Labels      []string
	Priority    *int
}

type CreateIssueInput struct {
	Title     string
	Body      string
	Status    string
	Priority  int
	ProjectID *int64
	CycleID   *int64
	ParentID  *int64
	DueDate   *string
	LabelIDs  []int64
}

type PatchIssueInput struct {
	Title     *string
	Body      *string
	Status    *string
	Priority  *int
	ProjectID **int64
	CycleID   **int64
	ParentID  **int64
	DueDate   **string
	LabelIDs  *[]int64
	SortOrder *float64
}

type CreateViewInput struct {
	Name     string
	Slug     string
	Display  string
	Status   *string
	Project  *string
	Cycle    *int
	Labels   []string
	Priority *int
}

type mem struct {
	Workspace   workspaceFile
	Labels      []Label
	Projects    []Project
	Cycles      []Cycle
	Views       []View
	Issues      []Issue
	Comments    map[string][]Comment
	Pages       []Page
	Activities  []Activity
	commentSeq  map[string]int64
	dirtyMeta   bool
	Diagnostics []Diagnostic
}

type workspaceFile struct {
	Name             string  `toml:"name"`
	GHCRRef          string  `toml:"ghcrRef"`
	Timezone         string  `toml:"timezone"`
	IssueCounter     int     `toml:"issueCounter"`
	NextID           int64   `toml:"nextID"`
	LastPushedAt     *string `toml:"lastPushedAt,omitempty"`
	LastPushedDigest *string `toml:"lastPushedDigest,omitempty"`
	UpdatedAt        string  `toml:"updatedAt"`
}

type labelsFile struct {
	Labels []Label `toml:"labels"`
}

func Open(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	s := &Store{root: root}
	marker := filepath.Join(root, "workspace.toml")
	if _, err := os.Stat(marker); err != nil {
		if _, yamlErr := os.Stat(filepath.Join(root, "workspace.yaml")); yamlErr == nil {
			return nil, fmt.Errorf("found workspace.yaml in %s; this version uses workspace.toml (move the directory aside and run sen init)", root)
		}
		if err := seed(root); err != nil {
			return nil, err
		}
	}
	m, err := load(root)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", root, err)
	}
	if m.dirtyMeta {
		if err := save(root, m); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Close() error { return nil }

func (s *Store) Path() string { return s.root }

func validationf(format string, args ...any) error {
	return errf(ErrValidation, format, args...)
}
