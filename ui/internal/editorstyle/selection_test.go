package editorstyle

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSelectionInkAlignment(t *testing.T) {
	for _, tc := range []struct {
		src                 string
		multiline, password bool
	}{
		{"第一行", false, false}, {"123456", false, false}, {"Agyp", false, false}, {"Go语言123", false, false},
		{"第一行\n\n第三行\n第四行", true, false}, {strings.Repeat("中文", 15), true, false}, {"密码123", false, true},
	} {
		src := tc.src
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("%q/%dx", src, scale), func(t *testing.T) {
				viewport := image.Pt(110*scale, 180*scale)
				win, err := headless.NewWindow(viewport.X, viewport.Y)
				if err != nil {
					t.Fatal(err)
				}
				defer win.Release()
				var ed widget.Editor
				ed.SingleLine = !tc.multiline
				if tc.password {
					ed.Mask = '•'
				}
				ed.PaintSelectionWhenUnfocused = true
				ed.SetText(src)
				ed.SetCaret(0, ed.Len())
				var caret Caret
				var ops op.Ops
				var router input.Router
				gtx := layout.Context{Ops: &ops, Source: router.Source(), Now: time.Now(), Constraints: layout.Exact(viewport), Metric: unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}}
				paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
				style := material.Editor(theme.Material, &ed, "")
				style.Color = color.NRGBA{R: 255, A: 255}
				style.SelectionColor = color.NRGBA{B: 255, A: 255}
				caret.Layout(gtx, style, theme.Material.Shaper)
				if err := win.Frame(&ops); err != nil {
					t.Fatal(err)
				}
				img := image.NewRGBA(image.Rectangle{Max: viewport})
				if err := win.Screenshot(img); err != nil {
					t.Fatal(err)
				}
				// Compare each contiguous highlighted row to its visible text.
				rows := 0
				for y := 0; y < viewport.Y; {
					top, bottom, inkTop, inkBottom := -1, -1, viewport.Y, -1
					for ; y < viewport.Y; y++ {
						selected := false
						for x := 0; x < viewport.X; x++ {
							p := img.RGBAAt(x, y)
							if p.B > 180 && p.R < 80 && p.G < 80 {
								selected = true
							}
							if int(p.R) > int(p.G)+40 && int(p.R) > int(p.B)+40 {
								inkTop, inkBottom = min(inkTop, y), max(inkBottom, y)
							}
						}
						if selected {
							if top < 0 {
								top = y
							}
							bottom = y
						} else if top >= 0 {
							y++
							break
						}
					}
					if top < 0 {
						break
					}
					rows++
					if inkBottom < 0 {
						t.Fatal("highlight has no visible text")
					}
					if delta := inkTop + inkBottom - top - bottom; delta < -2 || delta > 2 {
						t.Fatalf("text and selection centers differ: ink %d..%d selection %d..%d", inkTop, inkBottom, top, bottom)
					}
				}
				if rows == 0 {
					t.Fatal("no selection rendered")
				}
				if tc.multiline && rows < 2 {
					t.Fatal("multiline selection lost rows")
				}
			})
		}
	}
}

func TestSelectionSurvivesScrollAndReplace(t *testing.T) {
	var ed widget.Editor
	var caret Caret
	src := strings.Repeat("第一行\n第二行\n", 20)
	ed.SetText(src)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Exact(image.Pt(160, 70))
		for {
			if _, ok := ed.Update(gtx); !ok {
				break
			}
		}
		caret.Layout(gtx, material.Editor(theme.Material, &ed, ""), theme.Material.Shaper)
	})
	h.Click(5, 5)
	h.Key("A", key.ModShortcut)
	before := ed.CaretCoords()
	h.Scroll(30, 30, 35)
	if ed.CaretCoords() == before {
		t.Fatal("selected editor did not scroll")
	}
	if ed.SelectedText() != src || len(caret.regions) == 0 {
		t.Fatal("scroll lost selection")
	}
	// The platform sends the selected rune range with the replacement event.
	start, end := ed.Selection()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: start, End: end}, Text: "替换"}, key.SelectionEvent{Start: 2, End: 2})
	h.Frame()
	if ed.Text() != "替换" || len(caret.regions) != 0 {
		t.Fatalf("replacement left stale selection: text=%q regions=%d", ed.Text(), len(caret.regions))
	}
	h.Key("Z", key.ModShortcut)
	if ed.Text() != src {
		t.Fatal("selection replacement broke undo")
	}
}
