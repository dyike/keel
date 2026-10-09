package kit

import (
	"fmt"
	"slices"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
)

func TestTableSelectionModesKeepEmbeddedEditorFocus(t *testing.T) {
	for _, mode := range []string{"cell", "column"} {
		for _, scale := range []int{1, 2} {
			for _, split := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%dx/split=%v", mode, scale, split), func(t *testing.T) {
					v := Table(Col("A").Width(140).Cell(func(cx *el.Context, row int) el.Element { return el.Input().ID("nested-editor").Name("editor") }), Col("B").Width(140)).Height(100)
					if mode == "cell" {
						v.CellSelect()
					} else {
						v.ColumnSelect()
					}
					v.SetRows([][]string{{"a", "b"}})
					h := renderView(v, 400, scale)
					if split {
						r := bounds(h, "editor")
						p := f32.Pt(float32(r.Min.X+5*scale), float32(r.Min.Y+5*scale))
						h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
						h.Frame()
						h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
						h.Frame()
					} else {
						clickClass(t, h, "Editor", "editor")
					}
					h.Type("typed")
					if desc(h, "editor") != "typed" {
						t.Fatal("editor lost focus", desc(h, "editor"))
					}
					h.Key(key.NameRightArrow, 0)
					if mode == "column" && !slices.Equal(v.SelectedColumns(), []int{0}) {
						t.Fatal("editor arrow changed column selection")
					}
					click(t, h, "B")
					h.Key(key.NameLeftArrow, 0)
					if mode == "column" && !slices.Equal(v.SelectedColumns(), []int{0}) {
						t.Fatal("header keyboard navigation broken", v.SelectedColumns())
					}
					if mode == "cell" && !slices.Equal(v.SelectedCells(), []TableCell{{0, 0}}) {
						t.Fatal("cell keyboard navigation broken", v.SelectedCells())
					}
					v.MoveColumn(0, 1)
					h.Frame()
					if desc(h, "editor") != "typed" {
						t.Fatal("move lost editor contents")
					}
				})
			}
		}
	}
}
