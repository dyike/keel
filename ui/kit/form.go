package kit

import (
	"slices"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
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
	err      string
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
	fields         []formField
	actions        []el.View
	labelWidth     float32
	disabled, busy bool
	request        uint64
	focusErrors    bool
}

func Form() *FormView { return &FormView{labelWidth: 72} }

// Field adds a row. validate may be nil; then control needs no Validatable.
func (v *FormView) Field(label string, control el.View, validate func() string) *FormView {
	v.CancelSubmit()
	v.fields = append(v.fields, formField{label: label, control: control, validate: validate})
	return v
}

// Actions sets the buttons below the fields, aligned with the control column.
// They stay usable while a submission is pending, so one can cancel it.
func (v *FormView) Actions(views ...el.View) *FormView {
	v.actions = slices.Clone(views)
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
	v.CancelSubmit()
	// Commit editor drafts before observing values in validators.
	for _, f := range v.fields {
		if c, ok := f.control.(interface{ commitForm() }); ok {
			if target, ok := f.control.(interface{ FocusID() string }); ok && !cx.Enabled(target.FocusID()) {
				continue
			}
			c.commitForm()
		}
	}
	ok := true
	for i := range v.fields {
		msg := ""
		if v.fields[i].validate != nil {
			msg = v.fields[i].validate()
		}
		v.setError(i, msg)
		if msg != "" {
			ok = false
		}
	}
	if !ok {
		v.focusFirstError(cx)
	}
	return ok
}
func (v *FormView) setError(i int, msg string) {
	v.fields[i].err = msg
	if c, ok := v.fields[i].control.(interface{ SetError(string) }); ok {
		c.SetError(msg)
	}
}
func (v *FormView) fieldID(i int) string { return autoID("form", v) + "/field/" + strconv.Itoa(i) }
func (v *FormView) focusFirstError(cx *el.Context) bool {
	for i, f := range v.fields {
		if f.err == "" {
			continue
		}
		id := v.fieldID(i)
		if c, ok := f.control.(interface{ FocusID() string }); ok {
			id = c.FocusID()
		}
		if id != "" && cx.Enabled(id) {
			cx.Focus(id)
			return true
		}
	}
	return false
}

// BeginSubmit runs synchronous validators and freezes fields while an async
// validation/submission is in flight. Zero means invalid, disabled or busy.
// Capture field values here, then return from the UI callback before doing I/O.
func (v *FormView) BeginSubmit(cx *el.Context) uint64 {
	if v.disabled || v.busy || !cx.Enabled(autoID("form", v)) {
		return 0
	}
	if !v.Validate(cx) {
		return 0
	}
	v.request++
	v.busy = true
	return v.request
}

// FinishSubmit accepts errors in Field order (nil means success). Call under
// core.Update for background work. Stale/canceled tokens are ignored.
func (v *FormView) FinishSubmit(token uint64, errors []string) bool {
	if !v.busy || token != v.request || len(errors) > len(v.fields) {
		return false
	}
	v.busy = false
	for i := range v.fields {
		msg := ""
		if i < len(errors) {
			msg = errors[i]
		}
		v.setError(i, msg)
		if msg != "" {
			v.focusErrors = true
		}
	}
	return true
}
func (v *FormView) Submitting() bool { return v.busy }

// CancelSubmit invalidates pending results and restores editing. Call before
// programmatically replacing data during a submission.
func (v *FormView) CancelSubmit() { v.request++; v.busy = false; v.focusErrors = false }
func (v *FormView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.CancelSubmit()
	}
}

// Errors returns a copy in Field order.
func (v *FormView) Errors() []string {
	out := make([]string, len(v.fields))
	for i, f := range v.fields {
		out[i] = f.err
		if c, ok := f.control.(interface{ Error() string }); ok {
			out[i] = c.Error()
		}
	}
	return out
}

// Required returns msg when s is empty or blank, else "".
func Required(s, msg string) string {
	if strings.TrimSpace(s) == "" {
		return msg
	}
	return ""
}

func (v *FormView) Render(cx *el.Context) el.Element {
	id := autoID("form", v)
	if v.busy && !cx.Enabled(id) {
		v.CancelSubmit()
	}
	if v.focusErrors {
		if v.focusFirstError(cx) {
			v.focusErrors = false
		} else {
			cx.AfterEnabled(id, v, 0, func() {})
		}
	}
	form := el.Div().ID(id).Role("form").Gap(14).Items(el.Stretch).Disabled(v.disabled)
	for i, f := range v.fields {
		if n, ok := f.control.(interface{ setName(string) }); ok {
			n.setName(f.label)
		}
		control := el.Div().ID(v.fieldID(i)).Grow().Items(el.Stretch).Gap(theme.SpaceSm).Child(f.control.Render(cx))
		if _, ok := f.control.(interface{ SetError(string) }); !ok && f.err != "" {
			control.Child(el.Text(f.err).TextColor(theme.Danger).TextSize(theme.TextSm))
		}
		if _, ok := f.control.(interface{ FocusID() string }); !ok {
			control.Focusable(f.err != "")
		}
		form.Child(el.Div().Row().Gap(theme.SpaceLg).Items(el.Start).Disabled(v.busy).Child(
			el.Div().W(el.Dp(v.labelWidth)).NoShrink().Items(el.End).Pt(theme.SpaceMd).Child(el.Text(f.label).TextColor(theme.Muted)), control))
	}
	if len(v.actions) > 0 {
		row := el.Div().Grow().W(el.Dp(0)).Wrap().Gap(theme.SpaceLg).Items(el.Center)
		for _, a := range v.actions {
			row.Child(a.Render(cx))
		}
		form.Child(el.Div().Row().Gap(theme.SpaceLg).Child(el.Div().W(el.Dp(v.labelWidth)).NoShrink(), row))
	}
	return form
}
