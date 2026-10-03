package kit

import (
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestButtonPointerKeyboardAndState(t *testing.T) {
	calls := 0
	v := Button("保存 Save 123", func() { calls++ })
	h := renderView(v, 300, 1)
	click(t, h, "保存 Save 123")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if calls != 3 {
		t.Fatalf("activation count %d", calls)
	}
	for _, loading := range []bool{false, true} {
		v.SetDisabled(!loading)
		v.SetLoading(loading)
		h.Frame()
		click(t, h, "保存 Save 123")
		h.Key(key.NameReturn, 0)
		if calls != 3 {
			t.Fatal("disabled/loading button activated")
		}
		n, ok := buttonNode(h, "保存 Save 123")
		if !ok || n.Desc.Disabled != !loading {
			t.Fatal("missing disabled semantics")
		}
	}
	v.SetLoading(false)
	v.SetDisabled(false)
	v.SetText("继续")
	h.Frame()
	if calls != 3 {
		t.Fatal("program setters triggered callback")
	}
	click(t, h, "继续")
	h.Frame()
	if calls != 4 {
		t.Fatalf("button did not recover: calls=%d bounds=%v", calls, buttonBounds(h, "继续"))
	}
}

func TestButtonLoadingPreservesBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, height := range []float32{28, 32, 40} {
			for _, withIcon := range []bool{false, true} {
				for _, label := range []string{"保存", "Save 123", "中 English 123"} {
					v := Button(label, func() {}).Size(height)
					if withIcon {
						v.Icon(IconPlus)
					}
					h := renderView(v, 300, scale)
					before := buttonBounds(h, label)
					v.SetLoading(true)
					h.Frame()
					if got := buttonBounds(h, label); got != before || got.Dy() != int(height)*scale {
						t.Fatalf("loading moved button: %v -> %v scale=%d height=%g icon=%v", before, got, scale, height, withIcon)
					}
					n, ok := semanticNode(h, "button:loading")
					if !ok || n.Desc.Disabled {
						t.Fatal("missing loading state")
					}
					v.SetLoading(false)
					h.Frame()
					if buttonBounds(h, label) != before {
						t.Fatal("loading restore moved button")
					}
				}
			}
		}
	}
}

func TestButtonInheritedDisableAndReadOnlyFrames(t *testing.T) {
	calls := 0
	disabled := true
	v := Button("action", func() { calls++ })
	root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }))
	h := uitest.New(root)
	click(t, h, "action")
	h.Key(key.NameSpace, 0)
	if calls != 0 {
		t.Fatal("parent disable bypassed")
	}
	disabled = false
	h.Frame()
	click(t, h, "action")
	if calls != 1 {
		t.Fatal("parent enable did not restore")
	}
	// Exercise the same view through disabled layouts without mutating its state.
	h = uitest.NewFunc(func(gtx core.C) { root.Layout(gtx.Disabled()); root.Layout(gtx) })
	click(t, h, "action")
	if calls != 2 {
		t.Fatal("readonly frame discarded interaction state")
	}
}

func TestButtonNarrowConstraint(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Button("很长的按钮 Button 123", nil).Icon(IconPlus)
		h := renderView(v, 100, scale)
		before := buttonBounds(h, "很长的按钮 Button 123")
		v.SetLoading(true)
		h.Frame()
		if got := buttonBounds(h, "很长的按钮 Button 123"); got != before || got.Dx() > 100*scale {
			t.Fatalf("narrow loading bounds %v before %v", got, before)
		}
	}
}

func buttonNode(h *uitest.Harness, label string) (input.SemanticNode, bool) {
	var found input.SemanticNode
	ok := false
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Class == semantic.Button && n.Desc.Label == label {
			found, ok = n, true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	return found, ok
}
func buttonBounds(h *uitest.Harness, label string) image.Rectangle {
	n, _ := buttonNode(h, label)
	return n.Desc.Bounds
}

func TestButtonAdditionalVariantsCompactAndContent(t *testing.T) {
	for _, variant := range []ButtonVariant{ButtonLink, ButtonText, ButtonSuccess, ButtonWarning, ButtonInfo} {
		for _, scale := range []int{1, 2} {
			calls := 0
			v := Button("Action", func() { calls++ }).Variant(variant).Outline(true)
			h := renderView(v, 200, scale)
			normal := buttonBounds(h, "Action")
			v.Compact(true)
			h.Frame()
			compact := buttonBounds(h, "Action")
			if compact.Dy() != normal.Dy() || (variant != ButtonLink && compact.Dx() >= normal.Dx()) {
				t.Fatalf("compact %d: %v %v", variant, normal, compact)
			}
			v.Content(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().Row().Gap(6).Child(el.Text("Custom"), Icon(IconPlus).Render(cx))
			}))
			h.Frame()
			before := buttonBounds(h, "Action")
			click(t, h, "Action")
			h.Key(key.NameSpace, 0)
			if calls != 2 || !shown(h, "Custom") {
				t.Fatal("custom content activation")
			}
			v.SetLoading(true)
			h.Frame()
			if buttonBounds(h, "Action") != before {
				t.Fatal("loading custom content moved button")
			}
			click(t, h, "Action")
			h.Key(key.NameReturn, 0)
			if calls != 2 {
				t.Fatal("loading custom action activated")
			}
			v.SetLoading(false)
			h.Frame()
			h.Key(key.NameReturn, 0)
			if calls != 3 {
				t.Fatal("loading lost keyboard focus")
			}
			v.Content(nil).Compact(false)
			h.Frame()
			if buttonBounds(h, "Action") != normal || shown(h, "Custom") {
				t.Fatal("content/compact reset")
			}
		}
	}
	for _, height := range []float32{24, 28, 32, 40} {
		v := Button("", nil).Name("Icon").Icon(IconPlus).Size(height)
		h := renderView(v, 100, 1)
		b := buttonBounds(h, "Icon")
		if b.Dx() != int(height) || b.Dy() != int(height) {
			t.Fatal("icon-only button not square", b)
		}
	}
}
