package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// BubbleView is a chat bubble. Mine marks the current user's: it sits on the
// right in the primary color; others sit on the left on a subtle background.
type BubbleView struct {
	content el.View
	mine    bool
}

func Bubble(content el.View) *BubbleView { return &BubbleView{content: content} }
func (v *BubbleView) Mine() *BubbleView  { v.mine = true; return v }

func (v *BubbleView) Render(cx *el.Context) el.Element {
	box := el.Div().MaxW(el.Frac(0.75)).Rounded(12).Px(14).Py(10)
	if v.mine {
		box.Bg(theme.Primary).TextColor(theme.OnColor)
	} else {
		box.Bg(theme.Subtle)
	}
	if v.content != nil {
		box.Child(v.content.Render(cx))
	}
	row := el.Div().Row().Child(box)
	if v.mine {
		row.Justify(el.End)
	}
	return row
}

// MessageState describes delivery of a message.
type MessageState uint8

const (
	MessageReady MessageState = iota
	MessageSending
	MessageFailed
)

// MessageReaction is an application-named reaction and its current count.
type MessageReaction struct {
	Name   string
	Count  int
	Active bool
}

// MessageView is one turn of a conversation: an avatar, the content and a row
// of actions (copy, retry…) under it. User messages show as a Bubble on the
// right without an avatar; others show full width beside the author's avatar,
// which suits long Markdown answers.
type MessageView struct {
	author     string
	content    el.View
	user       bool
	actions    []el.View
	state      MessageState
	failure    string
	disabled   bool
	retry      func()
	reactions  []MessageReaction
	onReaction func(int, bool)
}

// Message creates a message from author; its avatar shows author's initials.
func Message(author string, content el.View) *MessageView {
	return &MessageView{author: author, content: content}
}

// User marks the current user's message.
func (v *MessageView) User() *MessageView { v.user = true; return v }

// Actions sets the views under the content and copies the slice.
func (v *MessageView) Actions(views ...el.View) *MessageView {
	v.actions = slices.Clone(views)
	return v
}

// SetState updates delivery without replacing the content or action views.
func (v *MessageView) SetState(state MessageState, failure string) {
	if state > MessageFailed {
		state = MessageReady
	}
	v.state, v.failure = state, failure
}
func (v *MessageView) SetDisabled(on bool)            { v.disabled = on }
func (v *MessageView) OnRetry(fn func()) *MessageView { v.retry = fn; return v }
func (v *MessageView) Reactions(items ...MessageReaction) *MessageView {
	v.reactions = slices.Clone(items)
	for i := range v.reactions {
		v.reactions[i].Count = max(0, v.reactions[i].Count)
	}
	return v
}
func (v *MessageView) OnReaction(fn func(int, bool)) *MessageView { v.onReaction = fn; return v }

func (v *MessageView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	state := ""
	body := el.Div().Gap(8).Items(el.Stretch)
	if v.user {
		body.Child(Bubble(v.content).Mine().Render(cx))
	} else if v.content != nil {
		body.Child(v.content.Render(cx))
	}
	if v.state == MessageSending {
		state = "sending"
		status := el.Div().Row().Child(el.Text(text.Sending).TextSize(12).TextColor(theme.Muted))
		if v.user {
			status.Justify(el.End)
		}
		body.Child(status)
	} else if v.state == MessageFailed {
		state = "failed"
		detail := text.SendFailed
		if v.failure != "" {
			detail += ": " + v.failure
		}
		status := el.Div().Row().Wrap().Gap(8).Items(el.Center).Child(el.Text(detail).TextSize(12).TextColor(theme.DangerText))
		if v.user {
			status.Justify(el.End)
		}
		if v.retry != nil {
			status.Child(Button(text.Retry, func() {
				if v.state != MessageFailed || v.disabled {
					return
				}
				v.SetState(MessageSending, "")
				v.retry()
			}).Variant(ButtonGhost).Size(28).Render(cx))
		}
		body.Child(status)
	}
	if len(v.actions) > 0 {
		row := el.Div().Row().Wrap().Gap(4)
		if v.user {
			row.Justify(el.End)
		}
		for _, a := range v.actions {
			if a != nil {
				row.Child(a.Render(cx))
			}
		}
		body.Child(row)
	}
	if len(v.reactions) > 0 {
		row := el.Div().Row().Wrap().Gap(4)
		if v.user {
			row.Justify(el.End)
		}
		for i, r := range v.reactions {
			label := r.Name + " " + strconv.Itoa(r.Count)
			row.Child(toggleButton(cx, autoID("message", v)+"/reaction/"+strconv.Itoa(i), label, nil, r.Active, v.onReaction == nil, func() {
				if v.disabled || i >= len(v.reactions) || v.onReaction == nil {
					return
				}
				r := &v.reactions[i]
				r.Active = !r.Active
				if r.Active {
					r.Count++
				} else {
					r.Count = max(0, r.Count-1)
				}
				v.onReaction(i, r.Active)
			}))
		}
		body.Child(row)
	}
	article := el.Div().Role("article").Name(v.author).Value(state).Disabled(v.disabled).Items(el.Stretch)
	if v.user {
		return article.Child(body)
	}
	body.Grow().W(el.Dp(0)).Pt(4)
	return article.Row().Gap(10).Items(el.Start).Child(Avatar(v.author).Size(28).Render(cx), body)
}
