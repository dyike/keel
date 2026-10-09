package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"strings"
)

// ComboboxTriggerContext is a per-render snapshot. Selection is caller-owned.
// Toggle and Clear are for UI event callbacks, not for use during rendering.
type ComboboxTriggerContext struct {
	Selection      []ComboboxItem
	Open, Disabled bool
	Size           float32
	Placeholder    string
	Toggle, Clear  func()
}

// RenderTrigger replaces the complete default field presentation. The component
// retains an accessible keyboard target and a background click surface. Custom
// buttons may use the context's Toggle/Clear actions without double activation.
// Searchable mode places its editor inside the popup. Nil restores the default.
func (v *ComboboxView) RenderTrigger(fn func(ComboboxTriggerContext) el.View) *ComboboxView {
	if (v.triggerRenderer == nil) != (fn == nil) {
		v.cancelDraft()
	}
	v.triggerRenderer = fn
	return v
}

func (v *ComboboxView) selectionFocusID() string {
	if v.customTrigger && !v.nonsearchable && v.open {
		return autoID("combobox", v) + "/search"
	}
	return v.FocusID()
}
func (v *ComboboxView) renderCustomTrigger(cx *el.Context, id string) *el.DivEl {
	if v.triggerRenderer == nil {
		return nil
	}
	toggle := func() {
		if v.disabled || !cx.Enabled(id) {
			return
		}
		if v.open {
			v.confirmClose()
			cx.Focus(v.FocusID())
			return
		}
		v.open = true
		if !v.nonsearchable {
			v.text = ""
		}
		v.searchChanged()
		cx.Focus(v.FocusID())
		v.searchFocus = !v.nonsearchable
	}
	clear := func() {
		if !v.disabled && cx.Enabled(id) {
			cx.Focus(v.FocusID())
			v.clearSelection()
		}
	}
	selection := make([]ComboboxItem, 0, len(v.Values()))
	for _, value := range v.Values() {
		selection = append(selection, ComboboxItem{Value: value, Label: v.optionLabel(value), Disabled: v.optionDisabled(value)})
	}
	content := v.triggerRenderer(ComboboxTriggerContext{Selection: selection, Open: v.open, Disabled: v.disabled, Size: float32(theme.ControlHeight) * v.sizeRatio(), Placeholder: v.placeholder, Toggle: toggle, Clear: clear})
	if content == nil {
		return nil
	}
	surface := el.Div().ID(v.FocusID()).Role("button").Name(v.a11y()).Value(strings.Join(v.Values(), ", ")).Focusable(true).
		OnKey(func(e el.KeyEvent) bool {
			name := key.Name(e.Name)
			if name == key.NameReturn || name == key.NameSpace || name == key.NameDownArrow || name == key.NameUpArrow {
				if e.Modifiers != 0 {
					return false
				}
				if e.State == el.KeyPress {
					if !v.open {
						toggle()
					} else if name == key.NameReturn || name == key.NameSpace {
						v.submit()
						cx.Focus(v.selectionFocusID())
					} else {
						return v.optionKey(cx, e)
					}
				}
				return true
			}
			return false
		}).Child(el.Div().ID("activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).CursorPointer().OnClick(toggle), el.Div().ID("content").Child(content.Render(cx)))
	return el.Div().ID(id).Role("combobox").Name(v.a11y()).Value(strings.Join(v.Values(), ", ")).Disabled(v.disabled).Child(surface)
}
func (v *ComboboxView) popupSearch(cx *el.Context, id string) el.Element {
	name := locale.Current().Name(locale.Current().Search, v.a11y())
	if v.searchFocus {
		cx.AfterEnabled(id+"/search", v, 0, func() {
			if v.open && v.customTrigger && !v.nonsearchable {
				cx.Focus(id + "/search")
			}
			v.searchFocus = false
		})
	}
	return el.Div().ID(id + "/search-frame").Px(theme.SpaceSm).Py(theme.SpaceXs * v.sizeRatio()).Child(
		el.Input().ID(id + "/search").Name(name).Placeholder(locale.Current().Search).Bind(&v.text).H(el.Dp(float32(theme.ControlHeight) * v.sizeRatio())).TextSize(float32(theme.BodySize) * v.sizeRatio()).OnChange(func(string) { v.searchChanged(); cx.ScrollTo(v.virtual.ID(), 0) }).OnKey(func(e el.KeyEvent) bool { return v.optionKey(cx, e) }).OnSubmit(func(string) { v.submit(); cx.Focus(v.selectionFocusID()) }),
	)
}
