package kit

import (
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestStatusMarkerLayoutAndContent(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, align := range []el.Align{el.Start, el.Center, el.End} {
			v := StatusMarker("Today").Variant(StatusMarkerSeparator).Alignment(align)
			v.PartStyle(StatusMarkerPartSeparator, func(d *el.DivEl) { d.Name("line") })
			h := renderView(v, 240, scale)
			label := bounds(h, "Today")
			line := bounds(h, "line")
			if label.Empty() || line.Empty() || label.Max.X > 240*scale {
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
			label = bounds(h, "Today")
			if absInt((label.Min.X+label.Max.X)/2-120*scale) > 2*scale {
				t.Fatal("centered default", label)
			}
		}
		v := StatusMarker("This long message should wrap inside a narrow row and remain readable").Icon(Icon(IconInfo).Size(16)).Variant(StatusMarkerBorder)
		h := renderView(v, 140, scale)
		b := bounds(h, v.text)
		if b.Max.X > 140*scale || b.Dy() < 24*scale {
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
