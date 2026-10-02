package kit

import (
	"slices"
	"strconv"

	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type tabPage struct {
	id    uint64
	title string
	page  el.View
}

// TabsView switches between pages. Only the current page renders; each page
// is a view the app keeps, so its state survives switching. Tab reaches the
// current tab; ← → move between tabs, Home / End jump. Tabs that do not fit
// move into a 更多 menu; with Closable each tab has a close button.
type TabsView struct {
	pages                  []tabPage
	current                int
	onChange               func(int)
	onClose                func(int)
	widths                 []float32
	avail                  float32
	more                   *MenuView
	nextID                 uint64
	disabled, focusPending bool
	height                 float32
	leading, trailing      el.View
	reorder                bool
	onMove                 func(int, int)
	dragID                 uint64
	offsets                map[int]float32
}

func Tabs() *TabsView { return &TabsView{more: Menu(), height: 40} }
func (v *TabsView) Add(title string, page el.View) *TabsView {
	v.nextID++
	v.pages = append(v.pages, tabPage{id: v.nextID, title: title, page: page})
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
	if i == v.current {
		v.focusPending = true
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
	used := float32(0)
	for i := range v.pages {
		used += v.tabWidth(i)
		if used > v.avail {
			for i > 0 && used-v.tabWidth(i)+36 > v.avail {
				i--
				used -= v.tabWidth(i)
			}
			return i
		}
	}
	return len(v.pages)
}
func (v *TabsView) tabWidth(i int) float32 {
	if v.widths[i] > 0 {
		return v.widths[i]
	}
	return 100
}
func (v *TabsView) visible() []int {
	n := v.fit()
	out := make([]int, n)
	for i := range n {
		out[i] = i
	}
	if n < len(v.pages) && !slices.Contains(out, v.current) {
		used := float32(36) + v.tabWidth(v.current)
		for _, i := range out {
			used += v.tabWidth(i)
		}
		for len(out) > 0 && used > v.avail {
			used -= v.tabWidth(out[len(out)-1])
			out = out[:len(out)-1]
		}
		out = append(out, v.current)
	}
	return out
}
func (v *TabsView) tabID(i int) string {
	return autoID("tabs", v) + "/" + strconv.FormatUint(v.pages[i].id, 10)
}
func (v *TabsView) OnChange(fn func(index int)) *TabsView { v.onChange = fn; return v }
func (v *TabsView) Value() int                            { return v.current }

// SetValue shows tab i without calling OnChange.
func (v *TabsView) SetValue(i int) { v.current = min(max(i, 0), max(len(v.pages)-1, 0)) }

func (v *TabsView) choose(cx *el.Context, i int) {
	if v.disabled || i < 0 || i >= len(v.pages) {
		return
	}
	v.focusPending = true
	cx.Focus(v.tabID(i))
	if i == v.current {
		return
	}
	v.current = i
	if v.onChange != nil {
		v.onChange(i)
	}
}

func (v *TabsView) Render(cx *el.Context) el.Element {
	bar := el.Div().Role("tablist").Row()
	v.offsets = map[int]float32{}
	x := float32(0)
	visible := v.visible()
	for _, i := range visible {
		p := v.pages[i]
		v.offsets[i] = x
		x += v.tabWidth(i)
		on := i == v.current
		tab := el.Div().ID(v.tabID(i)).Role("tab").Name(p.title).Selected(on).
			Px(14).H(el.Dp(v.height - 2)).CursorPointer().TextColor(theme.Muted).Focusable(on).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			Hover(func(s *el.Style) { s.TextColor(theme.Text) }).
			OnClick(func() { v.choose(cx, i) }).
			OnKey(func(e el.KeyEvent) bool {
				j, ok := i, true
				switch key.Name(e.Name) {
				case key.NameDeleteForward:
					if v.onClose == nil {
						return false
					}
					if e.State == el.KeyPress {
						v.onClose(i)
					}
					return true
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
			Row().Items(el.Center).Gap(theme.SpaceSm).Child(el.Text(p.title).Grow().MinW(el.Dp(0)).MaxLines(1))
		if v.avail > 0 {
			tab.MaxW(el.Dp(max(24, v.avail-60)))
		}
		if v.reorder {
			tab.OnDrag(func(e el.DragEvent) { v.dragTab(cx, i, visible, e) })
		}
		head := el.Div().Row().Items(el.Center).Child(tab)
		if v.onClose != nil {
			// Beside the tab, not inside it: a click on the button must not
			// also select the tab it is closing.
			head.Pr(theme.SpaceSm).Child(el.Div().Name(locale.Current().Name(locale.Current().Close, p.title)).P(theme.SpaceXxs).Rounded(theme.RadiusSm).
				Focusable(false).CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				OnClick(func() {
					if !v.disabled {
						v.onClose(i)
					}
				}).Child(Icon(IconClose).Size(12).Color(theme.Muted).Render(cx)))
		}
		underline := el.Div().H(el.Dp(2))
		if on {
			tab.TextColor(theme.PrimaryText)
			underline.Bg(theme.Primary)
		}
		bar.Child(el.Div().ID(v.tabID(i)+"/head").NoShrink().Items(el.Stretch).Child(head, underline).Decorate(func(gtx core.C, draw func()) {
			if px := gtx.Metric.PxPerDp; px > 0 && i < len(v.widths) {
				w := float32(gtx.Constraints.Max.X) / px
				if v.widths[i] != w {
					v.widths[i] = w
					gtx.Execute(op.InvalidateCmd{})
				}
			}
			draw()
		}))
	}
	if len(visible) < len(v.pages) {
		v.more.items = nil
		for i := range v.pages {
			if !slices.Contains(visible, i) {
				v.more.Item(v.pages[i].title, "", func() { v.choose(cx, i) })
			}
		}
		v.more.Trigger(Button("", v.more.Toggle).Name(locale.Current().More).Icon(IconChevronDown).Variant(ButtonGhost).Size(v.height - 2))
		bar.Child(v.more.Render(cx))
	}
	if v.focusPending && len(v.pages) > 0 && !v.disabled {
		cx.Focus(v.tabID(v.current))
		v.focusPending = false
	}
	header := el.Div().Row().Items(el.Center)
	id := autoID("tabs", v)
	if v.leading != nil {
		header.Child(el.Div().ID(id + "/leading").NoShrink().Child(v.leading.Render(cx)))
	}
	header.Child(el.Div().ID(id + "/bar").Grow().MinW(el.Dp(0)).Row().Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			w := float32(gtx.Constraints.Max.X) / px
			if v.avail != w {
				v.avail = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
	}).Child(bar))
	if v.trailing != nil {
		header.Child(el.Div().ID(id + "/trailing").NoShrink().Child(v.trailing.Render(cx)))
	}
	out := el.Div().Disabled(v.disabled).Gap(theme.SpaceXl).Items(el.Stretch).Child(el.Div().Items(el.Stretch).Child(header, el.Div().H(el.Dp(1)).Bg(theme.Border)))
	if v.current < len(v.pages) && v.pages[v.current].page != nil {
		out.Child(el.Div().ID(v.tabID(v.current) + "/panel").Role("tabpanel").Name(v.pages[v.current].title).Items(el.Stretch).Child(v.pages[v.current].page.Render(cx)))
	}
	return out
}
