package kit

import (
	"strconv"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// NotificationTimeout is the default time a notice stays on screen.
const NotificationTimeout = 5 * time.Second

// MaxNotifications is how many notices show at once; later ones wait.
const MaxNotifications = 5

// Notice is one notification. Timeout 0 means NotificationTimeout; a negative
// Timeout keeps it until the user closes it.
type Notice struct {
	Title, Body string
	Tone        Tone
	Timeout     time.Duration
}

type notice struct {
	Notice
	id int
}

// NotifierView stacks notices in the top right corner of the window. Render
// one as a direct child of the root view's top element, then call Notify from
// callbacks (or core.Update from other goroutines). Hovering a notice restarts
// its timeout; notices never take focus or Esc.
type NotifierView struct {
	items []notice
	next  int
}

func Notifier() *NotifierView { return &NotifierView{} }

// Notify shows n and returns its id for Dismiss.
func (v *NotifierView) Notify(n Notice) int {
	v.next++
	v.items = append(v.items, notice{n, v.next})
	return v.next
}

// Dismiss removes a notice, shown or waiting.
func (v *NotifierView) Dismiss(id int) {
	for i, it := range v.items {
		if it.id == id {
			v.items = append(v.items[:i], v.items[i+1:]...)
			return
		}
	}
}

// Len reports how many notices are shown or waiting.
func (v *NotifierView) Len() int { return len(v.items) }

func (v *NotifierView) Render(cx *el.Context) el.Element {
	id := autoID("notifier", v)
	if len(v.items) > 0 {
		stack := el.Div().W(el.Dp(320)).Gap(8).Items(el.Stretch)
		for i, it := range v.items {
			if i == MaxNotifications {
				break
			}
			stack.Child(v.card(cx, id, it))
		}
		cx.Overlay(id, el.Anchored(id, stack).Placement(el.Bottom, el.End).Offset(0))
	}
	return el.Div().ID(id).Absolute().Top(16).Right(16).Size(el.Dp(0))
}

func (v *NotifierView) card(cx *el.Context, base string, n notice) el.Element {
	id := base + "/" + strconv.Itoa(n.id)
	timeout := n.Timeout
	if timeout == 0 {
		timeout = NotificationTimeout
	}
	if timeout > 0 && !cx.Hovered(id) {
		cx.After(noticeKey{base, n.id}, timeout, func() { v.Dismiss(n.id) })
	}
	text := el.Div().Grow().Gap(4).Child(el.Text(n.Title).Bold().TextColor(n.Tone.color()))
	if n.Body != "" {
		text.Child(el.Text(n.Body).TextSize(13).TextColor(theme.Muted))
	}
	return surface().ID(id).Role("status").Name(n.Title).Value(n.Tone.name()).P(12).Row().Gap(10).Items(el.Start).Child(
		el.Div().W(el.Dp(4)).H(el.Dp(20)).Rounded(2).Bg(n.Tone.color()),
		text,
		Button("", func() { v.Dismiss(n.id) }).Name("关闭 "+n.Title).Icon(IconClose).Variant(ButtonGhost).Size(24).Render(cx),
	)
}

type noticeKey struct {
	base string
	id   int
}
