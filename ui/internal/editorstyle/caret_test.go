package editorstyle

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestCaretFollowsEditingAndViewport(t *testing.T) {
	for _, tc := range []struct {
		name, src       string
		multi, password bool
	}{
		{"Chinese", "对对对", false, false}, {"Latin", "Agyp", false, false}, {"mixed", "Go语言123", false, false},
		{"password", "秘密Password", false, true}, {"multiline", "第一行\nSecond行", true, false}, {"wrapped", "中文English 中文English 中文English 中文English 中文English", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ed widget.Editor
			var caret Caret
			ed.SingleLine = !tc.multi
			if tc.password {
				ed.Mask = '•'
			}
			now := time.Now()
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				gtx.Constraints = layout.Exact(image.Pt(120, 90))
				for {
					_, ok := ed.Update(gtx)
					if !ok {
						break
					}
				}
				style := material.Editor(theme.Material, &ed, "提示")
				caret.Layout(gtx, style, theme.Material.Shaper)
				if ed.ReadOnly {
					t.Fatal("paint leaked ReadOnly into input")
				}
			})
			h.Click(8, 8)
			h.Type(tc.src)
			if ed.Text() != tc.src {
				t.Fatalf("typing lost: %q", ed.Text())
			}
			h.Key(key.NameEnd, key.ModShortcut)
			if caret.bounds.Empty() {
				t.Fatal("focused editor has no caret")
			}
			pos := ed.CaretCoords().Round()
			if caret.bounds.Min.X > pos.X || caret.bounds.Max.X < pos.X || caret.bounds.Min.Y != pos.Y+caret.top {
				t.Fatalf("caret %v diverges from editor %v", caret.bounds, pos)
			}
			if caret.bounds.Dy() > 19 || caret.bounds.Dy() < 12 {
				t.Fatalf("unexpected ink height %v", caret.bounds)
			}
			h.Key(key.NameLeftArrow, 0)
			if ed.CaretCoords().Round() == pos {
				t.Fatal("keyboard caret did not move")
			}
			h.Type("新")
			if ed.Text() == tc.src {
				t.Fatal("input stopped after moving caret")
			}
		})
	}
}

func TestCaretScaleFocusAndReadOnly(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		t.Run(string(rune('0'+int(scale))), func(t *testing.T) {
			var ed widget.Editor
			var caret Caret
			ed.SingleLine = true
			enabled := true
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
				if !enabled {
					gtx = gtx.Disabled()
				}
				for {
					_, ok := ed.Update(gtx)
					if !ok {
						break
					}
				}
				caret.Layout(gtx, material.Editor(theme.Material, &ed, "提示"), theme.Material.Shaper)
			})
			if !caret.bounds.Empty() {
				t.Fatal("unfocused caret visible")
			}
			h.Click(8, 8)
			h.Type("中文g")
			if caret.bounds.Empty() {
				t.Fatal("no caret after typing")
			}
			if caret.bottom-caret.top > int(19*scale) {
				t.Fatal("caret includes font leading")
			}
			ed.ReadOnly = true
			h.Frame()
			h.Type("不应输入")
			if ed.Text() != "中文g" || !ed.ReadOnly || !caret.bounds.Empty() {
				t.Fatal("read-only state changed")
			}
			ed.ReadOnly = false
			enabled = false
			h.Frame()
			if !caret.bounds.Empty() {
				t.Fatal("disabled caret visible")
			}
		})
	}
}

func TestCaretPreservesCompositionAndUndo(t *testing.T) {
	var ed widget.Editor
	var caret Caret
	ed.SingleLine = true
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Exact(image.Pt(180, 60))
		for {
			_, ok := ed.Update(gtx)
			if !ok {
				break
			}
		}
		caret.Layout(gtx, material.Editor(theme.Material, &ed, ""), theme.Material.Shaper)
	})
	h.Click(2, 8)
	// Platforms update composition text and the selection separately.
	h.Router.Queue(key.CompositionEvent{Start: 0, End: 3}, key.EditEvent{Range: key.Range{}, Text: "dui"}, key.SelectionEvent{Start: 3, End: 3})
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 3}, Text: "对"}, key.SelectionEvent{Start: 1, End: 1}, key.CompositionEvent{Start: -1, End: -1})
	h.Frame()
	if ed.Text() != "对" {
		t.Fatalf("composition lost: %q", ed.Text())
	}
	start, end := ed.Selection()
	if start != 1 || end != 1 {
		t.Fatalf("IME caret moved: %d %d", start, end)
	}
	h.Key("Z", key.ModShortcut)
	if ed.Text() == "对" {
		t.Fatal("undo stopped working")
	}
	h.Key("Z", key.ModShortcut|key.ModShift)
	if ed.Text() != "对" {
		t.Fatal("redo stopped working")
	}
}

// Compare painted pixels, not just font metrics: an empty editor can choose a
// different fallback baseline from the separately painted placeholder.
func TestEmptyCaretAlignsWithPlaceholderPixels(t *testing.T) {
	for _, hint := range []string{"收起后保留输入内容", "Placeholder text", "Go 输入名称"} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/%dx", hint, scale), func(t *testing.T) {
				size := image.Pt(280*scale, 55*scale)
				win, err := headless.NewWindow(size.X, size.Y)
				if err != nil {
					t.Fatal(err)
				}
				defer win.Release()
				var ed widget.Editor
				ed.SingleLine = true
				var caret Caret
				var ops op.Ops
				var router input.Router
				for frame := 0; frame < 3; frame++ {
					ops.Reset()
					gtx := layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(), Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}}
					paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					for {
						_, ok := ed.Update(gtx)
						if !ok {
							break
						}
					}
					if frame == 1 {
						gtx.Execute(key.FocusCmd{Tag: &ed})
					}
					style := material.Editor(theme.Material, &ed, hint)
					style.Color = color.NRGBA{R: 255, A: 255}
					style.HintColor = color.NRGBA{B: 255, A: 255}
					// Like a real field: the editor sits inside 10dp of padding,
					// where the empty caret is drawn.
					pad := 10 * scale
					off := op.Offset(image.Pt(pad, 0)).Push(&ops)
					inner := gtx
					inner.Constraints = layout.Exact(image.Pt(size.X-pad, size.Y))
					caret.Layout(inner, style, theme.Material.Shaper)
					off.Pop()
					router.Frame(&ops)
				}
				if err := win.Frame(&ops); err != nil {
					t.Fatal(err)
				}
				img := image.NewRGBA(image.Rectangle{Max: size})
				if err := win.Screenshot(img); err != nil {
					t.Fatal(err)
				}
				redTop, redBottom, blueTop, blueBottom := size.Y, -1, size.Y, -1
				redRight, blueLeft := -1, size.X
				for y := 0; y < size.Y; y++ {
					for x := 0; x < size.X; x++ {
						p := img.RGBAAt(x, y)
						if p.R > 180 && p.G < 100 && p.B < 100 {
							redTop = min(redTop, y)
							redBottom = max(redBottom, y)
							redRight = max(redRight, x)
						}
						if p.B > 180 && p.R < 100 && p.G < 100 {
							blueTop = min(blueTop, y)
							blueBottom = max(blueBottom, y)
							blueLeft = min(blueLeft, x)
						}
					}
				}
				if redBottom < 0 || blueBottom < 0 {
					t.Fatal("missing caret or placeholder pixels")
				}
				// The caret must keep clear of the hint's first glyph: a caret
				// within a pixel or two of it reads as a stroke (搜 becomes 锼).
				if gap := blueLeft - redRight - 1; gap < 2*scale {
					t.Fatalf("caret ends at x=%d, %dpx from hint ink at x=%d; want at least %d", redRight, gap, blueLeft, 2*scale)
				}
				delta := (redTop + redBottom) - (blueTop + blueBottom)
				if delta < -2 || delta > 2 {
					t.Fatalf("painted centers differ: caret y=%d..%d hint y=%d..%d", redTop, redBottom, blueTop, blueBottom)
				}
			})
		}
	}
}
