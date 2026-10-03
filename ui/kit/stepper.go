package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// StepperItem describes a step. Content optionally replaces the visible label;
// Label remains its accessible name. Content should contain no interactive controls.
type StepperItem struct {
	Label    string
	Icon     IconName
	Disabled bool
	Content  el.View
}

// StepperView shows progress through a sequence of steps. Steps before the
// current one are done; with Navigable the user may click back to them.
type StepperView struct {
	steps     []StepperItem
	vertical  bool
	size      float32
	current   int
	navigable bool
	disabled  bool
	onChange  func(int)
}

func Stepper(steps ...string) *StepperView {
	v := &StepperView{size: 24}
	for _, label := range steps {
		v.steps = append(v.steps, StepperItem{Label: label})
	}
	return v
}

// SetEntries copies the items and clamps the current index, without OnChange.
// Views in Content remain owned by the caller and are not deep-copied.
func (v *StepperView) SetEntries(items ...StepperItem) {
	v.steps = slices.Clone(items)
	v.SetValue(v.current)
}
func (v *StepperView) Entries() []StepperItem { return slices.Clone(v.steps) }
func (v *StepperView) SetItemDisabled(index int, on bool) {
	if index >= 0 && index < len(v.steps) {
		v.steps[index].Disabled = on
	}
}

// Vertical stacks steps and their connecting lines from top to bottom.
func (v *StepperView) Vertical() *StepperView { v.vertical = true; return v }

// Size sets the marker diameter in dp; 20, 24 (default), and 32 are typical.
func (v *StepperView) Size(dp float32) *StepperView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}

// Navigable lets the user click a finished step to return to it.
func (v *StepperView) Navigable() *StepperView            { v.navigable = true; return v }
func (v *StepperView) OnChange(fn func(int)) *StepperView { v.onChange = fn; return v }

func (v *StepperView) SetDisabled(on bool) { v.disabled = on }

// Value is the index of the current step; len(steps) means all are done.
func (v *StepperView) Value() int { return v.current }

// SetValue moves to step i without calling OnChange, e.g. after "下一步".
func (v *StepperView) SetValue(i int) { v.current = min(max(i, 0), len(v.steps)) }

func (v *StepperView) Render(cx *el.Context) el.Element {
	id := autoID("stepper", v)
	size := v.size
	if size <= 0 {
		size = 24
	}
	font, iconSize := float32(theme.TextMd), size*14/24
	if size <= 20 {
		font = theme.TextSm
	} else if size >= 32 {
		font = theme.TextBody
	}
	row := el.Div().Role("list").Gap(theme.SpaceMd)
	if v.vertical {
		row.WFull().Items(el.Start)
	} else {
		row.Row().Items(el.Center)
	}
	for i, s := range v.steps {
		i := i
		state, bg, fg := "upcoming", theme.Subtle, theme.Muted
		icon := s.Icon
		switch {
		case i < v.current:
			state, bg, fg = "done", theme.Primary, theme.OnColor
			if icon == IconNone {
				icon = IconDone
			}
		case i == v.current:
			state, bg, fg = "current", theme.Primary, theme.OnColor
		}
		if s.Disabled {
			bg, fg = theme.Subtle, theme.Muted
		}
		var mark el.Element = el.Text(strconv.Itoa(i + 1)).TextSize(font)
		if icon != IconNone {
			mark = Icon(icon).Size(iconSize).Color(fg).Render(cx)
		}

		dot := el.Div().Size(el.Dp(size)).NoShrink().Rounded(theme.RadiusFull).Bg(bg).TextColor(fg).Center().Child(mark)
		label := el.Text(s.Label).TextSize(font).MaxLines(1)
		if i > v.current || s.Disabled {
			label.TextColor(theme.Muted)
		}
		if i == v.current {
			label.Bold()
		}
		step := el.Div().ID(id + "/" + strconv.Itoa(i)).Role("step").Name(s.Label).Value(state).Disabled(s.Disabled).Row().Items(el.Center).Gap(theme.SpaceSm).TextSize(font).Child(dot)
		if s.Content != nil {
			step.Child(s.Content.Render(cx))
		} else {
			step.Child(label)
		}
		if v.vertical {
			step.MaxW(el.Full)
		}
		if v.navigable && i < v.current && !s.Disabled {
			step.Focusable(true).CursorPointer().OnClick(func() {
				v.current = i
				if v.onChange != nil {
					v.onChange(i)
				}
			})
		}
		if i > 0 {
			line := theme.Border
			if i <= v.current {
				line = theme.Primary
			}
			connector := el.Div().NoShrink().Rounded(theme.RadiusFull).Bg(line)
			if v.vertical {
				connector.W(el.Dp(2)).H(el.Dp(size * 4 / 3))
				row.Child(el.Div().NoShrink().Pl(max(0, (size-2)/2)).Child(connector))
			} else {
				connector.W(el.Dp(size * 4 / 3)).H(el.Dp(2))
				row.Child(connector)
			}
		}
		row.Child(step)
	}
	viewport := el.Div().ID(id + "/scroll").WFull().Disabled(v.disabled).Child(row)
	if v.vertical {
		return viewport.ScrollY()
	}
	return viewport.ScrollX().Pb(theme.SpaceSm)
}
