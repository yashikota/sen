package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yashikota/sen/internal/domain"
	"github.com/yashikota/sen/internal/store"
)

var ErrDirty = errors.New("local workspace has unpushed changes")
var ErrNoRef = errors.New("workspace ghcrRef is empty")
var ErrNoWorkspace = errors.New("artifact is missing workspace.toml")

const artifactType = "application/vnd.sen.workspace.v1"

type Registry interface {
	Push(ctx context.Context, ref, tag, dir string) (digest string, err error)
	Pull(ctx context.Context, ref, tag, dir string) error
}

type Manifest struct {
	CreatedAt string `json:"createdAt"`
	Version   string `json:"version"`
	Issues    int    `json:"issues"`
	Pages     int    `json:"pages"`
}

type Service struct {
	Store    *store.Store
	Registry Registry
	Version  string
	Warn     func(string)
}

func (s *Service) Export(dir string) (*Manifest, error) {
	if err := s.Store.Snapshot(dir); err != nil {
		return nil, err
	}
	issues, pageCount, err := s.Store.Counts()
	if err != nil {
		return nil, err
	}
	man := Manifest{
		CreatedAt: domain.Now(),
		Version:   s.Version,
		Issues:    issues,
		Pages:     pageCount,
	}
	raw, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), append(raw, '\n'), 0o644); err != nil {
		return nil, err
	}
	return &man, nil
}

func (s *Service) Push(ctx context.Context) (string, string, error) {
	ws, err := s.Store.Workspace()
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(ws.GHCRRef) == "" {
		return "", "", ErrNoRef
	}
	dir, err := os.MkdirTemp("", "sen-push-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(dir)
	if _, err := s.Export(dir); err != nil {
		return "", "", err
	}
	tag := time.Now().UTC().Format("20060102T150405Z")
	digest, err := s.Registry.Push(ctx, ws.GHCRRef, tag, dir)
	if err != nil {
		return "", "", err
	}
	if _, err := s.Registry.Push(ctx, ws.GHCRRef, "latest", dir); err != nil {
		return "", "", fmt.Errorf("push latest after %s: %w", tag, err)
	}
	if err := s.Store.MarkPushed(digest); err != nil {
		return "", "", err
	}
	return tag, digest, nil
}

func (s *Service) Pull(ctx context.Context, tag string) error {
	if tag == "" {
		tag = "latest"
	}
	dirty, err := s.Store.Dirty()
	if err != nil {
		return err
	}
	if dirty {
		return ErrDirty
	}
	ws, err := s.Store.Workspace()
	if err != nil {
		return err
	}
	if strings.TrimSpace(ws.GHCRRef) == "" {
		return ErrNoRef
	}
	dir, err := os.MkdirTemp("", "sen-pull-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := s.Registry.Pull(ctx, ws.GHCRRef, tag, dir); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, "workspace.toml")); err != nil {
		return ErrNoWorkspace
	}
	if err := s.Store.ReplaceFrom(dir); err != nil {
		return err
	}
	return nil
}
