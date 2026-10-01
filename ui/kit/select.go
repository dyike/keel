package kit

import (
	"strconv"
	"strings"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SelectView picks one option from a list that opens under the field.
// Enter, Space or ↓ opens it; in the list ↑ ↓ move, Enter chooses, Esc closes
// and focus returns to the field. Searchable adds a filter box on top; ↓
// moves from it into the list.
type SelectView struct {
	name                    string // accessible name from a Form row when label is empty
	label, hint, value, err string
	options                 []string
	open, searchable        bool
	disabled                bool
	query                   string
	onChange                func(string)
}

func Select(label string, options ...string) *SelectView {
	return &SelectView{label: label, options: options}
}

// Hint is shown while nothing is chosen; empty uses the locale's SelectHint.
func (v *SelectView) Hint(s string) *SelectView                  { v.hint = s; return v }
func (v *SelectView) Searchable() *SelectView                    { v.searchable = true; return v }
func (v *SelectView) OnChange(fn func(value string)) *SelectView { v.onChange = fn; return v }
func (v *SelectView) Value() string                              { return v.value }

// SetValue chooses an option ("" clears) without calling OnChange.
func (v *SelectView) SetValue(s string) { v.value = s }
func (v *SelectView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.open = false
	}
}
func (v *SelectView) SetError(msg string) { v.err = msg }
func (v *SelectView) Error() string       { return v.err }
func (v *SelectView) FocusID() string     { return autoID("select", v) }

// SetOptions replaces the options, keeping the choice if it is still offered.
func (v *SelectView) SetOptions(options ...string) {
	v.options = options
	if !v.offered(v.value) {
		v.value = ""
	}
}

func (v *SelectView) offered(s string) bool {
	for _, o := range v.options {
		if o == s {
			return true
		}
	}
	return false
}

func (v *SelectView) visible() []string {
	if !v.searchable || v.query == "" {
		return v.options
	}
	var out []string
	q := strings.ToLower(v.query)
	for _, o := range v.options {
		if strings.Contains(strings.ToLower(o), q) {
			out = append(out, o)
		}
	}
	return out
}

func (v *SelectView) setOpen(cx *el.Context, open bool) {
	v.open, v.query = open, ""
	if open {
		// Start on the chosen option, or the first one.
		i := 0
		for j, o := range v.options {
			if o == v.value {
				i = j
			}
		}
		if v.searchable {
			cx.Focus(v.FocusID() + "/search")
		} else {
			cx.Focus(v.FocusID() + "/" + strconv.Itoa(i))
		}
	}
}

func (v *SelectView) choose(s string) {
	v.open = false
	if v.value == s {
		return
	}
	v.value, v.err = s, ""
	if v.onChange != nil {
		v.onChange(s)
	}
}

func (v *SelectView) Render(cx *el.Context) el.Element {
	id := v.FocusID()
	name := v.a11y()
	if name == "" { // name an unlabelled select after the hint it shows
		name = v.hint
		if name == "" {
			name = locale.Current().SelectHint
		}
	}
	shown, color := v.value, theme.Text
	if shown == "" {
		shown, color = v.hint, theme.Muted
		if shown == "" {
			shown = locale.Current().SelectHint
		}
	}
	border := theme.Border
	if v.err != "" {
		border = theme.Danger
	}
	field := el.Div().ID(id).WFull().Role("select").Name(name).Value(v.value).Disabled(v.disabled).
		Row().Items(el.Center).Gap(8).H(el.Dp(36)).Px(10).Rounded(6).Bg(theme.Surface).Border(1, border).
		Focusable(true).OnClick(func() { v.setOpen(cx, !v.open) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) != key.NameDownArrow {
				return false
			}
			if e.State == el.KeyPress && !v.open {
				v.setOpen(cx, true)
			}
			return true
		}).
		Child(el.Text(shown).TextColor(color).Grow().MaxLines(1), Icon(IconChevronDown).Size(16).Color(theme.Muted).Render(cx))
	if v.disabled {
		field.Bg(theme.Subtle).TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	if v.open {
		cx.Overlay(id, el.Anchored(id, v.list(cx, id)).MatchAnchorWidth().Modal().TrapFocus().
			OnDismiss(func() { v.open, v.query = false, "" }))
	}
	return labelled(v.label, field, v.err)
}

func (v *SelectView) list(cx *el.Context, id string) el.Element {
	list := surface().Role("listbox").Name(v.a11y()).Py(4).Items(el.Stretch).MaxH(el.Dp(280))
	if v.searchable {
		list.Child(el.Div().Px(4).Pb(4).Child(el.Input().ID(id + "/search").Name(locale.Current().Search).
			Placeholder(locale.Current().Search).Bind(&v.query).Py(6).
			OnSubmit(func(string) {
				if opts := v.visible(); len(opts) > 0 {
					v.choose(opts[0])
				}
			}).
			OnKey(func(e el.KeyEvent) bool {
				if e.State == el.KeyPress && key.Name(e.Name) == key.NameDownArrow && len(v.visible()) > 0 {
					cx.Focus(id + "/0") // from the search box into the list
				}
				return true
			})))
	}
	items := el.Div().ScrollY().Items(el.Stretch)
	opts := v.visible()
	if len(opts) == 0 {
		items.Child(el.Div().Px(12).Py(6).Child(el.Text(locale.Current().NoMatches).TextColor(theme.Muted)))
	}
	for i, o := range opts {
		i, o := i, o
		move := func(j int) { cx.Focus(id + "/" + strconv.Itoa((j+len(opts))%len(opts))) }
		items.Child(el.Div().ID(id+"/"+strconv.Itoa(i)).Role("option").Name(o).Selected(o == v.value).
			Mx(4).Px(8).H(el.Dp(30)).Row().Items(el.Center).Rounded(4).CursorPointer().
			Focusable(true).OnClick(func() { v.choose(o) }).
			FocusStyle(func(s *el.Style) { s.Bg(theme.Subtle).BorderColor(theme.Subtle) }).
			Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
			OnKey(func(e el.KeyEvent) bool {
				switch key.Name(e.Name) {
				case key.NameDownArrow:
					if e.State == el.KeyPress {
						move(i + 1)
					}
				case key.NameUpArrow:
					if e.State == el.KeyPress {
						move(i - 1)
					}
				case key.NameHome:
					if e.State == el.KeyPress {
						move(0)
					}
				case key.NameEnd:
					if e.State == el.KeyPress {
						move(-1)
					}
				default:
					return false
				}
				return true
			}).
			Child(el.Text(o).Grow().MaxLines(1), checkMark(cx, o == v.value)))
	}
	return list.Child(items)
}

func checkMark(cx *el.Context, on bool) el.Element {
	if !on {
		return el.Div().Size(el.Dp(14))
	}
	return Icon(IconDone).Size(14).Color(theme.PrimaryText).Render(cx)
}

func (v *SelectView) setName(s string) { v.name = s }
func (v *SelectView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
