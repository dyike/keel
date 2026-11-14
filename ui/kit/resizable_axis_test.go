package kit

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestResizableAxisChangeRetainsContent(t *testing.T) {
	input := Input("Editor")
	input.SetValue("retained input")
	calls := 0
	v := Resizable(input, nil).Min(20, 20).OnChange(func(float32) { calls++ })
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(180)).H(el.Dp(500)).Child(v.Render(cx))
	}), 500, 1)
	h.Frame()
	v.SetVertical(true).SetValue(300)
	h.Frame()
	h.Frame()
	if v.Value() != 300 {
		t.Fatal("new axis clamped against old width", v.Value())
	}
	b := bounds(h, locale.Current().Resize)
	if b.Dy() != handleSize || b.Dx() != 180 {
		t.Fatal("vertical handle", b)
	}
	v.SetVertical(false)
	h.Frame()
	h.Frame()
	if b = bounds(h, locale.Current().Resize); b.Dx() != handleSize || b.Dy() != 500 {
		t.Fatal("horizontal handle", b)
	}
	if input.Value() != "retained input" || calls != 0 {
		t.Fatal("axis change lost content or notified", input.Value(), calls)
	}
}

func TestResizableGroupAxisChangeRetainsHiddenPanels(t *testing.T) {
	input := Input("Editor")
	input.SetValue("retained input")
	calls := 0
	v := ResizableGroup(ResizablePanel{ID: "editor", Content: input, Size: 100}, ResizablePanel{ID: "preview", Size: 100}, ResizablePanel{ID: "hidden", Size: 77, Hidden: true}).OnChange(func(map[string]float32) { calls++ })
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(180)).H(el.Dp(500)).Child(v.Render(cx))
	}), 500, 1)
	h.Frame()
	v.SetVertical(true).SetSizes(map[string]float32{"editor": 300, "preview": 194})
	h.Frame()
	h.Frame()
	if v.Sizes()["editor"] != 300 || v.Sizes()["hidden"] != 77 {
		t.Fatal("new axis sizes", v.Sizes())
	}
	v.SetVertical(false)
	h.Frame()
	h.Frame()
	if input.Value() != "retained input" || v.Sizes()["hidden"] != 77 || calls != 0 {
		t.Fatal("axis change lost state or notified", v.Sizes(), calls)
	}
}
