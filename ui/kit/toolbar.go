package kit

import (
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ToolbarItem is one button of a Toolbar, or a separator.
type ToolbarItem struct {
	Label     string
	Icon      IconName
	HasIcon   bool // show Icon; with IconOnly the label becomes the tooltip
	IconOnly  bool
	Action    func()
	Disabled  bool
	Separator bool
}

// ToolbarView is a row of commands. Buttons that do not fit move into a 更多
// menu at the end. It is one Tab stop: ← → Home End move between buttons.
type ToolbarView struct {
	items  []ToolbarItem
	active int
	widths []float32 // painted width of each item, dp
	avail  float32   // painted width of the toolbar, dp
	more   *MenuView
}

func Toolbar(items ...ToolbarItem) *ToolbarView {
	v := &ToolbarView{items: items, more: Menu()}
	v.widths = make([]float32, len(items))
	return v
}

// SetItems replaces the buttons.
func (v *ToolbarView) SetItems(items ...ToolbarItem) {
	v.items, v.widths, v.active = items, make([]float32, len(items)), 0
}

// fit returns how many leading items fit beside the 更多 button.
func (v *ToolbarView) fit() int {
	if v.avail <= 0 {
		return len(v.items) // first frame: nothing measured yet
	}
	const gap, moreW = 4, 36
	used := float32(0)
	for i, w := range v.widths {
		used += w + gap
		if used > v.avail {
			// Leave room for 更多 unless everything else fits exactly.
			for i > 0 && used-v.widths[i]-gap+moreW > v.avail {
				i--
				used -= v.widths[i] + gap
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
		if !v.items[i].Separator && !v.items[i].Disabled {
			buttons = append(buttons, i)
		}
	}
	if len(buttons) > 0 && !containsInt(buttons, v.active) {
		v.active = buttons[0]
	}
	row := el.Div().Role("toolbar").Row().Items(el.Center).Gap(4)
	move := func(to int) {
		v.active = to
		cx.Focus(id + "/" + strconv.Itoa(to))
	}
	for i := 0; i < n; i++ {
		i, it := i, v.items[i]
		var e *el.DivEl
		if it.Separator {
			e = el.Div().W(el.Dp(1)).H(el.Dp(20)).Mx(4).Bg(theme.Border)
		} else {
			e = v.button(cx, id, i, it, buttons, move)
		}
		row.Child(e.Decorate(func(gtx core.C, draw func()) {
			if px := gtx.Metric.PxPerDp; px > 0 && i < len(v.widths) {
				v.widths[i] = float32(gtx.Constraints.Max.X) / px
			}
			draw()
		}))
	}
	if n < len(v.items) {
		v.more.items = v.more.items[:0]
		for _, it := range v.items[n:] {
			switch {
			case it.Separator:
				v.more.Separator()
			default:
				v.more.Item(it.Label, "", it.Action)
				v.more.SetItemDisabled(it.Label, it.Disabled)
			}
		}
		v.more.Trigger(Button("", v.more.Toggle).Name(locale.Current().More).Icon(IconChevronDown).Variant(ButtonGhost).Size(32))
		row.Child(v.more.Render(cx))
	}
	// The stretched wrapper measures the width the parent allows; a parent
	// sized by its content allows the toolbar's own width.
	return el.Div().Row().Items(el.Center).Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			v.avail = float32(gtx.Constraints.Max.X) / px
		}
		draw()
	}).Child(row)
}

func (v *ToolbarView) button(cx *el.Context, id string, i int, it ToolbarItem, buttons []int, move func(int)) *el.DivEl {
	fg := theme.Text
	if it.Disabled {
		fg = theme.Muted
	}
	b := el.Div().ID(id + "/" + strconv.Itoa(i)).Role("button").Name(it.Label).Disabled(it.Disabled).
		Row().Items(el.Center).Gap(6).H(el.Dp(32)).Px(10).Rounded(6).TextColor(fg).TextSize(14).
		Focusable(i == v.active).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnClick(func() {
			v.active = i
			if it.Action != nil {
				it.Action()
			}
		}).
		OnKey(func(e el.KeyEvent) bool {
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
	if it.HasIcon {
		b.Child(Icon(it.Icon).Size(16).Color(fg).Render(cx))
	}
	if !it.IconOnly || !it.HasIcon {
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
