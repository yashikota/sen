package syncer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type ORAS struct {
	Token string
}

func (o *ORAS) Push(ctx context.Context, ref, tag, dir string) (string, error) {
	store, err := file.New(dir)
	if err != nil {
		return "", err
	}
	defer func() { _ = store.Close() }()

	var layers []ocispec.Descriptor
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
		rel = filepath.ToSlash(rel)
		desc, err := store.Add(ctx, rel, "application/vnd.sen.file.v1", "")
		if err != nil {
			return fmt.Errorf("add %s: %w", rel, err)
		}
		layers = append(layers, desc)
		return nil
	})
	if err != nil {
		return "", err
	}

	manifest, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, artifactType, oras.PackManifestOptions{Layers: layers})
	if err != nil {
		return "", err
	}
	if err := store.Tag(ctx, manifest, tag); err != nil {
		return "", err
	}
	repo, err := o.repository(ref)
	if err != nil {
		return "", err
	}
	desc, err := oras.Copy(ctx, store, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return "", err
	}
	return desc.Digest.String(), nil
}

func (o *ORAS) Pull(ctx context.Context, ref, tag, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	store, err := file.New(dir)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()
	repo, err := o.repository(ref)
	if err != nil {
		return err
	}
	_, err = oras.Copy(ctx, repo, tag, store, tag, oras.DefaultCopyOptions)
	return err
}

func (o *ORAS) repository(ref string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, err
	}
	host, _, _ := strings.Cut(ref, "/")
	if host == "" {
		host = "ghcr.io"
	}
	user := "token"
	if _, rest, ok := strings.Cut(ref, "/"); ok {
		user, _, _ = strings.Cut(rest, "/")
	}
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(host, auth.Credential{
			Username: user,
			Password: o.Token,
		}),
	}
	return repo, nil
}
