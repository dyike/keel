package widget

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"strings"
	"testing"
)

func TestTextAreaAutoHeight(t *testing.T) {
	area := TextArea("").AutoHeight(1, 3)
	var dims core.D
	h := uitest.NewFunc(func(gtx core.C) { dims = area.Layout(gtx) })
	one := dims.Size.Y
	area.SetValue("一\n二\n三")
	h.Frame()
	three := dims.Size.Y
	if three <= one {
		t.Fatalf("textarea did not grow: %d -> %d", one, three)
	}
	area.SetValue("一\n二\n三\n四\n五\n六")
	h.Frame()
	if dims.Size.Y != three {
		t.Fatalf("textarea exceeded max lines: %d -> %d", three, dims.Size.Y)
	}
	area.SetValue("一")
	h.Frame()
	if dims.Size.Y != one {
		t.Fatal("textarea did not shrink")
	}
	// A long single paragraph must grow through soft wrapping as well.
	area.SetValue(strings.Repeat("中文自动换行 ", 50))
	h.Frame()
	if dims.Size.Y <= one || dims.Size.Y > three {
		t.Fatal("soft wraps ignored or exceeded maximum")
	}
}
