package syncer

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type Memory struct {
	mu       sync.Mutex
	blobs    map[string]map[string][]byte
	FailPush error
}

func NewMemory() *Memory {
	return &Memory{blobs: map[string]map[string][]byte{}}
}

func key(ref, tag string) string {
	return ref + ":" + tag
}

func (m *Memory) Push(_ context.Context, ref, tag, dir string) (string, error) {
	if m.FailPush != nil {
		return "", m.FailPush
	}
	files := map[string][]byte{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = b
		return nil
	})
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobs[key(ref, tag)] = files
	return "sha256:testdigest", nil
}

func (m *Memory) Pull(_ context.Context, ref, tag, dir string) error {
	m.mu.Lock()
	files, ok := m.blobs[key(ref, tag)]
	m.mu.Unlock()
	if !ok {
		return errors.New("artifact not found")
	}
	for name, b := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
