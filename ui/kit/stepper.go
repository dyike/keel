package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// StepperView shows progress through a sequence of steps. Steps before the
// current one are done; with Navigable the user may click back to them.
type StepperView struct {
	steps     []string
	current   int
	navigable bool
	disabled  bool
	onChange  func(int)
}

func Stepper(steps ...string) *StepperView { return &StepperView{steps: slices.Clone(steps)} }

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
	row := el.Div().Role("list").Row().Items(el.Center).Gap(8)
	for i, s := range v.steps {
		i := i
		state, bg, fg, mark := "upcoming", theme.Subtle, theme.Muted, el.Element(el.Text(strconv.Itoa(i+1)).TextSize(13))
		switch {
		case i < v.current:
			state, bg, fg = "done", theme.Primary, theme.OnColor
			mark = Icon(IconDone).Size(14).Color(theme.OnColor).Render(cx)
		case i == v.current:
			state, bg, fg = "current", theme.Primary, theme.OnColor
		}
		dot := el.Div().Size(el.Dp(24)).NoShrink().Rounded(12).Bg(bg).TextColor(fg).Center().Child(mark)
		label := el.Text(s).MaxLines(1)
		if i > v.current {
			label.TextColor(theme.Muted)
		}
		if i == v.current {
			label.Bold()
		}
		step := el.Div().ID(id+"/"+strconv.Itoa(i)).Role("step").Name(s).Value(state).Row().Items(el.Center).Gap(6).Child(dot, label)
		if v.navigable && i < v.current {
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
			row.Child(el.Div().W(el.Dp(32)).H(el.Dp(2)).Rounded(1).Bg(line))
		}
		row.Child(step)
	}
	return el.Div().ID(id + "/scroll").WFull().Disabled(v.disabled).ScrollX().Child(row)
}
