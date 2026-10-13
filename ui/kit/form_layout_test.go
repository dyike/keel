package kit

import (
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/unit"
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

// Horizontal labels line up with the control's first line of text, whatever
// frame or padding sits around it: a field, a checkbox or switch row, the
// first of several options and the first line of a text area.
func TestFormLabelsLineUpWithControlText(t *testing.T) {
	f := Form().
		Field("输入", Input("").Placeholder("请输入"), nil).
		Field("选择", Select("", "选项一", "选项二"), nil).
		Field("勾选", Checkbox("记住", true), nil).
		Field("开关", Switch("自动", true), nil).
		Field("单选", RadioGroup("", "甲", "乙"), nil).
		Field("多行", TextArea("").Placeholder("多行内容"), nil)
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Child(f.Render(cx)) })))
	top := func(label string) int {
		for _, n := range h.Router.AppendSemantics(nil) {
			if n.Desc.Label == label && n.Desc.Class.String() == "Unknown" && len(n.Children) == 0 {
				return n.Desc.Bounds.Min.Y
			}
		}
		t.Fatalf("no text %q", label)
		return 0
	}
	for _, pair := range [][2]string{{"输入", "请输入"}, {"选择", "请选择"}, {"勾选", "记住"}, {"开关", "自动"}, {"单选", "甲"}, {"多行", "多行内容"}} {
		if a, b := top(pair[0]), top(pair[1]); a != b {
			t.Errorf("label %q at y=%d, control text %q at y=%d", pair[0], a, pair[1], b)
		}
	}
}
