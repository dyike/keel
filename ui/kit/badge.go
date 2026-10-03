package kit

import (
	"image/color"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// BadgeView is a count or dot, alone or on a child's top right corner. A
// count of 0 or less hides count/dot badges. On a child it never changes the child's
// layout: it is drawn over the corner.
type BadgeView struct {
	count, max int
	dot        bool
	tone       Tone
	child      el.View
	icon       IconName
	size       float32
	color      *color.NRGBA
	name       string
}

func Badge(count int) *BadgeView                    { return &BadgeView{count: count, max: 99, tone: ToneDanger} }
func (v *BadgeView) Dot() *BadgeView                { v.dot = true; v.icon = IconNone; return v }
func (v *BadgeView) Tone(t Tone) *BadgeView         { v.tone = t; v.color = nil; return v }
func (v *BadgeView) Child(child el.View) *BadgeView { v.child = child; return v }

// Icon replaces the count with a status icon, regardless of count. IconNone
// restores count mode. Icon badges sit at the child's bottom right.
func (v *BadgeView) Icon(name IconName) *BadgeView { v.icon = name; v.dot = false; return v }

// Name sets an accessible status name; empty restores the raw count label.
func (v *BadgeView) Name(name string) *BadgeView { v.name = name; return v }

// Size sets count/icon height in dp, 18 by default. Dots scale proportionally.
func (v *BadgeView) Size(dp float32) *BadgeView {
	if dp >= 12 && dp <= 128 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}

// Color overrides the background. Tone restores theme-derived colors.
func (v *BadgeView) Color(c color.NRGBA) *BadgeView { v.color = &c; return v }
func (v *BadgeView) diameter() float32 {
	if v.size > 0 {
		return v.size
	}
	return 18
}
func (v *BadgeView) visible() bool { return v.count > 0 || v.icon != IconNone }

// Max caps the shown number: above it the badge reads "99+" (default 99).
func (v *BadgeView) Max(n int) *BadgeView {
	if n > 0 {
		v.max = n
	}
	return v
}
func (v *BadgeView) Value() int     { return v.count }
func (v *BadgeView) SetValue(n int) { v.count = n }

func (v *BadgeView) badge(cx *el.Context) el.Element {
	label := strconv.Itoa(v.count)
	if v.count > v.max {
		label = strconv.Itoa(v.max) + "+"
	}
	value := label
	bg := v.tone.solid()
	if v.color != nil {
		bg = *v.color
	}
	fg := contrastingText(bg)
	name := v.name
	if name == "" {
		name = strconv.Itoa(v.count)
	}
	size := v.diameter()
	b := el.Div().Role("badge").Name(name).Bg(bg).TextColor(fg)
	if v.dot {
		value = "dot"
		b.Size(el.Dp(size * 8 / 18)).Rounded(theme.RadiusFull)
	} else if v.icon != IconNone {
		value = "icon"
		b.Size(el.Dp(size)).Rounded(theme.RadiusFull).Border(1, theme.Surface).Center().Child(Icon(v.icon).Size(size * 2 / 3).Color(fg).Render(cx))
	} else {
		b.H(el.Dp(size)).MinW(el.Dp(size)).Px(size * 5 / 18).Rounded(theme.RadiusFull).Center().TextSize(size * theme.TextXs / 18).Child(el.Text(label).Bold())
	}
	return b.Value(value)
}

func (v *BadgeView) Render(cx *el.Context) el.Element {
	if v.child == nil {
		if !v.visible() {
			return el.Div().Hidden(true)
		}
		return el.Div().Items(el.Start).Child(v.badge(cx))
	}
	box := el.Div().Items(el.Start).Child(v.child.Render(cx))
	if v.visible() {
		off := -v.diameter() * 7 / 18
		if v.dot {
			off = -v.diameter() / 6
		}
		position := el.Div().Absolute().Right(off)
		if v.icon != IconNone {
			position.Bottom(off)
		} else {
			position.Top(off)
		}
		box.Child(position.Child(v.badge(cx)))
	}
	return box
}
