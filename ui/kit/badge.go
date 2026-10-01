package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// BadgeView is a count or dot, alone or on a child's top right corner. A
// count of 0 or less hides it. On a child it never changes the child's
// layout: it is drawn over the corner.
type BadgeView struct {
	count, max int
	dot        bool
	tone       Tone
	child      el.View
}

func Badge(count int) *BadgeView                    { return &BadgeView{count: count, max: 99, tone: ToneDanger} }
func (v *BadgeView) Dot() *BadgeView                { v.dot = true; return v }
func (v *BadgeView) Tone(t Tone) *BadgeView         { v.tone = t; return v }
func (v *BadgeView) Child(child el.View) *BadgeView { v.child = child; return v }

// Max caps the shown number: above it the badge reads "99+" (default 99).
func (v *BadgeView) Max(n int) *BadgeView {
	if n > 0 {
		v.max = n
	}
	return v
}
func (v *BadgeView) Value() int     { return v.count }
func (v *BadgeView) SetValue(n int) { v.count = n }

func (v *BadgeView) badge() el.Element {
	label := strconv.Itoa(v.count)
	if v.count > v.max {
		label = strconv.Itoa(v.max) + "+"
	}
	value := label
	b := el.Div().Role("badge").Name(strconv.Itoa(v.count)).Bg(v.tone.solid()).TextColor(theme.OnColor)
	if v.dot {
		value = "dot"
		b.Size(el.Dp(8)).Rounded(4)
	} else {
		b.H(el.Dp(18)).MinW(el.Dp(18)).Px(5).Rounded(9).Center().TextSize(11).Child(el.Text(label).Bold())
	}
	return b.Value(value)
}

func (v *BadgeView) Render(cx *el.Context) el.Element {
	if v.child == nil {
		if v.count <= 0 {
			return el.Div().Hidden(true)
		}
		return el.Div().Items(el.Start).Child(v.badge())
	}
	box := el.Div().Items(el.Start).Child(v.child.Render(cx))
	if v.count > 0 {
		off := float32(-7)
		if v.dot {
			off = -3
		}
		box.Child(el.Div().Absolute().Top(off).Right(off).Child(v.badge()))
	}
	return box
}
