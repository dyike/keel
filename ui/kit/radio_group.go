package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// RadioGroupView picks one of several options. Tab enters the group at the
// chosen option; arrow keys move the choice, like native radio groups.
type RadioGroupView struct {
	name           string // accessible name from a Form row when label is empty
	label          string
	options        []string
	value          string
	horizontal     bool
	disabled       bool
	optionDisabled map[string]bool
	onChange       func(string)
	size, textSize float32
	content        map[string]el.View
	itemSizes      map[string][2]float32
	tabStopEnabled *bool
	tabIndex       *int
	itemTabs       map[string]radioTab
}

type radioTab struct {
	stop  bool
	index int
}

// TabStop controls Tab entry into the group without disabling selection.
func (v *RadioGroupView) TabStop(on bool) *RadioGroupView { v.tabStopEnabled = &on; return v }

// TabIndex sets the group's position in a single el root's Tab order.
// Negative values skip Tab entry. Arrow navigation is unchanged.
func (v *RadioGroupView) TabIndex(index int) *RadioGroupView { v.tabIndex = &index; return v }

// ItemTab explicitly configures an individual option's Tab participation,
// overriding group settings and the default single-stop radio behavior.
// Mouse and arrow selection remain available. Unknown options are ignored.
func (v *RadioGroupView) ItemTab(value string, stop bool, index int) *RadioGroupView {
	if v.index(value) < 0 {
		return v
	}
	if v.itemTabs == nil {
		v.itemTabs = map[string]radioTab{}
	}
	v.itemTabs[value] = radioTab{stop, index}
	return v
}

// ClearItemTab restores group settings and the default single-stop behavior.
func (v *RadioGroupView) ClearItemTab(value string) *RadioGroupView {
	delete(v.itemTabs, value)
	return v
}

func RadioGroup(label string, options ...string) *RadioGroupView {
	v := &RadioGroupView{label: label}
	v.SetOptions(options...)
	return v
}
func (v *RadioGroupView) Horizontal() *RadioGroupView                    { v.horizontal = true; return v }
func (v *RadioGroupView) OnChange(fn func(value string)) *RadioGroupView { v.onChange = fn; return v }
func (v *RadioGroupView) Value() string                                  { return v.value }

// Size sets the indicator diameter in dp (12–64); zero restores 18dp.
func (v *RadioGroupView) Size(dp float32) *RadioGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.size = dp
		if dp > 0 {
			v.size = min(64, max(12, dp))
		}
	}
	return v
}

// TextSize sets inherited option text size in sp (8–128); zero restores inheritance.
func (v *RadioGroupView) TextSize(sp float32) *RadioGroupView {
	if sp >= 0 && finiteNumber(float64(sp)) {
		v.textSize = sp
		if sp > 0 {
			v.textSize = min(128, max(8, sp))
		}
	}
	return v
}

// ItemSize overrides an option's indicator diameter (dp) and inherited font
// size (sp). Each zero value inherits the group's setting. Unknown options or
// negative/non-finite sizes are ignored; positive sizes use the group limits.
func (v *RadioGroupView) ItemSize(value string, dp, sp float32) *RadioGroupView {
	if v.index(value) < 0 || dp < 0 || sp < 0 || !finiteNumber(float64(dp)) || !finiteNumber(float64(sp)) {
		return v
	}
	if dp == 0 && sp == 0 {
		delete(v.itemSizes, value)
		return v
	}
	if dp > 0 {
		dp = min(64, max(12, dp))
	}
	if sp > 0 {
		sp = min(128, max(8, sp))
	}
	if v.itemSizes == nil {
		v.itemSizes = map[string][2]float32{}
	}
	v.itemSizes[value] = [2]float32{dp, sp}
	return v
}

// Content replaces an option's visible label with display-only content.
// The option string remains its value and accessible name. Nil restores the
// text label; unknown options are ignored. Removed options discard their content.
func (v *RadioGroupView) Content(value string, content el.View) *RadioGroupView {
	if v.index(value) < 0 {
		return v
	}
	if content == nil {
		delete(v.content, value)
	} else {
		if v.content == nil {
			v.content = map[string]el.View{}
		}
		v.content[value] = content
	}
	return v
}

// SetValue chooses an option ("" clears) without calling OnChange.
func (v *RadioGroupView) SetValue(s string) {
	if s != "" && v.index(s) < 0 {
		s = ""
	}
	v.value = s
}
func (v *RadioGroupView) SetDisabled(on bool) { v.disabled = on }
func (v *RadioGroupView) SetOptions(options ...string) {
	v.options = nil
	seen := map[string]bool{}
	for _, o := range options {
		if o != "" && !seen[o] {
			v.options = append(v.options, o)
			seen[o] = true
		}
	}
	for o := range v.optionDisabled {
		if !seen[o] {
			delete(v.optionDisabled, o)
		}
	}
	for o := range v.content {
		if !seen[o] {
			delete(v.content, o)
		}
	}
	for o := range v.itemSizes {
		if !seen[o] {
			delete(v.itemSizes, o)
		}
	}
	for o := range v.itemTabs {
		if !seen[o] {
			delete(v.itemTabs, o)
		}
	}
	if v.index(v.value) < 0 {
		v.value = ""
	}
}

func (v *RadioGroupView) index(s string) int {
	for i, o := range v.options {
		if o == s {
			return i
		}
	}
	return -1
}

func (v *RadioGroupView) choose(s string) {
	if v.value == s || v.disabled || v.optionDisabled[s] || v.index(s) < 0 {
		return
	}
	v.value = s
	if v.onChange != nil {
		v.onChange(s)
	}
}

// Options returns a copy of the options, in keyboard navigation order.
func (v *RadioGroupView) Options() []string { return append([]string(nil), v.options...) }

// SetOptionDisabled prevents user selection while preserving a selected value.
func (v *RadioGroupView) SetOptionDisabled(value string, on bool) {
	if v.index(value) < 0 {
		return
	}
	if v.optionDisabled == nil {
		v.optionDisabled = map[string]bool{}
	}
	v.optionDisabled[value] = on
}
func (v *RadioGroupView) itemID(value string) string { return autoID("radio", v) + "/item/" + value }
func (v *RadioGroupView) tabStop() int {
	i := v.index(v.value)
	if i >= 0 && !v.optionDisabled[v.value] {
		return i
	}
	for i, o := range v.options {
		if !v.optionDisabled[o] {
			return i
		}
	}
	return -1
}
func (v *RadioGroupView) FocusID() string {
	if i := v.tabStop(); i >= 0 {
		return v.itemID(v.options[i])
	}
	return ""
}

// Item renders an option independently, sharing this group's selection and
// keyboard order. Render each option at most once per root/frame.
func (v *RadioGroupView) Item(value string) el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		if i := v.index(value); i >= 0 {
			return v.renderItem(cx, i)
		}
		return el.Div()
	})
}
func (v *RadioGroupView) renderItem(cx *el.Context, i int) el.Element {
	o := v.options[i]
	on := o == v.value
	disabled := v.disabled || v.optionDisabled[o]
	size, textSize := v.size, v.textSize
	if sizes := v.itemSizes[o]; sizes != [2]float32{} {
		if sizes[0] > 0 {
			size = sizes[0]
		}
		if sizes[1] > 0 {
			textSize = sizes[1]
		}
	}
	dot := radioDot(size, on, disabled)
	label := o
	if v.content[o] != nil {
		label = ""
	}
	row := check(v.itemID(o), "radio", label, o, on, disabled, dot, func() { v.choose(o); cx.Focus(v.itemID(o)) }).Focusable(i == v.tabStop()).
		OnKey(func(e el.KeyEvent) bool {
			if e.Modifiers != 0 {
				return false
			}
			direction, start := 0, i
			switch key.Name(e.Name) {
			case key.NameDownArrow, key.NameRightArrow:
				direction = 1
			case key.NameUpArrow, key.NameLeftArrow:
				direction = -1
			case key.NameHome:
				direction, start = 1, -1
			case key.NameEnd:
				direction, start = -1, 0
			default:
				return false
			}
			if e.State == el.KeyPress {
				for n := 1; n <= len(v.options); n++ {
					j := (start + direction*n + len(v.options)) % len(v.options)
					option := v.options[j]
					if !v.optionDisabled[option] && cx.Enabled(v.itemID(option)) {
						v.choose(option)
						cx.Focus(v.itemID(option))
						break
					}
				}
			}
			return true
		})
	if content := v.content[o]; content != nil {
		row.Child(el.Div().ID("label").Child(content.Render(cx)))
	}
	if textSize > 0 {
		row.TextSize(textSize)
	}

	if v.tabStopEnabled != nil {
		row.TabStop(*v.tabStopEnabled)
	}
	if v.tabIndex != nil {
		row.TabIndex(*v.tabIndex)
	}
	if tab, ok := v.itemTabs[o]; ok {
		row.Focusable(true).TabStop(tab.stop).TabIndex(tab.index)
	}
	return row
}
func (v *RadioGroupView) Render(cx *el.Context) el.Element {
	group := el.Div().Role("radiogroup").Name(v.a11y()).Gap(theme.SpaceMd).Disabled(v.disabled)
	if v.horizontal {
		group.Row().Wrap().Gap(theme.SpaceXl)
	}
	for i := range v.options {
		group.Child(v.renderItem(cx, i))
	}
	return labelled(v.label, group, "")
}

func (v *RadioGroupView) setName(s string) { v.name = s }
func (v *RadioGroupView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}

// radioDot draws a radio's ring, filled with a dot when on; size 0 is 18dp.
func radioDot(size float32, on, disabled bool) *el.DivEl {
	if size == 0 {
		size = 18
	}
	ring, fill := theme.Muted, theme.Surface
	if disabled {
		ring, fill = theme.Border, theme.Subtle
	} else if on {
		ring = theme.Primary
	}
	inner := el.Div().Size(el.Dp(size * 16 / 18)).Rounded(theme.RadiusFull).Bg(fill).Center()
	if on {
		inner.Child(el.Div().Size(el.Dp(size * 8 / 18)).Rounded(theme.RadiusFull).Bg(ring))
	}
	return el.Div().Size(el.Dp(size)).NoShrink().Rounded(theme.RadiusFull).Bg(ring).Center().Child(inner)
}
