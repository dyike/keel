package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestFormShowsNonTextErrorAndFocusesCheckbox(t *testing.T) {
	agree := Checkbox("", false)
	f := Form().Field("Terms", agree, func() string {
		if !agree.Value() {
			return "accept required"
		}
		return ""
	})
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return f.Render(c) })
	if f.Validate(cx) {
		t.Fatal("invalid checkbox accepted")
	}
	h.Frame()
	if !shown(h, "accept required") {
		t.Fatal("nontext error missing")
	}
	h.Key(key.NameSpace, 0)
	if !agree.Value() {
		t.Fatal("first error not focused")
	}
	if !f.Validate(cx) {
		t.Fatal("corrected checkbox rejected")
	}
	h.Frame()
	if shown(h, "accept required") {
		t.Fatal("corrected error retained")
	}
}

func TestFormAsyncSubmissionTokensErrorsAndFocus(t *testing.T) {
	in := Input("")
	in.SetValue("original")
	f := Form().Field("Name", in, nil)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return f.Render(c) })
	token := f.BeginSubmit(cx)
	if token == 0 || !f.Submitting() || f.BeginSubmit(cx) != 0 {
		t.Fatal("submit state")
	}
	h.Frame()
	clickClass(t, h, "Editor", "Name")
	h.Type("blocked")
	if in.Value() != "original" {
		t.Fatal("busy field editable")
	}
	if f.FinishSubmit(token+1, []string{"stale"}) || !f.Submitting() {
		t.Fatal("stale result accepted")
	}
	if !f.FinishSubmit(token, []string{"already exists"}) {
		t.Fatal("current result rejected")
	}
	h.Frame()
	h.Frame()
	if !shown(h, "already exists") || f.Submitting() {
		t.Fatal("async error missing")
	}
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	if in.Value() == "original" {
		t.Fatal("async error did not focus editor")
	}
	errs := f.Errors()
	errs[0] = "mutated"
	if f.Errors()[0] == "mutated" {
		t.Fatal("errors aliased")
	}
	token = f.BeginSubmit(cx)
	f.CancelSubmit()
	if f.FinishSubmit(token, nil) || f.Submitting() {
		t.Fatal("canceled result accepted")
	}
	token = f.BeginSubmit(cx)
	h.Frame()
	f.SetDisabled(true)
	h.Frame()
	if f.FinishSubmit(token, nil) || f.Submitting() {
		t.Fatal("disabled result accepted")
	}
}

func TestFormFlushesNumberDraftBeforeValidation(t *testing.T) {
	n := NumberInput("")
	f := Form().Field("Amount", n, func() string {
		if n.Value() != 7 {
			return "expected 7"
		}
		return ""
	})
	var accepted bool
	h := render(func(cx *el.Context) el.Element {
		return el.Div().Child(f.Render(cx), Button("Submit", func() { accepted = f.Validate(cx) }).Render(cx))
	})
	clickClass(t, h, "Editor", "Amount")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Type("7")
	click(t, h, "Submit")
	h.Frame()
	if !accepted || n.Value() != 7 {
		t.Fatal("validator saw stale number")
	}
}

func TestFormAncestorDisableCancelsSubmit(t *testing.T) {
	f := Form().Field("Name", Input(""), nil)
	disabled := false
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().Disabled(disabled).Child(f.Render(c)) })
	token := f.BeginSubmit(cx)
	h.Frame()
	disabled = true
	h.Frame()
	h.Frame()
	if f.Submitting() || f.FinishSubmit(token, nil) {
		t.Fatal("ancestor disable did not cancel")
	}
}

func TestFormSkipsReadonlyErrorFocusAndNamesSlider(t *testing.T) {
	readonly := Rating("", 5).ReadOnly()
	slider := Slider("", 0, 100)
	f := Form().Field("Average", readonly, func() string { return "locked" }).Field("Amount", slider, func() string { return "adjust" })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(300)).Child(f.Render(c)) })
	f.Validate(cx)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	if slider.Value() != 1 {
		t.Fatal("focus did not skip readonly rating")
	}
	n, ok := semanticNode(h, "slider:1")
	if !ok || n.Desc.Label != "Amount" {
		t.Fatal("slider form name")
	}
}
