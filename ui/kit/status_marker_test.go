package kit

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/el"
)

// statusMarkerLayout measures the content slot's actual layout box. Text
// semantics include font-dependent glyph padding and are not flex geometry.
func statusMarkerLayout(t *testing.T, v *StatusMarkerView, width, scale int) (*uitest.Harness, *image.Rectangle) {
	t.Helper()
	var content el.Element
	var box image.Rectangle
	v.PartStyle(StatusMarkerPartContent, func(d *el.DivEl) { content = d })
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		root := el.Div().WFull().Child(v.Render(cx))
		return root.Decorate(func(_ core.C, draw func()) {
			var ok bool
			box, ok = el.ElementBounds(root, content)
			if !ok {
				t.Fatal("missing content layout box")
			}
			draw()
		})
	}), width, scale)
	return h, &box
}

func TestStatusMarkerLayoutAndContent(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, align := range []el.Align{el.Start, el.Center, el.End} {
			v := StatusMarker("Today").Variant(StatusMarkerSeparator).Alignment(align)
			v.PartStyle(StatusMarkerPartSeparator, func(d *el.DivEl) { d.Name("line") })
			h, content := statusMarkerLayout(t, v, 240, scale)
			label := *content
			line := bounds(h, "line")
			if !shown(h, "Today") || label.Empty() || line.Empty() || label.Max.X > 240*scale {
				t.Fatal("missing/overflow", label, line)
			}
			if align == el.Start && line.Min.X < label.Max.X {
				t.Fatal("trailing line", label, line)
			}
			if align == el.End && line.Max.X > label.Min.X {
				t.Fatal("leading line", label, line)
			}
			v.ResetAlignment()
			h.Frame()
			label = *content
			if absInt((label.Min.X+label.Max.X)/2-120*scale) > 2*scale {
				t.Fatal("centered default", label)
			}
		}
		v := StatusMarker("This long message should wrap inside a narrow row and remain readable").Icon(Icon(IconInfo).Size(16)).Variant(StatusMarkerBorder)
		h, content := statusMarkerLayout(t, v, 140, scale)
		b := *content
		if !shown(h, v.text) || b.Max.X > 140*scale || b.Dy() < 24*scale {
			t.Fatal("wrapped text", b)
		}
	}
}

func TestStatusMarkerLoadingKeepsChildFocusAndDisabled(t *testing.T) {
	input := Input("reply")
	v := StatusMarker("").Loading(true).Content(input)
	h := renderView(v, 280, 1)
	clickClass(t, h, "Editor", "reply")
	h.Type("a")
	h.Key(key.NameRightArrow, 0)
	v.LoadingStyle(StatusMarkerLoadingStyleShimmer).Variant(StatusMarkerBorder).Icon(Icon(IconInfo))
	h.Frame()
	h.Key(key.NameDeleteBackward, 0)
	if input.Value() != "" {
		t.Fatal("loading switch lost focus", input.Value())
	}
	calls := 0
	v.Content(Button("Retry", func() { calls++ })).SetText("Failed")
	h.Frame()
	click(t, h, "Retry")
	if calls != 1 {
		t.Fatal("child action")
	}
	v.PartStyle(StatusMarkerPartRoot, func(d *el.DivEl) { d.Disabled(true) })
	h.Frame()
	click(t, h, "Retry")
	if calls != 1 {
		t.Fatal("inherited disabled")
	}
}
