package kit

import (
	"strconv"
	"time"

	"github.com/dyike/keel/ui/locale"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// NotificationTimeout is the default time a notice stays on screen.
const NotificationTimeout = 5 * time.Second

// MaxNotifications is how many notices show at each placement; later ones wait.
const MaxNotifications = 5

// Notice is one notification. Timeout 0 means NotificationTimeout; a negative
// Timeout keeps it until the user closes it.
type Notice struct {
	Title, Body string
	Tone        Tone
	Timeout     time.Duration
	Placement   NoticePlacement
	// Content replaces Body when non-nil; Title remains the accessible name.
	Content el.View
	// Action renders below the body. Call Dismiss explicitly to close on action.
	Action el.View
	// OnClick activates the card without dismissing it; child controls act independently.
	OnClick func()
	// OnClose runs once after removal, including timeout and programmatic Dismiss.
	OnClose func()
}

type notice struct {
	Notice
	id       int
	revision uint64
}

// NotifierView stacks notices by placement (top right by default). Render
// one as a direct child of the root view's top element, then call Notify from
// callbacks (or core.Update from other goroutines). Hovering or focusing a notice pauses
// its remaining timeout; notices never take focus or Esc.
type NotifierView struct {
	items     []notice
	next      int
	placement NoticePlacement
}

func Notifier() *NotifierView { return &NotifierView{} }

// Notify shows n and returns its id for Dismiss.
func (v *NotifierView) Notify(n Notice) int {
	v.next++
	v.items = append(v.items, notice{Notice: n, id: v.next})
	return v.next
}

// Dismiss removes a notice, shown or waiting.
func (v *NotifierView) Dismiss(id int) {
	for i, it := range v.items {
		if it.id == id {
			copy(v.items[i:], v.items[i+1:])
			v.items[len(v.items)-1] = notice{}
			v.items = v.items[:len(v.items)-1]
			if it.OnClose != nil {
				it.OnClose()
			}
			return
		}
	}
}

// Len reports how many notices are shown or waiting.
func (v *NotifierView) Len() int { return len(v.items) }

func (v *NotifierView) Render(cx *el.Context) el.Element {
	id := autoID("notifier", v)
	w, h := cx.ViewportSize()
	anchors := el.Div().ID(id).Absolute().Top(0).Left(0).Size(el.Dp(0))
	for position := NoticeTopRight; position <= NoticeRightCenter; position++ {
		stack := el.Div().W(el.Dp(320)).MaxW(el.Dp(max(0, w-32))).MaxH(el.Dp(max(0, h-32))).ScrollY().Gap(theme.SpaceMd).Items(el.Stretch)
		count := 0
		for _, it := range v.items {
			if v.position(it.Placement) != position {
				continue
			}
			if count == MaxNotifications {
				break
			}
			stack.Child(v.card(cx, id, it))
			count++
		}
		if count == 0 {
			continue
		}
		anchorID := id + "/position/" + strconv.Itoa(int(position))
		x, y, side, align := noticeAnchor(position, w, h)
		anchors.Child(el.Div().ID(anchorID).Absolute().Left(x).Top(y).Size(el.Dp(0)))
		cx.Overlay(anchorID, el.Anchored(anchorID, stack).Placement(side, align).Offset(0))
	}
	return anchors
}

func (v *NotifierView) card(cx *el.Context, base string, n notice) el.Element {
	id := base + "/" + strconv.Itoa(n.id)
	timeout := n.Timeout
	if timeout == 0 {
		timeout = NotificationTimeout
	}
	if timeout > 0 {
		cx.Countdown(id, noticeKey{base, n.id, n.revision}, timeout, cx.Hovered(id) || cx.FocusWithin(id), func() { v.Dismiss(n.id) })
	}
	text := el.Div().ID(id + "/text").Grow().MinW(el.Dp(0)).Gap(theme.SpaceXs).Child(el.Text(n.Title).Bold().TextColor(n.Tone.color()))
	body := el.Div().ID(id + "/body").W(el.Full).MinW(el.Dp(0))
	if n.Content != nil {
		body.Child(n.Content.Render(cx))
	} else if n.Body != "" {
		body.Child(el.Text(n.Body).TextSize(theme.TextMd).TextColor(theme.Muted))
	}
	text.Child(body.Hidden(n.Content == nil && n.Body == ""))
	action := el.Div().ID(id + "/action").W(el.Full).MinW(el.Dp(0)).Items(el.Start)
	if n.Action != nil {
		action.Child(n.Action.Render(cx))
	}
	text.Child(action.Hidden(n.Action == nil))
	card := floating(theme.ElevationMd).NoShrink().ID(id).Role("status").Name(n.Title).Value(n.Tone.name()).P(theme.SpaceLg).Row().Gap(10).Items(el.Start)
	if n.OnClick != nil {
		// A sibling hit surface lets interactive content consume clicks without
		// also activating the notification's background.
		card.Child(el.Div().ID(id + "/activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).
			Role("button").Name(n.Title).Focusable(true).CursorPointer().OnClick(func() {
			for _, current := range v.items {
				if current.id == n.id {
					if current.OnClick != nil {
						current.OnClick()
					}
					return
				}
			}
		}))
	}
	card.Child(
		el.Div().W(el.Dp(4)).H(el.Dp(20)).Rounded(theme.RadiusFull).Bg(n.Tone.color()),
		text,
		Button("", func() { v.Dismiss(n.id) }).Name(locale.Current().Name(locale.Current().Close, n.Title)).Icon(IconClose).Variant(ButtonGhost).Size(24).Render(cx),
	)
	return card
}

type noticeKey struct {
	base     string
	id       int
	revision uint64
}

// Update replaces a shown or queued notice in place and restarts its timeout.
// Like Notify, call it on the UI frame lock (core.Update for background work).
func (v *NotifierView) Update(id int, n Notice) bool {
	for i := range v.items {
		if v.items[i].id == id {
			v.items[i].Notice = n
			v.items[i].revision++
			return true
		}
	}
	return false
}
