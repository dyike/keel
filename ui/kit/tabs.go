package kit

import (
	"image/color"
	"slices"
	"strconv"

	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type TabsVariant uint8

const (
	TabsUnderline TabsVariant = iota
	TabsPill
	TabsOutline
	TabsSegmented
)

// TabItem describes a page and its label. Content replaces the visible title;
// use display-only content. Title remains the accessible and overflow name.
type TabItem struct {
	Title    string
	Page     el.View
	Icon     IconName
	Content  el.View
	Disabled bool
}

type tabPage struct {
	id uint64
	TabItem
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
	variant                TabsVariant
	scrollable             bool
	maxWidth               float32
	revealID, lastSelected uint64
	reorder                bool
	onMove                 func(int, int)
	dragID                 uint64
	offsets                map[int]float32
}

func Tabs() *TabsView { return &TabsView{more: Menu(), height: 40} }
func (v *TabsView) Add(title string, page el.View) *TabsView {
	return v.AddItem(TabItem{Title: title, Page: page})
}

func (v *TabsView) AddItem(item TabItem) *TabsView {
	v.nextID++
	v.pages = append(v.pages, tabPage{id: v.nextID, TabItem: item})
	v.widths = append(v.widths, 0)
	v.repairSelection()
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
	if v.revealID == v.pages[i].id {
		v.revealID = 0
	}
	if i == v.current {
		v.focusPending = true
	}
	v.pages = append(v.pages[:i], v.pages[i+1:]...)
	v.widths = append(v.widths[:i], v.widths[i+1:]...)
	if v.current > i || v.current >= len(v.pages) {
		v.current = max(v.current-1, 0)
	}
	v.repairSelection()
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
	if v.scrollable {
		n = len(v.pages)
	}
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
func (v *TabsView) SetValue(i int) {
	i = min(max(i, 0), max(len(v.pages)-1, 0))
	if len(v.pages) == 0 || !v.pages[i].Disabled {
		v.current = i
	}
}

func (v *TabsView) choose(cx *el.Context, i int) {
	if v.disabled || i < 0 || i >= len(v.pages) || v.pages[i].Disabled {
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
	if v.scrollable && len(v.pages) > 0 && v.lastSelected != v.pages[v.current].id && (v.lastSelected != 0 || v.revealID == 0) {
		v.revealID = v.pages[v.current].id
	}
	if len(v.pages) > 0 {
		v.lastSelected = v.pages[v.current].id
	}
	heads := make([]el.Element, 0, len(v.pages))
	if v.variant == TabsSegmented {
		bar.Bg(theme.Subtle).Rounded(theme.RadiusMd)
	}
	v.offsets = map[int]float32{}
	x := float32(0)
	visible := v.visible()
	for _, i := range visible {
		p := v.pages[i]
		v.offsets[i] = x
		x += v.tabWidth(i)
		on := i == v.current
		tab := el.Div().ID(v.tabID(i)).Role("tab").Name(p.Title).Selected(on).
			Px(14).H(el.Dp(v.height - 2)).CursorPointer().TextColor(theme.Muted).Focusable(on).Disabled(p.Disabled).DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) }).
			Rounded(theme.RadiusMd).FocusStyle(func(s *el.Style) { s.BorderColor(color.NRGBA{}); s.Bg(theme.SubtleHover) }).
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
					j = v.nextEnabled(i, 1)
				case key.NameLeftArrow:
					j = v.nextEnabled(i, -1)
				case key.NameHome:
					j = v.nextEnabled(-1, 1)
				case key.NameEnd:
					j = v.nextEnabled(len(v.pages), -1)
				default:
					ok = false
				}
				if ok && e.State == el.KeyPress {
					v.choose(cx, j)
				}
				return ok
			}).
			Row().Items(el.Center).Gap(theme.SpaceSm)
		if p.Icon != IconNone {
			c := theme.Muted
			if on && !p.Disabled {
				c = theme.PrimaryText
			}
			tab.Child(Icon(p.Icon).Size(16).Color(c).Render(cx))
		}
		if p.Content != nil {
			tab.Child(el.Div().Grow().MinW(el.Dp(0)).MaxW(el.Full).Child(p.Content.Render(cx)))
		} else {
			tab.Child(el.Text(p.Title).Grow().MinW(el.Dp(0)).MaxLines(1))
		}
		limit := v.maxWidth
		if !v.scrollable && v.avail > 0 {
			available := max(24, v.avail-60)
			if limit == 0 || available < limit {
				limit = available
			}
		}
		if limit > 0 {
			tab.MaxW(el.Dp(limit))
		}
		if v.reorder && !p.Disabled {
			tab.OnDrag(func(e el.DragEvent) { v.dragTab(cx, i, visible, e) })
		}
		head := el.Div().Row().Items(el.Center).Disabled(p.Disabled).Child(tab)
		if v.onClose != nil {
			// Beside the tab, not inside it: a click on the button must not
			// also select the tab it is closing.
			head.Pr(theme.SpaceSm).Child(el.Div().Name(locale.Current().Name(locale.Current().Close, p.Title)).P(theme.SpaceXxs).Rounded(theme.RadiusSm).
				Focusable(false).CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				OnClick(func() {
					if !v.disabled && !p.Disabled {
						v.onClose(i)
					}
				}).Child(Icon(IconClose).Size(12).Color(theme.Muted).Render(cx)))
		}
		underline := el.Div().H(el.Dp(2))
		if on {
			tab.TextColor(theme.PrimaryText)
			if v.variant == TabsUnderline {
				underline.Bg(theme.Primary)
			}
		}
		switch v.variant {
		case TabsPill:
			head.Rounded(theme.RadiusFull)
			if on {
				head.Bg(theme.Highlight)
			}
		case TabsOutline:
			head.Rounded(theme.RadiusMd).Border(1, theme.Border)
			if on {
				head.Border(1, theme.Primary).Bg(theme.Highlight)
			}
		case TabsSegmented:
			head.Rounded(theme.RadiusMd)
			if on {
				head.Bg(theme.Surface).Shadow(theme.ElevationSm)
			}
		}
		headBox := el.Div().ID(v.tabID(i)+"/head").NoShrink().Items(el.Stretch).Child(head, underline).Decorate(func(gtx core.C, draw func()) {
			if px := gtx.Metric.PxPerDp; px > 0 && i < len(v.widths) {
				w := float32(gtx.Constraints.Max.X) / px
				if v.widths[i] != w {
					v.widths[i] = w
					gtx.Execute(op.InvalidateCmd{})
				}
			}
			draw()
		})
		heads = append(heads, headBox)
		bar.Child(headBox)
	}
	if len(visible) < len(v.pages) {
		v.more.items = nil
		for i := range v.pages {
			if !slices.Contains(visible, i) {
				v.more.items = append(v.more.items, menuItem{label: v.pages[i].Title, disabled: v.pages[i].Disabled, action: func() { v.choose(cx, i) }})
			}
		}
		v.more.Trigger(Button("", v.more.Toggle).Name(locale.Current().More).Icon(IconChevronDown).Variant(ButtonGhost).Size(v.height - 2))
		bar.Child(v.more.Render(cx))
	}
	if v.focusPending && (!v.scrollable || v.revealID == 0) && len(v.pages) > 0 && !v.disabled && !v.pages[v.current].Disabled {
		cx.Focus(v.tabID(v.current))
		v.focusPending = false
	}
	header := el.Div().Row().Items(el.Center)
	id := autoID("tabs", v)
	if v.leading != nil {
		header.Child(el.Div().ID(id + "/leading").NoShrink().Child(v.leading.Render(cx)))
	}
	viewport := el.Div().ID(id + "/bar").Grow().MinW(el.Dp(0)).Row()
	if v.scrollable {
		viewport.ScrollX()
	}
	header.Child(viewport.Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			w := float32(gtx.Constraints.Max.X) / px
			if v.avail != w {
				v.avail = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
		if v.scrollable {
			for i, head := range heads {
				v.widths[i], _ = cx.LayoutSize(head)
			}
		}
		if v.scrollable && v.revealID != 0 && gtx.Enabled() {
			left := float32(0)
			for i, head := range heads {
				width, _ := cx.LayoutSize(head)
				if v.pages[i].id == v.revealID {
					before, view, _ := cx.ScrollStateX(id + "/bar")
					if view > 0 {
						cx.ScrollIntoViewX(id+"/bar", left, left+width)
						after, _, _ := cx.ScrollStateX(id + "/bar")
						if before != after {
							gtx.Execute(op.InvalidateCmd{})
						} else {
							v.revealID = 0
							if v.focusPending {
								gtx.Execute(op.InvalidateCmd{})
							}
						}
					}
					break
				}
				left += width
			}
		}
	}).Child(bar))
	if v.trailing != nil {
		header.Child(el.Div().ID(id + "/trailing").NoShrink().Child(v.trailing.Render(cx)))
	}
	headerBox := el.Div().Items(el.Stretch).Child(header)
	if v.variant == TabsUnderline {
		headerBox.Child(el.Div().H(el.Dp(1)).Bg(theme.Border))
	}
	out := el.Div().Disabled(v.disabled).Gap(theme.SpaceXl).Items(el.Stretch).Child(headerBox)
	if v.current < len(v.pages) && v.pages[v.current].Page != nil {
		out.Child(el.Div().ID(v.tabID(v.current) + "/panel").Role("tabpanel").Name(v.pages[v.current].Title).Items(el.Stretch).Child(v.pages[v.current].Page.Render(cx)))
	}
	return out
}

// Variant selects the tab appearance. Underline is the default.
func (v *TabsView) Variant(variant TabsVariant) *TabsView {
	if variant <= TabsSegmented {
		v.variant = variant
	}
	return v
}

// SetItem changes a tab without replacing its stable identity or firing callbacks.
func (v *TabsView) SetItem(i int, item TabItem) {
	if i < 0 || i >= len(v.pages) {
		return
	}
	v.pages[i].TabItem = item
	v.widths[i] = 0
	v.repairSelection()
}

// SetItemDisabled skips this tab in pointer, keyboard, overflow and drag interaction.
// Disabling the selected tab selects the next enabled tab without OnChange.
// If all tabs are disabled, the current page remains displayed without a tab stop.
func (v *TabsView) SetItemDisabled(i int, disabled bool) {
	if i < 0 || i >= len(v.pages) {
		return
	}
	v.pages[i].Disabled = disabled
	if disabled && v.dragID == v.pages[i].id {
		v.dragID = 0
	}
	v.repairSelection()
}

func (v *TabsView) nextEnabled(start, delta int) int {
	n := len(v.pages)
	for k := 0; k < n; k++ {
		start = (start + delta + n) % n
		if !v.pages[start].Disabled {
			return start
		}
	}
	return -1
}
func (v *TabsView) repairSelection() {
	if len(v.pages) == 0 || !v.pages[v.current].Disabled {
		return
	}
	if i := v.nextEnabled(v.current, 1); i >= 0 {
		v.current = i
		v.focusPending = true
	}
}

// MaxWidth caps each tab's label area in dp, excluding the close button.
// Zero removes the cap; negative and non-finite values are ignored.
func (v *TabsView) MaxWidth(dp float32) *TabsView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.maxWidth = dp
		clear(v.widths)
		if len(v.pages) > 0 {
			v.revealID = v.pages[v.current].id
		}
	}
	return v
}

// Scrollable shows all tabs in a horizontal viewport instead of an overflow menu.
func (v *TabsView) Scrollable(on bool) *TabsView {
	if v.scrollable != on {
		v.scrollable = on
		v.more.SetValue(false)
		if len(v.pages) > 0 {
			v.revealID = v.pages[v.current].id
		}
	}
	return v
}

// ScrollTo reveals a tab in scrollable mode without selecting it or firing callbacks.
// It is applied after layout, including when called before the first frame.
func (v *TabsView) ScrollTo(i int) {
	if i >= 0 && i < len(v.pages) {
		v.revealID = v.pages[i].id
	}
}

// ScrollState reports horizontal offset, viewport and content widths in dp.
func (v *TabsView) ScrollState(cx *el.Context) (offset, viewport, content float32) {
	return cx.ScrollStateX(autoID("tabs", v) + "/bar")
}
