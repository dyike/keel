package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// MessageScrollerView virtualizes a conversation with stable message keys.
// It follows the bottom while the user is there and preserves the visible
// message when history is prepended or an earlier message changes height.
type MessageScrollerView struct {
	list       *VariableListView
	atTop      bool
	onReachTop func()
}

// MessageScroller renders only the visible rows and nearby overscan. Keys are
// unique, nonempty IDs in oldest-first order. estimate is a positive height in
// dp until a message has been measured; its content can have any natural height.
func MessageScroller(keys []string, estimate float32, row func(*el.Context, int) el.Element) *MessageScrollerView {
	v := &MessageScrollerView{}
	v.list = VariableList(keys, estimate, func(cx *el.Context, i int) el.Element {
		box := el.Div().Px(20).Py(9).Items(el.Stretch)
		if i == 0 {
			box.Pt(theme.SpaceXl)
		}
		if i == v.list.Count()-1 {
			box.Pb(theme.SpaceXl)
		}
		if row != nil {
			box.Child(row(cx, i))
		}
		return box
	}).Fill()
	v.list.followEnd = true
	v.list.role = "log"
	return v
}
func (v *MessageScrollerView) OnReachTop(fn func()) *MessageScrollerView { v.onReachTop = fn; return v }

// SetKeys copies the new order, retaining measurements and the visible anchor.
// Call after inserting, appending, removing or reordering messages.
func (v *MessageScrollerView) SetKeys(keys []string) { v.list.SetKeys(keys) }
func (v *MessageScrollerView) ScrollToEnd()          { v.list.end++ }
func (v *MessageScrollerView) SetFollow(on bool)     { v.list.followEnd = on }
func (v *MessageScrollerView) SetDisabled(on bool)   { v.list.SetDisabled(on) }

// Invalidate refreshes estimated heights for changed offscreen messages.
// Visible streaming text and image growth are detected automatically.
func (v *MessageScrollerView) Invalidate(keys ...string) { v.list.Invalidate(keys...) }
func (v *MessageScrollerView) Render(cx *el.Context) el.Element {
	off, view, content := cx.ScrollState(v.list.ID())
	top := view > 0 && content > view && off < 1
	if top && !v.atTop && !v.list.disabled && cx.Enabled(v.list.ID()) && v.onReachTop != nil {
		v.atTop = true
		v.onReachTop()
	}
	// Small prepends and wheel inertia must not immediately request another batch.
	// Rearm after the reader has moved at least one viewport away from the top.
	if view > 0 && off >= view {
		v.atTop = false
	}
	wrap := el.Div().Grow().Items(el.Stretch).Child(v.list.Render(cx))
	if view > 0 && off+view < content-4 {
		wrap.Child(el.Div().Absolute().Right(20).Bottom(16).Child(Button(locale.Current().Latest, v.ScrollToEnd).Icon(IconChevronDown).Variant(ButtonSecondary).Size(28).Render(cx)))
	}
	return wrap.Disabled(v.list.disabled)
}
