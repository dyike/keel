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
	options  FormFieldOptions
}

// FormFieldOptions configures a field without changing its control or validator.
// ColStart is one-based; zero uses the next free column. Required is visual only.
type FormFieldOptions struct {
	Description        string
	DescriptionContent el.View
	Required, Hidden   bool
	ColSpan, ColStart  int
}

// FormView lays out labelled fields in a configurable grid and validates them on
// Submit. A validator returns an error message, or "" when the value is fine;
// the message appears under the field until the next Submit or the user
// changes the field.
//
//	f := kit.Form().
//	    Field("客户", customer, func() string { return kit.Required(customer.Value(), "请填写客户") }).
//	    Field("金额", amount, nil)
//	kit.Button("创建", func() { if f.Validate(cx) { create() } })
type FormView struct {
	fields             []formField
	actions            []el.View
	labelWidth         float32
	disabled, busy     bool
	request            uint64
	focusErrors        bool
	columns            int
	verticalLabels     bool
	gap, labelTextSize float32
	footer             el.View
}

func Form() *FormView { return &FormView{labelWidth: 72, columns: 1, gap: 14} }

// Field adds a row. validate may be nil; then control needs no Validatable.
func (v *FormView) Field(label string, control el.View, validate func() string) *FormView {
	v.CancelSubmit()
	v.fields = append(v.fields, formField{label: label, control: control, validate: validate})
	return v
}

// FieldWithOptions adds a configured field; options are copied.
func (v *FormView) FieldWithOptions(label string, control el.View, validate func() string, options FormFieldOptions) *FormView {
	v.Field(label, control, validate)
	v.fields[len(v.fields)-1].options = options
	return v
}

// SetFieldOptions replaces presentation options by insertion index. A visibility
// change cancels pending submission and clears the field's previous error.
func (v *FormView) SetFieldOptions(index int, options FormFieldOptions) bool {
	if index < 0 || index >= len(v.fields) {
		return false
	}
	if v.fields[index].options.Hidden != options.Hidden {
		v.CancelSubmit()
		v.setError(index, "")
	}
	v.fields[index].options = options
	return true
}

// Columns sets independent field columns. Responsive column choices belong to the caller.
func (v *FormView) Columns(n int) *FormView { v.columns = max(1, n); return v }

// VerticalLabels puts labels above controls; false restores side labels.
func (v *FormView) VerticalLabels(on bool) *FormView { v.verticalLabels = on; return v }
func (v *FormView) Gap(dp float32) *FormView {
	if dp >= 0 && dp < 1024 {
		v.gap = dp
	}
	return v
}
func (v *FormView) LabelTextSize(sp float32) *FormView {
	if sp >= 0 && sp < 1024 {
		v.labelTextSize = sp
	}
	return v
}

// Footer adds a full-width trailing-aligned view after Actions; nil removes it.
func (v *FormView) Footer(view el.View) *FormView { v.footer = view; return v }

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

// Validate runs visible fields' validators, shows their messages, and focuses the first
// invalid field. It reports whether all passed. Call it from a callback.
func (v *FormView) Validate(cx *el.Context) bool {
	v.CancelSubmit()
	// Commit editor drafts before observing values in validators.
	for _, f := range v.fields {
		if f.options.Hidden {
			continue
		}
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
		if !v.fields[i].options.Hidden && v.fields[i].validate != nil {
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
		if f.options.Hidden || f.err == "" {
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
		if !v.fields[i].options.Hidden && i < len(errors) {
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
	form := el.Div().ID(id).Role("form").Gap(v.gap).Items(el.Stretch).Disabled(v.disabled)
	grid := el.Div().Grid(v.columns).Gap(v.gap).Items(el.Stretch)
	cursor := 0
	for i, f := range v.fields {
		if f.options.Hidden {
			continue
		}
		span := min(max(1, f.options.ColSpan), v.columns)
		start := 0
		if f.options.ColStart > 0 {
			start = min(f.options.ColStart, v.columns) - 1
			span = min(span, v.columns-start)
		}
		if f.options.ColStart > 0 {
			if start < cursor {
				grid.Child(el.Div().ColSpan(v.columns - cursor))
				cursor = 0
			}
			if start > cursor {
				grid.Child(el.Div().ColSpan(start - cursor))
				cursor = start
			}
		} else if cursor+span > v.columns {
			cursor = 0
		}
		grid.Child(v.renderField(cx, i, f).ColSpan(span))
		cursor = (cursor + span) % v.columns
	}
	form.Child(grid)
	if len(v.actions) > 0 {
		row := el.Div().Grow().W(el.Dp(0)).Wrap().Gap(theme.SpaceLg).Items(el.Center)
		for _, a := range v.actions {
			row.Child(a.Render(cx))
		}
		if v.verticalLabels {
			form.Child(row.WFull())
		} else {
			form.Child(el.Div().Row().Gap(theme.SpaceLg).Child(el.Div().W(el.Dp(v.labelWidth)).NoShrink(), row))
		}
	}
	if v.footer != nil {
		form.Child(el.Div().Items(el.End).Child(v.footer.Render(cx)))
	}
	return form
}

func (v *FormView) renderField(cx *el.Context, i int, f formField) *el.DivEl {
	if n, ok := f.control.(interface{ setName(string) }); ok {
		n.setName(f.label)
	}
	control := el.Div().ID(v.fieldID(i)).Grow().W(el.Dp(0)).Items(el.Stretch).Gap(theme.SpaceSm)
	if f.control != nil {
		control.Child(f.control.Render(cx))
	}
	if f.options.DescriptionContent != nil {
		control.Child(f.options.DescriptionContent.Render(cx))
	} else if f.options.Description != "" {
		control.Child(el.Text(f.options.Description).TextSize(theme.TextSm).TextColor(theme.Muted))
	}
	if _, ok := f.control.(interface{ SetError(string) }); !ok && f.err != "" {
		control.Child(el.Text(f.err).TextColor(theme.Danger).TextSize(theme.TextSm))
	}
	if _, ok := f.control.(interface{ FocusID() string }); !ok {
		control.Focusable(f.err != "")
	}
	labelText := el.Text(f.label).TextColor(theme.Muted)
	if v.labelTextSize > 0 {
		labelText.TextSize(v.labelTextSize)
	}
	label := el.Div().Row().Gap(theme.SpaceXxs).Child(labelText)
	if f.options.Required {
		label.Child(el.Text("*").TextColor(theme.Danger))
	}
	field := el.Div().ID(autoID("form", v) + "/cell/" + strconv.Itoa(i)).Items(el.Start).Gap(theme.SpaceLg).Disabled(v.busy)
	if v.verticalLabels {
		field.Items(el.Stretch).Gap(theme.SpaceSm).Child(label, control.WFull())
	} else {
		field.Row().Child(el.Div().W(el.Dp(v.labelWidth)).NoShrink().Items(el.End).Pt(theme.SpaceMd).Child(label), control)
	}
	return field
}
