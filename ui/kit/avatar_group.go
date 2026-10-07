package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// AvatarGroupView displays overlapping avatars and an overflow count.
type AvatarGroupView struct {
	avatars  []*AvatarView
	size     float32
	limit    int
	ellipsis bool
}

func AvatarGroup(avatars ...*AvatarView) *AvatarGroupView {
	v := &AvatarGroupView{size: 40, limit: -1}
	v.SetAvatars(avatars...)
	return v
}

// SetAvatars copies the slice, ignoring nil entries. Avatar instances remain shared.
func (v *AvatarGroupView) SetAvatars(avatars ...*AvatarView) {
	v.avatars = nil
	for _, a := range avatars {
		if a != nil {
			v.avatars = append(v.avatars, a)
		}
	}
}
func (v *AvatarGroupView) Avatars() []*AvatarView { return slices.Clone(v.avatars) }

// Limit caps visible avatars. Negative means unlimited; zero shows only overflow.
func (v *AvatarGroupView) Limit(n int) *AvatarGroupView { v.limit = n; return v }

// Ellipsis replaces the visible +N count with …; semantics retain the count.
func (v *AvatarGroupView) Ellipsis(on bool) *AvatarGroupView { v.ellipsis = on; return v }
func (v *AvatarGroupView) Size(dp float32) *AvatarGroupView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = max(16, min(256, dp))
	}
	return v
}
func (v *AvatarGroupView) Render(cx *el.Context) el.Element {
	n := len(v.avatars)
	if v.limit >= 0 {
		n = min(n, v.limit)
	}
	hidden := len(v.avatars) - n
	if len(v.avatars) == 0 {
		return el.Div().Role("group")
	}
	count := n
	if hidden > 0 {
		count++
	}
	size := v.size
	if size == 0 {
		size = 40
	}
	stride := size * .75
	width := float32(0)
	if count > 0 {
		width = size + float32(count-1)*stride
	}
	row := el.Div().W(el.Dp(width)).H(el.Dp(size))
	for i := 0; i < count; i++ {
		var content el.Element
		if i < n {
			a := *v.avatars[i]
			a.size = size
			content = a.Render(cx)
		} else {
			number := strconv.Itoa(hidden)
			label := "+" + number
			if v.ellipsis {
				label = "…"
			}
			content = el.Div().Size(el.Dp(size)).Rounded(theme.RadiusFull).Bg(theme.Subtle).Center().
				Role("avatar").Name(locale.Current().Name(locale.Current().More, number)).Value(number).
				Child(el.Text(label).TextSize(size * .32).TextColor(theme.Text).MaxLines(1))
		}
		row.Child(el.Div().Absolute().Left(float32(i)*stride).Top(0).Size(el.Dp(size)).Rounded(theme.RadiusFull).
			Child(content, el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Rounded(theme.RadiusFull).Border(2, theme.Surface)))
	}
	return el.Div().Role("group").W(el.Dp(width)).MaxW(el.Full).ScrollX().Pb(scrollbarGutter).Child(row)
}
