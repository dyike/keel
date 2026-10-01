package kit

import (
	"slices"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type accordionItem struct {
	title    string
	body     el.View
	disabled bool
}

// AccordionView stacks sections that expand and collapse. One section is open
// at a time unless Multiple. Headers take focus: ↑ ↓ Home End move between
// them, Enter or Space toggles.
type AccordionView struct {
	items    []accordionItem
	open     []int
	multiple bool
	onChange func(open []int)
}

func Accordion() *AccordionView { return &AccordionView{} }
func (v *AccordionView) Add(title string, body el.View) *AccordionView {
	v.items = append(v.items, accordionItem{title: title, body: body})
	return v
}
func (v *AccordionView) Multiple() *AccordionView                    { v.multiple = true; return v }
func (v *AccordionView) OnChange(fn func(open []int)) *AccordionView { v.onChange = fn; return v }

// Value returns the open sections' indexes in order (a copy).
func (v *AccordionView) Value() []int { return slices.Clone(v.open) }

// SetValue opens exactly these sections without calling OnChange; a single
// accordion keeps only the first.
func (v *AccordionView) SetValue(open ...int) {
	v.open = nil
	for i := range v.items {
		if slices.Contains(open, i) && (v.multiple || len(v.open) == 0) {
			v.open = append(v.open, i)
		}
	}
}
func (v *AccordionView) SetItemDisabled(i int, on bool) {
	if i >= 0 && i < len(v.items) {
		v.items[i].disabled = on
	}
}

func (v *AccordionView) toggle(i int) {
	switch on := slices.Contains(v.open, i); {
	case on:
		v.open = slices.DeleteFunc(v.open, func(j int) bool { return j == i })
	case v.multiple:
		v.SetValue(append(v.open, i)...)
	default:
		v.open = []int{i}
	}
	if v.onChange != nil {
		v.onChange(v.Value())
	}
}

func (v *AccordionView) Render(cx *el.Context) el.Element {
	id := autoID("accordion", v)
	out := el.Div().Role("group").Items(el.Stretch).Rounded(8).Border(1, theme.Border).Bg(theme.Surface)
	enabled := func(from, d int) int {
		for k := 1; k <= len(v.items); k++ {
			j := ((from+d*k)%len(v.items) + len(v.items)) % len(v.items)
			if !v.items[j].disabled {
				return j
			}
		}
		return from
	}
	for i, it := range v.items {
		i, it := i, it
		on := slices.Contains(v.open, i)
		state, icon := "collapsed", IconChevronRight
		if on {
			state, icon = "expanded", IconChevronDown
		}
		if i > 0 {
			out.Child(el.Div().H(el.Dp(1)).Bg(theme.Border))
		}
		head := el.Div().ID(id+"/"+strconv.Itoa(i)).Role("disclosure").Name(it.title).Value(state).Disabled(it.disabled).
			Row().Items(el.Center).Gap(8).Px(14).Py(12).Focusable(true).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			OnClick(func() { v.toggle(i) }).
			OnKey(func(e el.KeyEvent) bool {
				j, ok := i, true
				switch key.Name(e.Name) {
				case key.NameDownArrow:
					j = enabled(i, 1)
				case key.NameUpArrow:
					j = enabled(i, -1)
				case key.NameHome:
					j = enabled(-1, 1)
				case key.NameEnd:
					j = enabled(len(v.items), -1)
				default:
					ok = false
				}
				if ok && e.State == el.KeyPress {
					cx.Focus(id + "/" + strconv.Itoa(j))
				}
				return ok
			}).
			Child(el.Text(it.title).Bold().Grow(), Icon(icon).Size(16).Color(theme.Muted).Render(cx))
		if !it.disabled {
			head.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
		out.Child(head)
		if on && it.body != nil {
			out.Child(el.Div().Px(14).Pb(14).Items(el.Stretch).Child(it.body.Render(cx)))
		}
	}
	return out
}
