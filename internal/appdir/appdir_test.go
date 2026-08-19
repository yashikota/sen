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
