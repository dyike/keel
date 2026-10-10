package el

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"gioui.org/io/key"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTextAreaScrollbarWheelDragAndEditing(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, readonly := range []bool{false, true} {
			t.Run(fmt.Sprintf("scale=%d/readonly=%v", scale, readonly), func(t *testing.T) {
				value := strings.Repeat("中文 line of text\n", 40)
				original := value
				var field *InputEl
				root := Root(viewFunc(func(*Context) Element {
					field = TextArea().ID("draft").Bind(&value).ReadOnly(readonly).AutoGrow(2, 4).W(Dp(240)).Scrollbars(ScrollbarAlways)
					return Div().Items(Start).Child(field)
				}))
				h := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
					root.Layout(gtx)
				})
				h.Frame()
				st := root.store.states[field.n.key]
				sc, ok := any(&st.editor).(editorViewport)
				if !ok {
					t.Skip("upstream Gio has native scrolling only")
				}
				if sc.ScrollBounds().Max.Y <= 0 || st.scrollContent <= st.scrollView {
					t.Fatal("long editor has no scroll range")
				}
				h.Scroll(40*float32(scale), 30*float32(scale), 60*float32(scale))
				h.Frame()
				if sc.ScrollOffset().Y <= 0 {
					t.Fatal("wheel did not scroll text")
				}
				beforeGutter := sc.ScrollOffset().Y
				h.Scroll(float32(field.n.size.X-16*scale), 30*float32(scale), 50*float32(scale))
				h.Frame()
				if sc.ScrollOffset().Y <= beforeGutter {
					t.Fatal("wheel over gutter did not scroll text")
				}
				sc.ScrollTo(image.Point{})
				st.editor.SetCaret(0, 3)
				h.Frame()
				start, end := st.editor.Selection()
				x := float32(field.n.size.X - 16*scale)
				h.Drag(x, 14*float32(scale), x, float32(field.n.size.Y-12*scale))
				h.Frame()
				if sc.ScrollOffset().Y < sc.ScrollBounds().Max.Y/2 {
					t.Fatal("thumb did not scroll text", sc.ScrollOffset(), sc.ScrollBounds())
				}
				s, e := st.editor.Selection()
				if s != start || e != end || value != original {
					t.Fatal("drag changed selection or text")
				}
				if !readonly {
					h.Key(key.NameEnd, key.ModShortcut)
					h.Type("tail")
					h.Frame()
					if !strings.Contains(value, "tail") {
						t.Fatal("drag lost editor focus")
					}
					if sc.ScrollOffset().Y < sc.ScrollBounds().Max.Y-5*scale {
						t.Fatal("editing no longer reveals caret")
					}
				}
				value = "short"
				h.Frame()
				h.Frame()
				if sc.ScrollOffset().Y != 0 || st.scrollContent > st.scrollView {
					t.Fatal("short text retained overflow")
				}
			})
		}
	}
}
