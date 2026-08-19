package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	urfavecli "github.com/urfave/cli/v3"

	"github.com/yashikota/sen/internal/appdir"
	"github.com/yashikota/sen/internal/httpapi"
	"github.com/yashikota/sen/internal/store"
	"github.com/yashikota/sen/internal/syncer"
	"github.com/yashikota/sen/internal/webembed"
)

func Run(ctx context.Context, osArgs []string, stdout, stderr io.Writer, version string) error {
	if len(osArgs) == 0 {
		osArgs = []string{"sen"}
	} else if osArgs[0] != "sen" {
		osArgs = append([]string{"sen"}, osArgs...)
	}

	prevPrinter := urfavecli.VersionPrinter
	urfavecli.VersionPrinter = func(c *urfavecli.Command) {
		fmt.Fprintln(c.Root().Writer, c.Version)
	}
	defer func() { urfavecli.VersionPrinter = prevPrinter }()

	root := newRootCommand(stdout, stderr, version)
	root.ExitErrHandler = func(_ context.Context, _ *urfavecli.Command, _ error) {}
	return root.Run(ctx, osArgs)
}

func newRootCommand(stdout, stderr io.Writer, version string) *urfavecli.Command {
	return &urfavecli.Command{
		Name:           "sen",
		Usage:          "local-first issue tracker",
		Description:    "Workspace defaults to ./.sen (override with SEN_HOME).",
		Writer:         stdout,
		ErrWriter:      stderr,
		Version:        version,
		DefaultCommand: "help",
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Development (Task)": "task dev   API + Vite dev server",
			}
		},
		Commands: []*urfavecli.Command{
			{
				Name:  "init",
				Usage: "create .sen/ in the current directory",
				Action: func(_ context.Context, _ *urfavecli.Command) error {
					return cmdInit(stdout)
				},
			},
			{
				Name:  "serve",
				Usage: "JSON API and UI on 127.0.0.1:7730",
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:  "addr",
						Value: "127.0.0.1:7730",
						Usage: "listen address",
					},
				},
				Action: func(ctx context.Context, c *urfavecli.Command) error {
					return cmdServe(ctx, c.String("addr"), c.ErrWriter)
				},
			},
			{
				Name:  "push",
				Usage: "snapshot workspace to GHCR",
				Action: func(ctx context.Context, _ *urfavecli.Command) error {
					return cmdPush(ctx, stdout, version)
				},
			},
			{
				Name:  "pull",
				Usage: "restore workspace from GHCR",
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:  "tag",
						Value: "latest",
						Usage: "artifact tag",
					},
				},
				Action: func(ctx context.Context, c *urfavecli.Command) error {
					return cmdPull(ctx, c.String("tag"), stdout, c.ErrWriter)
				},
			},
			{
				Name:  "status",
				Usage: "show path, dirty state, last digest",
				Action: func(_ context.Context, _ *urfavecli.Command) error {
					return cmdStatus(stdout)
				},
			},
			{
				Name:  "check",
				Usage: "list semantic file issues",
				Action: func(_ context.Context, _ *urfavecli.Command) error {
					return cmdCheck(stdout)
				},
			},
			{
				Name:  "version",
				Usage: "print build version",
				Action: func(_ context.Context, c *urfavecli.Command) error {
					fmt.Fprintln(c.Root().Writer, version)
					return nil
				},
			},
		},
	}
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

func cmdServe(ctx context.Context, addr string, stderr io.Writer) error {
	st, err := openStore()
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	h := httpapi.New(st, webembed.Dist())
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	ln, err := net.Listen("tcp", addr)
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

func cmdPull(ctx context.Context, tag string, stdout, stderr io.Writer) error {
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
	if err := svc.Pull(ctx, tag); err != nil {
		_ = st.Close()
		return err
	}
	fmt.Fprintf(stdout, "pulled %s into %s\n", tag, st.Path())
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
