package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/kota/sen/internal/appdir"
	"github.com/kota/sen/internal/httpapi"
	"github.com/kota/sen/internal/store"
	"github.com/kota/sen/internal/syncer"
	"github.com/kota/sen/internal/webembed"
)

var ErrUsage = errors.New("usage")

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, version string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "init":
		return cmdInit(stdout)
	case "serve":
		return cmdServe(ctx, args[1:], stderr)
	case "push":
		return cmdPush(ctx, stdout, version)
	case "pull":
		return cmdPull(ctx, args[1:], stdout, stderr)
	case "status":
		return cmdStatus(stdout)
	case "check":
		return cmdCheck(stdout)
	case "version", "-v", "--version":
		fmt.Fprintln(stdout, version)
		return nil
	default:
		return usage()
	}
}

func usage() error {
	return fmt.Errorf("%w: sen <init|serve|push|pull|status|check|version>", ErrUsage)
}

func openStore() (*store.Store, error) {
	root, err := appdir.Home()
	if err != nil {
		return nil, err
	}
	if !appdir.Initialized(root) {
		return nil, fmt.Errorf("workspace not found at %s (run sen init)", root)
	}
	st, err := store.Open(root)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", root, err)
	}
	return st, nil
}

func cmdInit(stdout io.Writer) error {
	root, err := appdir.Home()
	if err != nil {
		return err
	}
	if appdir.Initialized(root) {
		return fmt.Errorf("already initialized: %s", root)
	}
	st, err := store.Open(root)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	fmt.Fprintf(stdout, "initialized %s\n", root)
	return nil
}

func cmdServe(ctx context.Context, args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", "127.0.0.1:7730", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	st, err := openStore()
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	h := httpapi.New(st, webembed.Dist())
	srv := &http.Server{
		Addr:              *addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	fmt.Fprintf(stderr, "sen listening on http://%s\n", ln.Addr().String())
	err = srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func newService(st *store.Store, version string, warn func(string)) (*syncer.Service, error) {
	token, err := syncer.GitHubToken()
	if err != nil {
		return nil, err
	}
	return &syncer.Service{
		Store:    st,
		Registry: &syncer.ORAS{Token: token},
		Version:  version,
		Warn:     warn,
	}, nil
}

func cmdPush(ctx context.Context, stdout io.Writer, version string) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	svc, err := newService(st, version, nil)
	if err != nil {
		return err
	}
	tag, digest, err := svc.Push(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "pushed %s (%s)\n", tag, digest)
	return nil
}

func cmdPull(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tag := fs.String("tag", "latest", "artifact tag")
	if err := fs.Parse(args); err != nil {
		return err
	}
	st, err := openStore()
	if err != nil {
		return err
	}
	svc, err := newService(st, "", func(msg string) {
		fmt.Fprintln(stderr, msg)
	})
	if err != nil {
		_ = st.Close()
		return err
	}
	if err := svc.Pull(ctx, *tag); err != nil {
		_ = st.Close()
		return err
	}
	fmt.Fprintf(stdout, "pulled %s into %s\n", *tag, st.Path())
	return nil
}

func cmdStatus(stdout io.Writer) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	ws, err := st.Workspace()
	if err != nil {
		return err
	}
	dirty, err := st.Dirty()
	if err != nil {
		return err
	}
	digest := ""
	if ws.LastPushedDigest != nil {
		digest = *ws.LastPushedDigest
	}
	fmt.Fprintf(stdout, "path\t%s\n", st.Path())
	fmt.Fprintf(stdout, "name\t%s\n", ws.Name)
	fmt.Fprintf(stdout, "ghcr\t%s\n", ws.GHCRRef)
	fmt.Fprintf(stdout, "dirty\t%t\n", dirty)
	fmt.Fprintf(stdout, "digest\t%s\n", digest)
	return nil
}

func cmdCheck(stdout io.Writer) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	diags, err := st.Diagnostics()
	if err != nil {
		return err
	}
	if len(diags) == 0 {
		fmt.Fprintln(stdout, "ok")
		return nil
	}
	for _, d := range diags {
		fmt.Fprintf(stdout, "%s: %s\n", d.Path, d.Message)
	}
	return fmt.Errorf("check failed: %d issue(s)", len(diags))
}
