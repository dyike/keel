package widget

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTextAreaAutoHeightFitsEveryVisibleLine(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		for _, src := range []string{"第一行\n第二行\n第三行\n第四行", "第一行\n\n第三行\n第四行"} {
			t.Run(fmt.Sprintf("%q/%gx", src, scale), func(t *testing.T) {
				area := TextArea("").AutoHeight(1, 4)
				area.SetValue(src)
				var dims core.D
				uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
					dims = area.Layout(gtx)
				})
				regions := area.editor.Regions(0, area.editor.Len(), nil)
				viewport := image.Rect(0, 0, dims.Size.X-int(20*scale), dims.Size.Y-int(16*scale))
				if len(regions) != 4 {
					t.Fatalf("four-line viewport exposes %d lines", len(regions))
				}
				for _, region := range regions {
					if region.Bounds.Min.Y < 0 || region.Bounds.Max.Y > viewport.Max.Y {
						t.Fatalf("line %v is clipped by viewport %v", region.Bounds, viewport)
					}
				}
			})
		}
	}
}

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
