package kit

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
)

// The clickable/select semantic node must cover the frame, not just its text.
// Check all four padding corners and keyboard activation after dismissal.
func TestSelectWholeFieldTrigger(t *testing.T) {
	for _, plain := range []bool{false, true} {
		for _, height := range []float32{32, 36, 48} {
			v := Select("", "完全访问", "只读访问").Hint("访问权限").Size(height).Appearance(!plain)
			v.SetValue("完全访问")
			h := renderView(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().W(el.Dp(320)).Child(v.Render(cx), el.Div().Role("group").Name("after").H(el.Dp(1)))
			}), 320, 1)
			h.Frame()
			box, ok := node(h, "访问权限")
			if !ok {
				t.Fatal("select missing")
			}
			inset := 0
			if !plain {
				inset = 2
			}
			if box.Desc.Bounds.Dy() != int(height)-inset || bounds(h, "after").Min.Y != int(height) {
				t.Fatalf("plain=%v height=%v: trigger=%v after=%v", plain, height, box.Desc.Bounds, bounds(h, "after"))
			}
			for _, x := range []int{box.Desc.Bounds.Min.X + 1, box.Desc.Bounds.Max.X - 2} {
				for _, y := range []int{box.Desc.Bounds.Min.Y + 1, box.Desc.Bounds.Max.Y - 2} {
					h.Click(float32(x), float32(y))
					h.Frame()
					if !v.open {
						t.Fatalf("plain=%v padding corner (%d,%d) did not open", plain, x, y)
					}
					h.Key(key.NameEscape, 0)
					h.Frame()
				}
			}
			h.Key(key.NameSpace, 0)
			h.Frame()
			if !v.open {
				t.Fatal("Escape did not restore keyboard activation")
			}
		}
	}
}
