package window

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"image"
	"strings"
)

// applicationMenuView belongs to a window, so open menus never leak between windows.
type applicationMenuView struct {
	buttons    map[string]*widget.Clickable
	path       []string
	active     string
	positions  map[string]image.Rectangle
	outside    struct{}
	list       widget.List
	listPath   string
	bar        *MenuBar
	generation uint64
}

func currentMenu() (*MenuBar, uint64) {
	applicationMenu.Lock()
	defer applicationMenu.Unlock()
	return applicationMenu.current, applicationMenu.generation
}
func (v *applicationMenuView) button(id string) *widget.Clickable {
	if v.buttons == nil {
		v.buttons = map[string]*widget.Clickable{}
	}
	b := v.buttons[id]
	if b == nil {
		b = new(widget.Clickable)
		v.buttons[id] = b
	}
	return b
}
func (v *applicationMenuView) reset() {
	v.path = nil
	v.active = ""
}

func (w *Window) belowApplicationMenu(gtx core.C) (core.C, func()) {
	v := &w.menuView
	bar, generation := currentMenu()
	show := w.opts.MenuDisplay == MenuDisplayWindow || (w.opts.MenuDisplay == MenuDisplayAuto && platformDrawApplicationMenu(w))
	if !show || bar == nil {
		return gtx, func() {}
	}
	if v.bar != bar || v.generation != generation {
		v.reset()
		v.buttons = nil
		v.bar = bar
		v.generation = generation
	}
	items := bar.Items()
	if len(items) == 0 {
		return gtx, func() {}
	}
	v.keyboard(gtx, items)
	height := min(gtx.Dp(30), gtx.Constraints.Max.Y)
	v.positions = map[string]image.Rectangle{}
	paint.FillShape(gtx.Ops, theme.Subtle, clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, height)).Op())
	x := 0
	for _, item := range items {
		if item.Separator {
			continue
		}
		b := v.button(item.ID)
		for b.Clicked(gtx) {
			gtx.Execute(op.InvalidateCmd{})
			if item.Disabled {
				continue
			}
			if len(item.Children) > 0 {
				if len(v.path) > 0 && v.path[0] == item.ID {
					v.reset()
				} else {
					v.path = []string{item.ID}
					v.active = firstMenuItem(item.Children)

				}
			} else {
				core.Call(gtx, func() { bar.Invoke(item.ID) })
			}
		}
		cell := gtx
		cell.Constraints = layout.Constraints{Max: image.Pt(min(max(0, gtx.Constraints.Max.X-x), gtx.Constraints.Max.X/max(1, len(items))), height)}
		offset := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
		dims := b.Layout(cell, func(g core.C) core.D {
			if item.Disabled {
				g = g.Disabled()
			}
			if len(v.path) > 0 && v.path[0] == item.ID {
				paint.FillShape(g.Ops, theme.Border, clip.Rect{Max: g.Constraints.Max}.Op())
			}
			return layout.Inset{Left: 8, Right: 8, Top: 5, Bottom: 5}.Layout(g, func(g core.C) core.D {
				label := material.Label(theme.Material, theme.Material.TextSize, item.Title)
				label.Color = theme.Text
				label.MaxLines = 1
				label.Truncator = "…"
				return label.Layout(g)
			})
		})
		offset.Pop()
		v.positions[item.ID] = image.Rect(x, 0, x+dims.Size.X, height)
		x += dims.Size.X
	}
	// Register navigation immediately when a pointer or F10 opens the menu.
	if len(v.path) > 0 {
		v.keyboard(gtx, items)
	}
	// Defer the popup until after content and application overlays have drawn.
	if len(v.path) > 0 {
		macro := op.Record(gtx.Ops)
		v.popup(gtx, items, height)
		call := macro.Stop()
		op.Defer(gtx.Ops, call)
	}
	offset := op.Offset(image.Pt(0, height)).Push(gtx.Ops)
	gtx.Constraints.Max.Y = max(0, gtx.Constraints.Max.Y-height)
	gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	return gtx, func() { area.Pop(); offset.Pop() }
}
func firstMenuItem(items []MenuItem) string {
	for _, item := range items {
		if !item.Disabled && !item.Separator {
			return item.ID
		}
	}
	return ""
}
func (v *applicationMenuView) level(items []MenuItem) ([]MenuItem, bool) {
	for _, id := range v.path {
		item, disabled := findMenuItem(items, id, false)
		if item == nil || disabled {
			return nil, false
		}
		items = item.Children
	}
	return items, true
}
func (v *applicationMenuView) activate(gtx core.C, item MenuItem) {
	if item.ID == "@keel/view/back" && len(v.path) > 1 {
		v.path = v.path[:len(v.path)-1]
		items := v.bar.Items()
		level, _ := v.level(items)
		v.active = firstMenuItem(level)
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	if item.Disabled || item.Separator {
		return
	}
	if len(item.Children) > 0 {
		v.path = append(v.path, item.ID)
		v.active = firstMenuItem(item.Children)

		return
	}
	v.reset()
	core.Call(gtx, func() { v.bar.Invoke(item.ID) })
}
func (v *applicationMenuView) keyboard(gtx core.C, items []MenuItem) {
	filters := []event.Filter{key.Filter{Name: "F10"}, key.Filter{Name: "M", Required: key.ModAlt}}
	if len(v.path) > 0 {
		for _, name := range []key.Name{key.NameEscape, key.NameUpArrow, key.NameDownArrow, key.NameLeftArrow, key.NameRightArrow, key.NameReturn, key.NameSpace} {
			filters = append(filters, key.Filter{Name: name})
		}
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		k, ok := ev.(key.Event)
		if !ok || k.State != key.Press {
			continue
		}
		switch k.Name {
		case "F10", "M":
			if len(v.path) > 0 {
				v.reset()
			} else {
				for _, item := range items {
					if !item.Disabled && len(item.Children) > 0 {
						v.path = []string{item.ID}
						v.active = firstMenuItem(item.Children)
						break
					}
				}
			}
		case key.NameEscape:
			v.reset()
		case key.NameLeftArrow:
			if len(v.path) > 1 {
				v.path = v.path[:len(v.path)-1]
				level, _ := v.level(items)
				v.active = firstMenuItem(level)
			} else {
				v.switchTop(items, -1)
			}
		case key.NameRightArrow:
			level, _ := v.level(items)
			item, _ := findMenuItem(level, v.active, false)
			if item != nil && len(item.Children) > 0 {
				v.activate(gtx, *item)
			} else {
				v.switchTop(items, 1)
			}
		case key.NameDownArrow, key.NameUpArrow:
			level, _ := v.level(items)
			enabled := []string{}
			index := -1
			for _, item := range level {
				if !item.Disabled && !item.Separator {
					if item.ID == v.active {
						index = len(enabled)
					}
					enabled = append(enabled, item.ID)
				}
			}
			if len(enabled) > 0 {
				delta := 1
				if k.Name == key.NameUpArrow {
					delta = -1
				}
				index = (index + delta + len(enabled)) % len(enabled)
				v.active = enabled[index]
				v.list.ScrollTo(index)
			}
		case key.NameReturn, key.NameSpace:
			level, _ := v.level(items)
			item, _ := findMenuItem(level, v.active, false)
			if item != nil {
				v.activate(gtx, *item)
			}
		}
		gtx.Execute(op.InvalidateCmd{})
	}
}
func (v *applicationMenuView) switchTop(items []MenuItem, delta int) {
	enabled := []MenuItem{}
	index := 0
	for _, item := range items {
		if !item.Disabled && len(item.Children) > 0 {
			if len(v.path) > 0 && item.ID == v.path[0] {
				index = len(enabled)
			}
			enabled = append(enabled, item)
		}
	}
	if len(enabled) > 0 {
		item := enabled[(index+delta+len(enabled))%len(enabled)]
		v.path = []string{item.ID}
		v.active = firstMenuItem(item.Children)

	}
}
func (v *applicationMenuView) popup(gtx core.C, items []MenuItem, barHeight int) {
	level, ok := v.level(items)
	pathKey := strings.Join(v.path, "\x00")
	if v.listPath != pathKey {
		v.list = widget.List{}
		v.listPath = pathKey
	}
	if !ok {
		v.reset()
		return
	}
	if len(v.path) > 1 {
		level = append([]MenuItem{{ID: "@keel/view/back", Title: "‹ Back"}}, level...)
	}
	// Full-window scrim consumes outside clicks without stealing editor key focus.
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &v.outside, Kinds: pointer.Press})
		if !ok {
			break
		}
		if _, ok := ev.(pointer.Event); ok {
			gtx.Execute(op.InvalidateCmd{})
			v.reset()
			return
		}
	}
	screen := clip.Rect(image.Rect(0, barHeight, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)).Push(gtx.Ops)
	event.Op(gtx.Ops, &v.outside)
	screen.Pop()
	width := min(gtx.Dp(280), gtx.Constraints.Max.X)
	x := v.positions[v.path[0]].Min.X
	x = min(x, max(0, gtx.Constraints.Max.X-width))
	top := barHeight
	height := min(gtx.Dp(unit.Dp(len(level)*32+8)), max(0, gtx.Constraints.Max.Y-top))
	if width <= 0 || height <= 0 {
		return
	}
	offset := op.Offset(image.Pt(x, top)).Push(gtx.Ops)
	defer offset.Pop()
	rect := image.Rect(0, 0, width, height)
	area := clip.Rect(rect).Push(gtx.Ops)
	defer area.Pop()
	paint.FillShape(gtx.Ops, theme.Surface, clip.Rect(rect).Op())
	semantic.LabelOp("Application menu").Add(gtx.Ops)
	popup := gtx
	popup.Constraints = layout.Exact(rect.Size())
	v.list.Axis = layout.Vertical
	material.List(theme.Material, &v.list).Layout(popup, len(level), func(g core.C, i int) core.D {
		item := level[i]
		if item.Separator {
			h := g.Dp(8)
			paint.FillShape(g.Ops, theme.Border, clip.Rect(image.Rect(8, h/2, width-8, h/2+1)).Op())
			return core.D{Size: image.Pt(width, h)}
		}
		b := v.button(item.ID)
		for b.Clicked(g) {
			v.activate(g, item)
		}

		g.Constraints = layout.Exact(image.Pt(width, g.Dp(32)))
		if item.Disabled {
			g = g.Disabled()
		}
		return b.Layout(g, func(g core.C) core.D {
			if item.ID == v.active || b.Hovered() {
				paint.FillShape(g.Ops, theme.Subtle, clip.Rect{Max: g.Constraints.Max}.Op())
			}
			title := item.Title
			if item.Checked {
				title = "✓ " + title
			} else {
				title = "   " + title
			}
			suffix := ""
			if len(item.Children) > 0 {
				suffix = "›"
			} else if item.Shortcut != "" {
				suffix = menuWireShortcut(wireMenuItems([]MenuItem{item})[0])
			}
			return layout.Inset{Left: 10, Right: 10, Top: 6, Bottom: 6}.Layout(g, func(g core.C) core.D {
				return layout.Flex{}.Layout(g, layout.Flexed(1, func(g core.C) core.D {
					l := material.Label(theme.Material, theme.Material.TextSize, title)
					l.Color = theme.Text
					if item.Disabled {
						l.Color = theme.Muted
					}
					l.MaxLines = 1
					l.Truncator = "…"
					return l.Layout(g)
				}), layout.Rigid(func(g core.C) core.D {
					l := material.Label(theme.Material, theme.Material.TextSize, suffix)
					l.Color = theme.Muted
					return l.Layout(g)
				}))
			})
		})
	})
}
