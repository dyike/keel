package kit

import (
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type tabPage struct {
	title string
	page  el.View
}

// TabsView switches between pages. Only the current page renders; each page
// is a view the app keeps, so its state survives switching. Tab reaches the
// current tab; ← → move between tabs, Home / End jump.
type TabsView struct {
	pages    []tabPage
	current  int
	onChange func(int)
}

func Tabs() *TabsView { return &TabsView{} }
func (v *TabsView) Add(title string, page el.View) *TabsView {
	v.pages = append(v.pages, tabPage{title, page})
	return v
}
func (v *TabsView) OnChange(fn func(index int)) *TabsView { v.onChange = fn; return v }
func (v *TabsView) Value() int                            { return v.current }

// SetValue shows tab i without calling OnChange.
func (v *TabsView) SetValue(i int) { v.current = min(max(i, 0), max(len(v.pages)-1, 0)) }

func (v *TabsView) choose(cx *el.Context, i int) {
	cx.Focus(autoID("tabs", v) + "/" + strconv.Itoa(i))
	if i == v.current {
		return
	}
	v.current = i
	if v.onChange != nil {
		v.onChange(i)
	}
}

func (v *TabsView) Render(cx *el.Context) el.Element {
	id := autoID("tabs", v)
	bar := el.Div().Role("tablist").Row()
	for i, p := range v.pages {
		i := i
		on := i == v.current
		tab := el.Div().ID(id + "/" + strconv.Itoa(i)).Role("tab").Name(p.title).Selected(on).
			Px(14).Py(8).CursorPointer().TextColor(theme.Muted).Focusable(on).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			Hover(func(s *el.Style) { s.TextColor(theme.Text) }).
			OnClick(func() { v.choose(cx, i) }).
			OnKey(func(e el.KeyEvent) bool {
				j, ok := i, true
				switch key.Name(e.Name) {
				case key.NameRightArrow:
					j = (i + 1) % len(v.pages)
				case key.NameLeftArrow:
					j = (i - 1 + len(v.pages)) % len(v.pages)
				case key.NameHome:
					j = 0
				case key.NameEnd:
					j = len(v.pages) - 1
				default:
					ok = false
				}
				if ok && e.State == el.KeyPress {
					v.choose(cx, j)
				}
				return ok
			}).
			Child(el.Text(p.title))
		underline := el.Div().H(el.Dp(2))
		if on {
			tab.TextColor(theme.PrimaryText)
			underline.Bg(theme.Primary)
		}
		bar.Child(el.Div().Items(el.Stretch).Child(tab, underline))
	}
	out := el.Div().Gap(16).Items(el.Stretch).Child(
		el.Div().Items(el.Stretch).Child(bar, el.Div().H(el.Dp(1)).Bg(theme.Border)),
	)
	if v.current < len(v.pages) && v.pages[v.current].page != nil {
		out.Child(el.Div().Role("tabpanel").Name(v.pages[v.current].title).Items(el.Stretch).Child(v.pages[v.current].page.Render(cx)))
	}
	return out
}
