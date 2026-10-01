package widget

import (
	"gioui.org/io/key"
	giolayout "gioui.org/layout"

	"image"
)

// ToggleGroupView selects one option by default, or any subset in Multiple mode.
// Selecting the active option again clears it. Values are returned in display order.
type ToggleGroupView struct {
	items              []*ToggleButton
	multiple, disabled bool
	onChange           func([]string)
}

func ToggleGroup(options ...string) *ToggleGroupView {
	g := &ToggleGroupView{}
	for _, label := range options {
		// Duplicate values are ignored so a value always identifies exactly one item.
		duplicate := false
		for _, t := range g.items {
			if t.text == label {
				duplicate = true
			}
		}
		if duplicate {
			continue
		}
		t := Toggle(label, false)
		t.OnChange(func(on bool) {
			if on && !g.multiple {
				for _, other := range g.items {
					if other != t {
						other.SetValue(false)
					}
				}
			}
			if g.onChange != nil {
				g.onChange(g.Values())
			}
		})
		g.items = append(g.items, t)
	}
	return g
}
func (g *ToggleGroupView) Multiple() *ToggleGroupView                  { g.multiple = true; return g }
func (g *ToggleGroupView) OnChange(fn func([]string)) *ToggleGroupView { g.onChange = fn; return g }
func (g *ToggleGroupView) Size(size ComponentSize) *ToggleGroupView {
	for _, t := range g.items {
		t.Size(size)
	}
	return g
}
func (g *ToggleGroupView) Ghost() *ToggleGroupView {
	for _, t := range g.items {
		t.Ghost()
	}
	return g
}
func (g *ToggleGroupView) SetDisabled(v bool) { g.disabled = v }
func (g *ToggleGroupView) SetItemDisabled(i int, v bool) {
	if i >= 0 && i < len(g.items) {
		g.items[i].SetDisabled(v)
	}
}
func (g *ToggleGroupView) Values() []string {
	var out []string
	for _, t := range g.items {
		if t.Value() {
			out = append(out, t.text)
		}
	}
	return out
}

// SetValues ignores unknown values and updates silently. Single mode selects
// the first matching item in display order.
func (g *ToggleGroupView) SetValues(values ...string) {
	chosen := false
	for _, t := range g.items {
		on := false
		for _, v := range values {
			if t.text == v {
				on = true
			}
		}
		if !g.multiple && chosen {
			on = false
		}
		t.SetValue(on)
		chosen = chosen || on
	}
}
func (g *ToggleGroupView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if g.disabled {
		gtx = gtx.Disabled()
	}
	for i, t := range g.items {
		if g.disabled || t.disabled {
			continue
		}
		for {
			ev, ok := gtx.Event(key.Filter{Focus: &t.click, Name: key.NameLeftArrow}, key.Filter{Focus: &t.click, Name: key.NameRightArrow}, key.Filter{Focus: &t.click, Name: key.NameHome}, key.Filter{Focus: &t.click, Name: key.NameEnd})
			if !ok {
				break
			}
			e, ok := ev.(key.Event)
			if !ok || e.State != key.Press {
				continue
			}
			next, step := i, 1
			switch e.Name {
			case key.NameLeftArrow:
				step = -1
			case key.NameHome:
				next = -1
			case key.NameEnd:
				next = len(g.items)
				step = -1
			}
			for n := 0; n < len(g.items); n++ {
				next = (next + step + len(g.items)) % len(g.items)
				if !g.items[next].disabled {
					gtx.Execute(key.FocusCmd{Tag: &g.items[next].click})
					break
				}
			}
		}
	}
	var children []giolayout.FlexChild
	for i, t := range g.items {
		if i > 0 {
			children = append(children, giolayout.Rigid(giolayout.Spacer{Width: 6}.Layout))
		}
		children = append(children, giolayout.Rigid(t.Layout))
	}
	return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx, children...)
}
