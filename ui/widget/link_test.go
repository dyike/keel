package widget

import (
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestLinkClick(t *testing.T) {
	n := 0
	h := uitest.New(Link("文档", func() { n++ }))
	h.Click(5, 5)
	if n != 1 {
		t.Fatalf("clicked %d times", n)
	}
}
