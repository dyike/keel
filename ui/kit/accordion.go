package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type AccordionSize uint8

const (
	AccordionSizeMedium AccordionSize = iota
	AccordionSizeXSmall
	AccordionSizeSmall
	AccordionSizeLarge
)

type accordionItem struct {
	title    string
	body     el.View
	heading  el.View
	motion   disclosureMotion
	disabled bool
}

// AccordionView stacks sections that expand and collapse. One section is open
// at a time unless Multiple. Headers take focus: ↑ ↓ Home End move between
// them, Enter or Space toggles.
type AccordionView struct {
	items      []accordionItem
	open       []int
	multiple   bool
	disabled   bool
	onChange   func(open []int)
	borderless bool
	size       AccordionSize
}

// Bordered controls the outer border and section separators, on by default.
func (v *AccordionView) Bordered(on bool) *AccordionView { v.borderless = !on; return v }

// Size sets the title, chevron, spacing and inherited body text scale.
func (v *AccordionView) Size(size AccordionSize) *AccordionView {
	if size <= AccordionSizeLarge {
		v.size = size
	}
	return v
}
func (v *AccordionView) style() disclosureStyle {
	s := defaultDisclosureStyle()
	switch v.size {
	case AccordionSizeXSmall:
		s = disclosureStyle{theme.SpaceMd, theme.SpaceSm, theme.SpaceMd, theme.TextSm, 12, theme.SpaceXs}
	case AccordionSizeSmall:
		s = disclosureStyle{theme.SpaceLg, theme.SpaceMd, theme.SpaceLg, theme.TextMd, 14, theme.SpaceSm}
	case AccordionSizeLarge:
		s = disclosureStyle{theme.SpaceXl, theme.SpaceXl, theme.SpaceXl, theme.TextLg, 20, theme.SpaceLg}
	}
	return s
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
	if i < 0 || i >= len(v.items) || v.disabled || v.items[i].disabled {
		return
	}
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

// Heading replaces an item's visual heading while keeping its title as the
// accessible name. The view should contain display content, not nested controls.
func (v *AccordionView) Heading(i int, heading el.View) *AccordionView {
	if i >= 0 && i < len(v.items) {
		v.items[i].heading = heading
	}
	return v
}
func (v *AccordionView) SetDisabled(on bool) { v.disabled = on }
func (v *AccordionView) triggerID(i int) string {
	return autoID("accordion", v) + "/" + strconv.Itoa(i)
}

// Trigger renders just one header, for composition with Content.
func (v *AccordionView) Trigger(i int) el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		if i < 0 || i >= len(v.items) {
			return el.Div().Hidden(true)
		}
		it := v.items[i]
		enabled := func(from, d int) int {
			for k := 1; k <= len(v.items); k++ {
				j := ((from+d*k)%len(v.items) + len(v.items)) % len(v.items)
				if !v.items[j].disabled && cx.Enabled(v.triggerID(j)) {
					return j
				}
			}
			return from
		}
		return disclosureTrigger(cx, v.triggerID(i), it.title, it.heading, slices.Contains(v.open, i), v.disabled || it.disabled, func() { v.toggle(i) }, v.style()).
			OnKey(func(e el.KeyEvent) bool {
				j := i
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
					return false
				}
				if e.State == el.KeyPress && j >= 0 && j < len(v.items) {
					cx.Focus(v.triggerID(j))
				}
				return true
			})
	})
}

// Content renders an item's body with an interruptible expand/collapse animation.
func (v *AccordionView) Content(i int) el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		if i < 0 || i >= len(v.items) {
			return el.Div().Hidden(true)
		}
		it := &v.items[i]
		return disclosureContent(cx, autoID("accordion", v)+"/content/"+strconv.Itoa(i), v.triggerID(i), it.body, slices.Contains(v.open, i), v.disabled || it.disabled, &it.motion, v.style())
	})
}
func (v *AccordionView) Render(cx *el.Context) el.Element {
	out := el.Div().Role("group").Items(el.Stretch).Rounded(theme.RadiusLg).Bg(theme.Surface)
	if !v.borderless {
		out.Border(1, theme.Border)
	}
	for i := range v.items {
		if i > 0 {
			out.Child(el.Div().H(el.Dp(1)).Bg(theme.Border).Hidden(v.borderless))
		}
		out.Child(v.Trigger(i).Render(cx), v.Content(i).Render(cx))
	}
	return out
}
