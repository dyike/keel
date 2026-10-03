package kit

import (
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestFormGridSpansStartsVisibilityAndState(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a, b, c := Input(""), Input(""), Input("")
		f := Form().Columns(3).VerticalLabels(true).
			FieldWithOptions("Alpha", a, nil, FormFieldOptions{Required: true, Description: "help", ColSpan: 2}).
			Field("Beta", b, nil).
			FieldWithOptions("Gamma", c, nil, FormFieldOptions{ColStart: 2, ColSpan: 2}).
			Footer(Button("Save", nil))
		root := el.Root(f)
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints = layout.Exact(image.Pt(600*scale, 600*scale))
			root.Layout(gtx)
		})
		ab, bb, cb := formEditorBounds(h, "Alpha"), formEditorBounds(h, "Beta"), formEditorBounds(h, "Gamma")
		if ab.Min.Y != bb.Min.Y || ab.Max.X > bb.Min.X || cb.Min.Y <= ab.Max.Y || cb.Min.X <= ab.Min.X || cb.Max.X != bb.Max.X {
			t.Fatal("grid", ab, bb, cb)
		}
		if !shown(h, "help") || !shown(h, "*") || bounds(h, "Save").Max.X != 600*scale {
			t.Fatal("description/required/footer")
		}
		clickClass(t, h, "Editor", "Gamma")
		h.Type("retained")
		h.Frame()
		f.SetFieldOptions(0, FormFieldOptions{Hidden: true})
		h.Frame()
		if shown(h, "Alpha") || c.Value() != "retained" {
			t.Fatal("hide/state")
		}
		f.Columns(1).SetFieldOptions(0, FormFieldOptions{})
		h.Frame()
		if bounds(h, "Gamma").Max.X > 600*scale || c.Value() != "retained" {
			t.Fatal("column clamp/state")
		}
	}
}

func TestFormHiddenValidationAndAsyncInvalidation(t *testing.T) {
	hiddenCalls := 0
	f := Form().Field("Visible", Input(""), nil).FieldWithOptions("Conditional", Input(""), func() string { hiddenCalls++; return "invalid" }, FormFieldOptions{Hidden: true})
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return f.Render(c) })
	if !f.Validate(cx) || hiddenCalls != 0 {
		t.Fatal("hidden validation")
	}
	token := f.BeginSubmit(cx)
	if token == 0 {
		t.Fatal("begin")
	}
	f.SetFieldOptions(1, FormFieldOptions{Required: true})
	if f.Submitting() || f.FinishSubmit(token, []string{"stale"}) {
		t.Fatal("visibility did not invalidate submit")
	}
	h.Frame()
	if f.Validate(cx) || hiddenCalls != 1 {
		t.Fatal("visible validation")
	}
	f.SetFieldOptions(1, FormFieldOptions{Hidden: true})
	if f.Errors()[1] != "" {
		t.Fatal("hidden error")
	}
	h.Frame()
	if shown(h, "invalid") || f.SetFieldOptions(99, FormFieldOptions{}) {
		t.Fatal("hidden rendering/invalid index")
	}
}

func formEditorBounds(h *uitest.Harness, name string) image.Rectangle {
	var result image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Label == name && n.Desc.Class.String() == "Editor" {
			result = n.Desc.Bounds
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	return result
}
