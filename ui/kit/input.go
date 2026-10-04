package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// InputView is a labelled text field with optional prefix, suffix, clear
// button and error message. TextArea makes a multi-line one.
type InputView struct {
	size                           InputSize
	name                           string // accessible name from a Form row when label is empty
	label, placeholder, value, err string
	multiline, password, clearable bool
	readOnly, disabled             bool
	maxLen, rows                   int
	minRows, maxRows               int
	filter                         string
	prefix, suffix                 el.View
	onChange, onSubmit             func(string)
}

func Input(label string) *InputView { return &InputView{label: label} }

// TextArea is a multi-line field at least rows lines tall (3 by default);
// Enter inserts a newline.
func TextArea(label string) *InputView { return &InputView{label: label, multiline: true, rows: 3} }

func (v *InputView) Placeholder(s string) *InputView     { v.placeholder = s; return v }
func (v *InputView) Password() *InputView                { v.password = true; return v }
func (v *InputView) Clearable() *InputView               { v.clearable = true; return v }
func (v *InputView) MaxLength(n int) *InputView          { v.maxLen = n; return v }
func (v *InputView) Prefix(p el.View) *InputView         { v.prefix = p; return v }
func (v *InputView) Suffix(s el.View) *InputView         { v.suffix = s; return v }
func (v *InputView) OnChange(fn func(string)) *InputView { v.onChange = fn; return v }
func (v *InputView) OnSubmit(fn func(string)) *InputView { v.onSubmit = fn; return v }
func (v *InputView) Value() string                       { return v.value }
func (v *InputView) SetValue(s string)                   { v.value = s }
func (v *InputView) SetDisabled(on bool)                 { v.disabled = on }
func (v *InputView) SetReadOnly(on bool)                 { v.readOnly = on }
func (v *InputView) SetLabel(s string)                   { v.label = s }

// Rows sets the minimum height of a TextArea in lines.
func (v *InputView) Rows(n int) *InputView {
	if n > 0 {
		v.rows = n
		v.minRows, v.maxRows = 0, 0
	}
	return v
}

// AutoGrow sizes a TextArea between minRows and maxRows wrapped lines.
// Additional content scrolls inside the field. Invalid ranges are ignored;
// Rows restores the original minimum-height mode. Input ignores this option.
func (v *InputView) AutoGrow(minRows, maxRows int) *InputView {
	if minRows > 0 && maxRows >= minRows {
		v.minRows, v.maxRows = minRows, maxRows
	}
	return v
}

// Filter accepts only these runes, typed or pasted; "" accepts everything.
func (v *InputView) Filter(chars string) *InputView { v.filter = chars; return v }

// SetError shows msg under the field and marks it invalid; "" clears it.
// Form sets it from validators.
func (v *InputView) SetError(msg string) { v.err = msg }
func (v *InputView) Error() string       { return v.err }

// FocusID is the element ID that cx.Focus uses to focus the text box.
func (v *InputView) FocusID() string { return autoID("input", v) + "/text" }

func (v *InputView) Render(cx *el.Context) el.Element { return v.render(cx, true) }

func (v *InputView) render(cx *el.Context, chrome bool) el.Element {
	id := autoID("input", v)
	name := v.a11y()
	if !chrome && v.name != "" {
		name = v.name
	}
	if name == "" {
		name = v.placeholder
	}
	text := fieldText(el.Input().ID(v.FocusID()).Name(name).Placeholder(v.placeholder).Bind(&v.value)).
		MaxLen(v.maxLen).Filter(v.filter).ReadOnly(v.readOnly).
		OnChange(func(s string) {
			v.err = ""
			if v.onChange != nil {
				v.onChange(s)
			}
		}).
		OnSubmit(func(s string) {
			if v.onSubmit != nil {
				v.onSubmit(s)
			}
		})
	if v.multiline {
		text = fieldText(el.TextArea().ID(v.FocusID()).Name(name).Placeholder(v.placeholder).Bind(&v.value)).MinH(el.Dp(float32(v.rows) * 22)).
			MaxLen(v.maxLen).Filter(v.filter).ReadOnly(v.readOnly).
			OnChange(func(s string) {
				v.err = ""
				if v.onChange != nil {
					v.onChange(s)
				}
			})
	}
	if v.multiline && v.minRows > 0 {
		text.MinH(el.Auto).AutoGrow(v.minRows, v.maxRows)
	}
	if v.password {
		text.Password()
	}
	box := fieldFrame(id, cx.FocusWithin(id), v.err != "", v.disabled, v.readOnly).FocusOnPress(v.FocusID())
	if v.multiline {
		box.Items(el.Start).Py(theme.SpaceMd)
	}
	if v.prefix != nil {
		box.Child(el.Div().TextColor(theme.Muted).Child(v.prefix.Render(cx)))
	}
	box.Child(text)
	if v.clearable && v.value != "" && !v.disabled && !v.readOnly {
		box.Child(el.Div().Name(locale.Current().Name(locale.Current().Clear, name)).P(theme.SpaceXxs).Rounded(theme.RadiusSm).
			CursorPointer().Focusable(false).Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
			OnClick(func() {
				v.value = ""
				cx.Focus(v.FocusID())
				if v.onChange != nil {
					v.onChange("")
				}
			}).Child(Icon(IconClose).Size(14).Color(theme.Muted).Render(cx)))
	}
	if v.suffix != nil {
		box.Child(el.Div().TextColor(theme.Muted).Child(v.suffix.Render(cx)))
	}
	v.applySize(text, box)
	if !chrome {
		return box.Border(0, theme.Border).Rounded(0).P(0).MinH(el.Auto) // inside an InputGroup's frame
	}
	return labelled(v.label, box, v.err)
}

func (v *InputView) setName(s string) { v.name = s }
func (v *InputView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
