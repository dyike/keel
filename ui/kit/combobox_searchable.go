package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Searchable controls query editing (enabled by default). Changing the mode
// closes the popup, invalidates old requests and discards uncommitted text.
// In non-searchable mode async lookup receives an empty query on open/retry;
// AllowCustom is inactive, and existing selections remain unchanged.
func (v *ComboboxView) Searchable(on bool) *ComboboxView {
	if v.nonsearchable == !on {
		return v
	}
	v.cancelDraft()
	v.nonsearchable = !on
	v.cached = false
	v.active = -1
	return v
}

func (v *ComboboxView) staticTrigger(cx *el.Context) el.Element {
	text := v.optionLabel(v.value)
	if v.multiple {
		text = ""
	}
	fg := theme.Text
	if text == "" {
		text = v.placeholder
		fg = theme.Muted
	}
	trigger := el.Div().ID(v.FocusID()).Role("button").Name(v.a11y()).Value(v.value).Grow().MinW(el.Dp(0)).Focusable(true).CursorPointer().
		OnClick(func() {
			if v.open {
				v.close()
			} else {
				v.open = true
				v.searchChanged()
			}
			cx.Focus(v.FocusID())
		}).
		OnKey(func(e el.KeyEvent) bool {
			if v.open && (key.Name(e.Name) == key.NameReturn || key.Name(e.Name) == key.NameSpace) && e.Modifiers == 0 {
				if e.State == el.KeyPress {
					matches := v.matches()
					if !v.loading && v.searchError == "" && v.active >= 0 && v.active < len(matches) {
						v.choose(matches[v.active])
					}
				}
				return true
			}
			return v.optionKey(cx, e)
		}).Child(el.Text(text).TextColor(fg).MaxLines(1))
	if v.multiple {
		trigger.MinW(el.Dp(100))
	}
	return trigger
}
