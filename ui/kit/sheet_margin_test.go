package kit

import (
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func TestSheetMarginTopLayoutAndInput(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	for _, scale := range []int{1, 2} {
		for _, side := range []el.Side{el.Left, el.Right, el.Top, el.Bottom} {
			calls := 0
			s := Sheet(side, "Panel").Size(220).MarginTop(32).Body(text("Body")).Footer(Button("Apply", func() { calls++ }))
			s.SetValue(true)
			h := renderView(s, 400, scale)
			// Locate the dialog rather than its identically named title.
			panel := func() (int, int) {
				n, ok := semanticNode(h, "dialog")
				if !ok {
					t.Fatal("missing dialog")
				}
				return n.Desc.Bounds.Min.Y, n.Desc.Bounds.Max.Y
			}

			y, bottom := panel()
			expected := 32 * scale
			if side == el.Bottom {
				expected = (1000 - 220) * scale
			}
			if y != expected || bottom > 1000*scale {
				t.Fatal("panel constraints", side, scale, y, bottom)
			}
			click(t, h, "Apply")
			if calls != 1 {
				t.Fatal("footer hit area")
			}
			s.MarginTop(900)
			h.Frame()
			y, bottom = panel()
			if y < 900*scale || bottom > 1000*scale {
				t.Fatal("remaining height not constrained", side, y, bottom)
			}
			s.MarginTop(0)
			h.Frame()
			y, _ = panel()
			if side != el.Bottom && y != 0 {
				t.Fatal("reset margin", y)
			}
			s.MarginTop(64)
			h.Frame()
			h.Click(200*float32(scale), 16*float32(scale))
			h.Frame()
			if s.Value() {
				t.Fatal("top gap did not dismiss")
			}
		}
	}
}

func TestSheetExtremeMarginKeepsEscape(t *testing.T) {
	for _, side := range []el.Side{el.Left, el.Right, el.Top, el.Bottom} {
		s := Sheet(side, "Panel").MarginTop(math.MaxFloat32)
		s.MarginTop(-1).MarginTop(float32(math.NaN())).MarginTop(float32(math.Inf(1)))
		if s.marginTop != math.MaxFloat32 {
			t.Fatal("invalid margin accepted")
		}
		s.SetValue(true)
		h := renderView(s, 400, 1)
		h.Key(key.NameEscape, 0)
		h.Frame()
		if s.Value() {
			t.Fatal("clipped panel could not close")
		}
	}
}

func TestSheetTopAnimationStaysBelowMargin(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	c := &clock{now: time.Unix(100, 0)}
	s := Sheet(el.Top, "Panel").Size(220).MarginTop(80).Body(text("Body"))
	s.SetValue(true)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(s.Render(cx)) })
	c.advance(h, SheetSlide/2)
	// The still-sliding header must not capture a press in the reserved area.
	h.Click(200, 40)
	h.Frame()
	if s.Value() {
		t.Fatal("animation captured top gap")
	}
}
