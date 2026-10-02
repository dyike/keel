package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// InputGroupView composes a text input and independent leading/trailing views
// inside one field border. The input remains the owner of its value and events.
type InputGroupView struct {
	label, name    string
	input          *InputView
	prefix, suffix el.View
	disabled       bool
}

// InputGroup wraps input; nil creates an empty single-line input.
// Render the input through this group only, not separately in the same frame.
func InputGroup(label string, input *InputView) *InputGroupView {
	if input == nil {
		input = Input("")
	}
	return &InputGroupView{label: label, input: input}
}
func (v *InputGroupView) Prefix(view el.View) *InputGroupView      { v.prefix = view; return v }
func (v *InputGroupView) Suffix(view el.View) *InputGroupView      { v.suffix = view; return v }
func (v *InputGroupView) Value() string                            { return v.input.Value() }
func (v *InputGroupView) SetValue(s string)                        { v.input.SetValue(s) }
func (v *InputGroupView) OnChange(fn func(string)) *InputGroupView { v.input.OnChange(fn); return v }
func (v *InputGroupView) SetDisabled(on bool)                      { v.disabled = on }
func (v *InputGroupView) SetError(s string)                        { v.input.SetError(s) }
func (v *InputGroupView) Error() string                            { return v.input.Error() }
func (v *InputGroupView) FocusID() string                          { return v.input.FocusID() }
func (v *InputGroupView) setName(s string)                         { v.name = s }
func (v *InputGroupView) Render(cx *el.Context) el.Element {
	name := v.label
	if name == "" {
		name = v.name
	}
	if name == "" {
		name = v.input.a11y()
	}
	v.input.setName(name)
	id := autoID("inputgroup", v)
	disabled := v.disabled || v.input.disabled
	field := fieldFrame(id, cx.FocusWithin(id), v.Error() != "", disabled, v.input.readOnly).Role("group").Name(name)
	if v.prefix != nil {
		field.Child(el.Div().ID(id + "/prefix").NoShrink().Child(v.prefix.Render(cx)))
	}
	field.Child(el.Div().ID(id + "/field").Grow().MinW(el.Dp(24)).Child(v.input.render(cx, false)))
	if v.suffix != nil {
		field.Child(el.Div().ID(id + "/suffix").NoShrink().Child(v.suffix.Render(cx)))
	}
	root := el.Div().WFull().Gap(6).Disabled(disabled)
	label := v.label
	if label == "" {
		label = v.input.label
	}
	if label != "" {
		root.Child(el.Div().Name(label).OnClick(func() { cx.Focus(v.FocusID()) }).Child(el.Text(label).TextSize(13).TextColor(theme.Muted)))
	}
	root.Child(field)
	if v.Error() != "" {
		root.Child(el.Text(v.Error()).TextSize(12).TextColor(theme.Danger))
	}
	return root
}
