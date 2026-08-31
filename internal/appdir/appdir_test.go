package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeDefaultIsDotSenInCwd(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SEN_HOME", "")
	got, err := Home()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, DefaultDir)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestHomeRespectsSEN_HOME(t *testing.T) {
	custom := t.TempDir()
	t.Setenv("SEN_HOME", custom)
	got, err := Home()
	if err != nil {
		t.Fatal(err)
	}
	if got != custom {
		t.Fatalf("got %q want %q", got, custom)
	}
}

func TestHomeCleansSEN_HOME(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEN_HOME", dir+string(os.PathSeparator)+"."+string(os.PathSeparator))
	got, err := Home()
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(dir+"/.") {
		t.Fatalf("got %q", got)
	}
}

func TestInitializedAndMarker(t *testing.T) {
	dir := t.TempDir()
	if Initialized(dir) {
		t.Fatal("empty dir is not initialized")
	}
	if Marker(dir) != filepath.Join(dir, "workspace.toml") {
		t.Fatalf("marker %s", Marker(dir))
	}
	if err := os.WriteFile(Marker(dir), []byte("name = \"sen\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Initialized(dir) {
		t.Fatal("workspace.toml should initialize")
	}
}

func TestInitializedAcceptsLegacyYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte("name: sen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Initialized(dir) {
		t.Fatal("legacy yaml still counts as initialized")
	}
}
