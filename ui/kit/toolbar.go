package kit

import (
	"image"
	"strconv"
	"unicode/utf8"

	"gioui.org/op"
	"gioui.org/op/clip"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ToolbarItem is a command, separator, or custom group.
type ToolbarItem struct {
	// Content inserts an interactive view at this position. It takes precedence
	// over Action/Icon, but not Separator. Its children own their Tab stops.
	Content el.View
	// Width is the custom group's width in dp; invalid/zero values use 160dp.
	Width float32
	// OverflowContent optionally replaces Content in the overflow dialog.
	OverflowContent el.View
	Label           string
	Icon            IconName
	IconOnly        bool // with an Icon, the label becomes the tooltip
	Action          func()
	Disabled        bool
	Separator       bool
}

// ToolbarView is a row of commands. Buttons that do not fit move into a 更多
// menu at the end. Commands share one Tab stop: ← → Home End move between
// buttons. Custom groups retain their children's independent Tab stops.
type ToolbarView struct {
	customOpen        int
	items             []ToolbarItem
	active            int
	widths            []float32 // painted width of each item, dp
	avail             float32   // painted width of the toolbar, dp
	more              *MenuView
	height            float32
	disabled          bool
	leading, trailing el.View
	tips              map[int]*TooltipView
}

func Toolbar(items ...ToolbarItem) *ToolbarView {
	v := &ToolbarView{customOpen: -1, items: append([]ToolbarItem(nil), items...), more: Menu(), height: 32, tips: map[int]*TooltipView{}}
	v.widths = make([]float32, len(items))
	return v
}

// SetItems replaces the buttons.
func (v *ToolbarView) SetItems(items ...ToolbarItem) {
	v.items, v.widths, v.active = append([]ToolbarItem(nil), items...), make([]float32, len(items)), 0
	v.more.SetValue(false)
	v.customOpen = -1
	v.avail = 0
	v.tips = map[int]*TooltipView{}
}

// fit returns how many leading items fit beside the 更多 button.
func (v *ToolbarView) fit() int {
	if v.avail <= 0 {
		return 0 // Keep every action reachable in More until the command area is measured.
	}
	const gap = 4
	moreW := v.moreWidth()
	used := float32(0)
	for i := range v.widths {
		w := v.itemWidth(i)
		used += w + gap
		if used > v.avail {
			// Leave room for 更多 unless everything else fits exactly.
			for i > 0 && used-v.itemWidth(i)-gap+moreW > v.avail {
				i--
				used -= v.itemWidth(i) + gap
			}
			return i
		}
	}
	return len(v.items)
}

func (v *ToolbarView) Render(cx *el.Context) el.Element {
	id := autoID("toolbar", v)
	n := v.fit()
	var buttons []int // indexes of visible, focusable items
	for i := 0; i < n; i++ {
		if !v.items[i].Separator && v.items[i].Content == nil && !v.items[i].Disabled {
			buttons = append(buttons, i)
		}
	}
	if n < len(v.items) {
		buttons = append(buttons, -1)
	}
	if len(buttons) > 0 && !containsInt(buttons, v.active) {
		focused := cx.Focused(id + "/" + strconv.Itoa(v.active))
		v.active = buttons[0]
		if n < len(v.items) {
			v.active = -1
		}
		if focused {
			cx.Focus(id + "/" + strconv.Itoa(v.active))
		}
	}
	row := el.Div().Row().Items(el.Center).Gap(theme.SpaceXs)
	measured := map[int]el.Element{}
	move := func(to int) {
		v.active = to
		cx.Focus(id + "/" + strconv.Itoa(to))
	}
	for i := 0; i < n; i++ {
		i, it := i, v.items[i]
		var e *el.DivEl
		if it.Separator {
			e = el.Div().W(el.Dp(1)).H(el.Dp(20)).Mx(4).Bg(theme.Border)
		} else if it.Content != nil {
			e = el.Div().ID(id + "/custom/" + strconv.Itoa(i)).Role("group").Name(it.Label).Disabled(it.Disabled).W(el.Dp(v.customWidth(it))).Items(el.Stretch).Child(it.Content.Render(cx))
		} else {
			e = v.button(cx, id, i, it, buttons, move)
			if it.Icon != IconNone && it.IconOnly {
				tip := v.tips[i]
				if tip == nil {
					tip = &TooltipView{}
					v.tips[i] = tip
				}
				button := e
				tip.text = it.Label
				tip.target = el.ViewFunc(func(*el.Context) el.Element { return button })
				e = el.Div().Child(tip.Render(cx))
			}
		}
		measured[i] = e
		row.Child(e.NoShrink())
	}
	if n < len(v.items) {
		v.more.items = v.more.items[:0]
		for offset, it := range v.items[n:] {
			index := n + offset
			switch {
			case it.Separator:
				v.more.Separator()
			case it.Content != nil:
				label := it.Label
				if label == "" {
					label = locale.Current().More
				}
				v.more.Item(label, "", func() { v.customOpen = index })
				v.more.items[len(v.more.items)-1].disabled = it.Disabled
				if cx.FocusWithin(id + "/custom/" + strconv.Itoa(index)) {
					v.active = -1
					cx.Focus(id + "/-1")
				}
			default:
				v.more.Item(it.Label, "", it.Action)
				v.more.items[len(v.more.items)-1].disabled = it.Disabled
			}
		}
		v.more.Trigger(el.ViewFunc(func(cx *el.Context) el.Element {
			return v.button(cx, id, -1, ToolbarItem{Label: locale.Current().More, Icon: IconChevronDown, IconOnly: true, Action: v.more.Toggle}, buttons, move)
		}))
		row.Child(v.more.Render(cx))
	}
	outer := el.Div().ID(id).Role("toolbar").WFull().Row().Items(el.Center).Disabled(v.disabled)
	if v.leading != nil {
		outer.Child(el.Div().ID(id + "/leading").NoShrink().Child(v.leading.Render(cx)))
	}
	outer.Child(el.Div().ID(id + "/commands").Grow().W(el.Dp(0)).MinW(el.Dp(0)).Row().Items(el.Center).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			w := float32(gtx.Constraints.Max.X) / px
			if v.avail != w {
				v.avail = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
		draw()
	}).Child(row))
	if v.trailing != nil {
		outer.Child(el.Div().ID(id + "/trailing").NoShrink().Child(v.trailing.Render(cx)))
	}
	v.renderCustomOverflow(cx, id, n)
	return outer.Decorate(func(gtx core.C, draw func()) {
		for i, e := range measured {
			w, _ := cx.LayoutSize(e)
			if v.items[i].Separator {
				w += 8
			}
			if v.widths[i] != w {
				v.widths[i] = w
				gtx.Execute(op.InvalidateCmd{})
			}
		}
		draw()
	})
}

func (v *ToolbarView) button(cx *el.Context, id string, i int, it ToolbarItem, buttons []int, move func(int)) *el.DivEl {
	fg := theme.Text
	if it.Disabled || v.disabled {
		fg = theme.Muted
	}
	b := el.Div().ID(id + "/" + strconv.Itoa(i)).Role("button").Name(it.Label).Disabled(it.Disabled).
		Row().Items(el.Center).Gap(theme.SpaceSm).H(el.Dp(v.height)).Px(10).Rounded(theme.RadiusMd).TextColor(fg).TextSize(theme.TextControl).
		Focusable(i == v.active).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnClick(func() {
			if v.disabled || it.Disabled {
				return
			}
			v.active = i
			cx.Focus(id + "/" + strconv.Itoa(i))
			if it.Action != nil {
				it.Action()
			}
		}).
		OnKey(func(e el.KeyEvent) bool {
			if len(buttons) == 0 || e.Modifiers != 0 {
				return false
			}
			at := indexInt(buttons, i)
			to, ok := at, true
			switch key.Name(e.Name) {
			case key.NameRightArrow:
				to = (at + 1) % len(buttons)
			case key.NameLeftArrow:
				to = (at - 1 + len(buttons)) % len(buttons)
			case key.NameHome:
				to = 0
			case key.NameEnd:
				to = len(buttons) - 1
			default:
				ok = false
			}
			if ok && e.State == el.KeyPress && at >= 0 {
				move(buttons[to])
			}
			return ok
		})
	if !it.Disabled {
		b.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	if it.Icon != IconNone {
		b.Child(Icon(it.Icon).Size(v.iconSize()).Color(fg).Render(cx))
	}
	if !it.IconOnly || it.Icon == IconNone {
		b.Child(el.Text(it.Label).MaxLines(1))
	}
	return b
}

func containsInt(s []int, x int) bool { return indexInt(s, x) >= 0 }
func indexInt(s []int, x int) int {
	for i, y := range s {
		if y == x {
			return i
		}
	}
	return -1
}

func (v *ToolbarView) Leading(view el.View) *ToolbarView  { v.leading = view; return v }
func (v *ToolbarView) Trailing(view el.View) *ToolbarView { v.trailing = view; return v }
func (v *ToolbarView) Size(height float32) *ToolbarView {
	if height >= 24 && finiteNumber(float64(height)) && height != v.height {
		v.height = height
		v.avail = 0
		clear(v.widths)
	}
	return v
}
func (v *ToolbarView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.customOpen = -1
		v.more.SetValue(false)
	}
}
func (v *ToolbarView) SetItemDisabled(i int, on bool) {
	if i >= 0 && i < len(v.items) {
		v.items[i].Disabled = on
		if on && v.customOpen == i {
			v.customOpen = -1
		}
	}
}
func (v *ToolbarView) Items() []ToolbarItem { return append([]ToolbarItem(nil), v.items...) }
func (v *ToolbarView) iconSize() float32 {
	if v.height >= 40 {
		return 20
	}
	return 16
}
func (v *ToolbarView) moreWidth() float32 { return v.iconSize() + 20 }
func (v *ToolbarView) itemWidth(i int) float32 {
	if v.widths[i] > 0 {
		return v.widths[i]
	}
	it := v.items[i]
	if it.Separator {
		return 9
	}
	if it.Content != nil {
		return v.customWidth(it)
	}
	if it.IconOnly && it.Icon != IconNone {
		return v.moreWidth()
	}
	width := float32(utf8.RuneCountInString(it.Label))*14 + 20
	if it.Icon != IconNone {
		width += v.iconSize() + 6
	}
	return width
}
