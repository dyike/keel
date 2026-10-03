package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// InputGroupAlignment places an addon around the editor row.
type InputGroupAlignment uint8

const (
	InputGroupInlineStart InputGroupAlignment = iota
	InputGroupInlineEnd
	InputGroupBlockStart
	InputGroupBlockEnd
)

type inputGroupAddon struct {
	id        string
	alignment InputGroupAlignment
	view      el.View
}

// InputGroupView composes a text input and independent inline/block views
// inside one field border. The input remains the owner of its value and events.
type InputGroupView struct {
	label, name    string
	input          *InputView
	prefix, suffix el.View
	disabled       bool
	addons         []inputGroupAddon
}

// InputGroup wraps input; nil creates an empty single-line input.
// Render the input through this group only, not separately in the same frame.
func InputGroup(label string, input *InputView) *InputGroupView {
	if input == nil {
		input = Input("")
	}
	return &InputGroupView{label: label, input: input}
}
func (v *InputGroupView) Prefix(view el.View) *InputGroupView { v.prefix = view; return v }
func (v *InputGroupView) Suffix(view el.View) *InputGroupView { v.suffix = view; return v }

// Addon inserts content at one of four positions. IDs must be nonempty and
// stable. Reusing an ID replaces its content/position; nil removes it.
// Each side preserves insertion order. Buttons retain their own configuration.
func (v *InputGroupView) Addon(id string, alignment InputGroupAlignment, view el.View) *InputGroupView {
	if id == "" {
		panic("kit.InputGroup: empty addon ID")
	}
	if alignment > InputGroupBlockEnd {
		return v
	}
	index := slices.IndexFunc(v.addons, func(a inputGroupAddon) bool { return a.id == id })
	if view == nil {
		if index >= 0 {
			v.addons = slices.Delete(v.addons, index, index+1)
		}
	} else if index >= 0 {
		v.addons[index] = inputGroupAddon{id, alignment, view}
	} else {
		v.addons = append(v.addons, inputGroupAddon{id, alignment, view})
	}
	return v
}
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
	field := fieldFrame(id, cx.FocusWithin(id), v.Error() != "", disabled, v.input.readOnly).Col().Items(el.Stretch).Justify(el.Center).FocusOnPress(v.input.FocusID()).Role("group").Name(name)
	appendAddons := func(parent *el.DivEl, alignment InputGroupAlignment) {
		for _, addon := range v.addons {
			if addon.alignment != alignment {
				continue
			}
			box := el.Div().ID(id + "/addon/" + addon.id).TextColor(theme.Muted).TextSize(theme.TextSm)
			if alignment == InputGroupBlockStart || alignment == InputGroupBlockEnd {
				box.WFull()
			} else {
				box.NoShrink()
			}
			parent.Child(box.Child(addon.view.Render(cx)))
		}
	}
	appendAddons(field, InputGroupBlockStart)
	row := el.Div().ID(id + "/row").WFull().Row().Items(el.Center).Gap(theme.SpaceMd)
	if v.input.multiline {
		row.Items(el.Start)
	}
	if v.prefix != nil {
		row.Child(el.Div().ID(id + "/prefix").NoShrink().Child(v.prefix.Render(cx)))
	}
	appendAddons(row, InputGroupInlineStart)
	row.Child(el.Div().ID(id + "/field").Grow().MinW(el.Dp(24)).Child(v.input.render(cx, false)))
	appendAddons(row, InputGroupInlineEnd)
	if v.suffix != nil {
		row.Child(el.Div().ID(id + "/suffix").NoShrink().Child(v.suffix.Render(cx)))
	}
	field.Child(row)
	appendAddons(field, InputGroupBlockEnd)

	root := el.Div().WFull().Gap(theme.SpaceSm).Disabled(disabled)
	label := v.label
	if label == "" {
		label = v.input.label
	}
	if label != "" {
		root.Child(el.Div().Name(label).OnClick(func() { cx.Focus(v.FocusID()) }).Child(el.Text(label).TextSize(theme.TextMd).TextColor(theme.Muted)))
	}
	root.Child(field)
	if v.Error() != "" {
		root.Child(el.Text(v.Error()).TextSize(theme.TextSm).TextColor(theme.Danger))
	}
	return root
}
