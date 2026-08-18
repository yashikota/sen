package webembed

import (
	"io/fs"
	"testing"
)

func TestDistIsEmbeddable(t *testing.T) {
	ents, err := fs.ReadDir(Dist(), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) == 0 {
		t.Fatal("embedded dist is empty; commit internal/webembed/dist/.keep")
	}
}
