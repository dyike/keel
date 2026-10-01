package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// MessageScrollerView scrolls a conversation. It follows new content at the
// bottom while the view is at the end (a streaming answer stays in sight),
// stays put when the user has scrolled up, and then offers a 回到最新 button.
// Scrolling to the top calls OnReachTop, where older messages can be loaded;
// call HistoryPrepended after inserting them so the view does not jump.
type MessageScrollerView struct {
	items      func(cx *el.Context) []el.Element
	gap        float32
	end, keep  int
	atTop      bool
	noFollow   bool
	onReachTop func()
}

// MessageScroller shows the elements items returns, oldest first.
func MessageScroller(items func(cx *el.Context) []el.Element) *MessageScrollerView {
	return &MessageScrollerView{items: items, gap: 18}
}
func (v *MessageScrollerView) OnReachTop(fn func()) *MessageScrollerView { v.onReachTop = fn; return v }

// ScrollToEnd jumps to the newest message on the next frame, e.g. after the
// user sends one while scrolled up.
func (v *MessageScrollerView) ScrollToEnd() { v.end++ }

// SetFollow turns following new content at the bottom on (the default) or
// off, e.g. off while showing a finished document from its start.
func (v *MessageScrollerView) SetFollow(on bool) { v.noFollow = !on }

// HistoryPrepended keeps the visible messages in place after older ones were
// inserted at the start.
func (v *MessageScrollerView) HistoryPrepended() { v.keep++ }

func (v *MessageScrollerView) Render(cx *el.Context) el.Element {
	id := autoID("messages", v)
	off, view, content := cx.ScrollState(id)
	top := view > 0 && content > view && off < 1
	if top && !v.atTop && v.onReachTop != nil {
		v.onReachTop()
	}
	v.atTop = top
	list := el.Div().ID(id).Role("log").Grow().ScrollY().When(!v.noFollow, func(d *el.DivEl) { d.StickToBottom() }).
		ScrollToEndOn(v.end).KeepBottomOn(v.keep).
		Px(20).Py(16).Gap(v.gap).Items(el.Stretch)
	if v.items != nil {
		list.Children(v.items(cx))
	}
	wrap := el.Div().Grow().Items(el.Stretch).Child(list)
	if view > 0 && off+view < content-4 {
		wrap.Child(el.Div().Absolute().Right(20).Bottom(16).Child(
			Button(locale.Current().Latest, v.ScrollToEnd).Icon(IconChevronDown).Variant(ButtonSecondary).Size(28).Render(cx)))
	}
	return wrap
}
