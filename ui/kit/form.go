package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"strings"
)

// Validatable is a control that can show a validation error and be focused:
// Input, TextArea, Select, NumberInput, OtpInput, TimeField, Combobox and
// DatePicker.
type Validatable interface {
	el.View
	SetError(msg string)
	FocusID() string
}

type formField struct {
	label    string
	control  el.View
	validate func() string
}

// FormView lays out labelled fields in two columns and validates them on
// Submit. A validator returns an error message, or "" when the value is fine;
// the message appears under the field until the next Submit or the user
// changes the field.
//
//	f := kit.Form().
//	    Field("客户", customer, func() string { return kit.Required(customer.Value(), "请填写客户") }).
//	    Field("金额", amount, nil)
//	kit.Button("创建", func() { if f.Validate(cx) { create() } })
type FormView struct {
	fields     []formField
	labelWidth float32
}

func Form() *FormView { return &FormView{labelWidth: 72} }

// Field adds a row. validate may be nil; then control needs no Validatable.
func (v *FormView) Field(label string, control el.View, validate func() string) *FormView {
	v.fields = append(v.fields, formField{label, control, validate})
	return v
}

// LabelWidth sets the label column width in dp, 72 by default.
func (v *FormView) LabelWidth(dp float32) *FormView {
	if dp > 0 {
		v.labelWidth = dp
	}
	return v
}

// Validate runs every validator, shows their messages, and focuses the first
// invalid field. It reports whether all passed. Call it from a callback.
func (v *FormView) Validate(cx *el.Context) bool {
	ok := true
	for _, f := range v.fields {
		c, can := f.control.(Validatable)
		if f.validate == nil {
			continue
		}
		msg := f.validate()
		if can {
			c.SetError(msg)
		}
		if msg != "" && ok {
			ok = false
			if can {
				cx.Focus(c.FocusID())
			}
		}
	}
	return ok
}

// Required returns msg when s is empty or blank, else "".
func Required(s, msg string) string {
	if strings.TrimSpace(s) == "" {
		return msg
	}
	return ""
}

func (v *FormView) Render(cx *el.Context) el.Element {
	form := el.Div().Role("form").Gap(14).Items(el.Stretch)
	for _, f := range v.fields {
		if n, ok := f.control.(interface{ setName(string) }); ok {
			n.setName(f.label) // the row label names a control that has none
		}
		form.Child(el.Div().Row().Gap(12).Items(el.Start).Child(
			el.Div().W(el.Dp(v.labelWidth)).NoShrink().Items(el.End).Pt(8).Child(el.Text(f.label).TextColor(theme.Muted)),
			el.Div().Grow().Items(el.Stretch).Child(f.control.Render(cx)),
		))
	}
	return form
}
