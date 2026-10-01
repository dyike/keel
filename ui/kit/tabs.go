package kit

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"
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
// current tab; ← → move between tabs, Home / End jump. Tabs that do not fit
// move into a 更多 menu; with Closable each tab has a close button.
type TabsView struct {
	pages    []tabPage
	current  int
	onChange func(int)
	onClose  func(int)
	widths   []float32
	avail    float32
	more     *MenuView
}

func Tabs() *TabsView { return &TabsView{more: Menu()} }
func (v *TabsView) Add(title string, page el.View) *TabsView {
	v.pages = append(v.pages, tabPage{title, page})
	v.widths = append(v.widths, 0)
	return v
}

// Closable shows a close button on each tab; fn decides what closing means,
// usually Remove(i).
func (v *TabsView) Closable(fn func(index int)) *TabsView { v.onClose = fn; return v }

// Remove deletes tab i, keeping the current tab when it remains.
func (v *TabsView) Remove(i int) {
	if i < 0 || i >= len(v.pages) {
		return
	}
	v.pages = append(v.pages[:i], v.pages[i+1:]...)
	v.widths = append(v.widths[:i], v.widths[i+1:]...)
	if v.current > i || v.current >= len(v.pages) {
		v.current = max(v.current-1, 0)
	}
}

// Len is the number of tabs.
func (v *TabsView) Len() int { return len(v.pages) }

func (v *TabsView) fit() int {
	if v.avail <= 0 {
		return len(v.pages)
	}
	const moreW = 40
	used := float32(0)
	for i, w := range v.widths {
		used += w
		if used > v.avail {
			for i > 0 && used-v.widths[i]+moreW > v.avail {
				i--
				used -= v.widths[i]
			}
			return i
		}
	}
	return len(v.pages)
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
	n := v.fit()
	for i, p := range v.pages[:n] {
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
			Row().Items(el.Center).Gap(6).Child(el.Text(p.title).MaxLines(1))
		head := el.Div().Row().Items(el.Center).Child(tab)
		if v.onClose != nil {
			// Beside the tab, not inside it: a click on the button must not
			// also select the tab it is closing.
			head.Pr(6).Child(el.Div().Name(locale.Current().Name(locale.Current().Close, p.title)).P(2).Rounded(4).
				Focusable(false).CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				OnClick(func() { v.onClose(i) }).Child(Icon(IconClose).Size(12).Color(theme.Muted).Render(cx)))
		}
		underline := el.Div().H(el.Dp(2))
		if on {
			tab.TextColor(theme.PrimaryText)
			underline.Bg(theme.Primary)
		}
		bar.Child(el.Div().NoShrink().Items(el.Stretch).Child(head, underline).Decorate(func(gtx core.C, draw func()) {
			if px := gtx.Metric.PxPerDp; px > 0 && i < len(v.widths) {
				v.widths[i] = float32(gtx.Constraints.Max.X) / px
			}
			draw()
		}))
	}
	if n < len(v.pages) {
		v.more.items = v.more.items[:0]
		for i := n; i < len(v.pages); i++ {
			i := i
			v.more.Item(v.pages[i].title, "", func() { v.choose(cx, i) })
		}
		label := locale.Current().More
		if v.current >= n {
			label = v.pages[v.current].title // the hidden current tab names the menu
		}
		v.more.Trigger(Button(label, v.more.Toggle).Icon(IconChevronDown).Variant(ButtonGhost).Size(32))
		bar.Child(v.more.Render(cx))
	}
	out := el.Div().Gap(16).Items(el.Stretch).Child(
		// The row around the bar measures the width the parent allows; a parent
		// sized by its content allows the bar's own width, so nothing overflows.
		el.Div().Items(el.Stretch).Child(el.Div().Row().Decorate(func(gtx core.C, draw func()) {
			if px := gtx.Metric.PxPerDp; px > 0 {
				v.avail = float32(gtx.Constraints.Max.X) / px
			}
			draw()
		}).Child(bar), el.Div().H(el.Dp(1)).Bg(theme.Border)),
	)
	if v.current < len(v.pages) && v.pages[v.current].page != nil {
		out.Child(el.Div().Role("tabpanel").Name(v.pages[v.current].title).Items(el.Stretch).Child(v.pages[v.current].page.Render(cx)))
	}
	return out
}
