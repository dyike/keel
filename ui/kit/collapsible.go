package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// CollapsibleView owns one disclosure's state. Trigger and Content can be
// composed separately; Render combines them in a bordered panel.
type CollapsibleView struct {
	label          string
	body, heading  el.View
	open, disabled bool
	motion         disclosureMotion
	onChange       func(bool)
}

func Collapsible(label string, body el.View) *CollapsibleView {
	return &CollapsibleView{label: label, body: body}
}
func (v *CollapsibleView) Heading(view el.View) *CollapsibleView   { v.heading = view; return v }
func (v *CollapsibleView) Value() bool                             { return v.open }
func (v *CollapsibleView) SetValue(open bool)                      { v.open = open }
func (v *CollapsibleView) SetDisabled(on bool)                     { v.disabled = on }
func (v *CollapsibleView) OnChange(fn func(bool)) *CollapsibleView { v.onChange = fn; return v }
func (v *CollapsibleView) toggle() {
	if v.disabled {
		return
	}
	v.open = !v.open
	if v.onChange != nil {
		v.onChange(v.open)
	}
}
func (v *CollapsibleView) Trigger() el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		return disclosureTrigger(cx, autoID("collapsible", v)+"/trigger", v.label, v.heading, v.open, v.disabled, v.toggle)
	})
}
func (v *CollapsibleView) Content() el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		return disclosureContent(cx, autoID("collapsible", v)+"/content", autoID("collapsible", v)+"/trigger", v.body, v.open, v.disabled, &v.motion)
	})
}
func (v *CollapsibleView) Render(cx *el.Context) el.Element {
	return el.Div().Role("group").Items(el.Stretch).Rounded(8).Border(1, theme.Border).Bg(theme.Surface).Child(v.Trigger().Render(cx), v.Content().Render(cx))
}
