package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

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
	author                   string
	bubble                   *BubbleView
	surface                  *BubbleView
	headerInset, footerInset *bool
	content                  el.View
	avatar, header, footer   el.View
	avatarSet                bool
	user                     bool
	actions                  []el.View
	state                    MessageState
	failure                  string
	disabled                 bool
	retry                    func()
	reactions                []MessageReaction
	onReaction               func(int, bool)
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
	id := autoID("message", v)
	body := el.Div().ID(id + "/body").Gap(theme.SpaceMd).Items(el.Stretch)
	if v.header != nil {
		body.Child(v.metadata(cx, id+"/header", v.header, v.headerInset))
	}
	content := el.Div().ID(id + "/content").Items(el.Stretch)
	if v.surface != nil {
		if v.bubble == nil {
			v.bubble = &BubbleView{}
		}
		*v.bubble = *v.surface
		v.bubble.mine = v.user
		content.Child(v.bubble.Render(cx))
	} else if v.user {
		if v.bubble == nil {
			v.bubble = Bubble(v.content).Mine()
		}
		*v.bubble = BubbleView{content: v.content, mine: true, reactionAlign: el.End}
		content.Child(v.bubble.Render(cx))
	} else if v.content != nil {
		content.Child(v.content.Render(cx))
	}
	body.Child(content.Hidden(v.content == nil))
	if v.state == MessageSending {
		state = "sending"
		status := el.Div().ID(id + "/status").Row().Child(el.Text(text.Sending).TextSize(theme.TextSm).TextColor(theme.Muted))
		if v.user {
			status.Justify(el.End)
		}
		body.Child(status)
	} else if v.state == MessageFailed {
		state = "failed"
		detail := text.SendFailed
		if v.failure != "" {
			detail = text.Detail(detail, v.failure)
		}
		status := el.Div().ID(id + "/status").Row().Wrap().Gap(theme.SpaceMd).Items(el.Center).Child(el.Text(detail).TextSize(theme.TextSm).TextColor(theme.DangerText))
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
		row := el.Div().ID(id + "/actions").Row().Wrap().Gap(theme.SpaceXs)
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
		row := el.Div().ID(id + "/reactions").Row().Wrap().Gap(theme.SpaceXs)
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
	if v.footer != nil {
		body.Child(v.metadata(cx, id+"/footer", v.footer, v.footerInset))
	}
	article := el.Div().ID(id).W(el.Full).MinW(el.Dp(0)).Role("article").Name(v.author).Value(state).Disabled(v.disabled).Items(el.Stretch)
	avatar := v.avatar
	if !v.avatarSet && !v.user {
		avatar = Avatar(v.author).Size(28)
	}
	if avatar == nil {
		return article.Child(body)
	}
	identity := el.Div().ID(id + "/avatar").NoShrink().Child(avatar.Render(cx))
	body.Grow().W(el.Dp(0)).Pt(theme.SpaceXs)
	article.Row().Gap(10).Items(el.Start)
	if v.user {
		return article.Child(body, identity)
	}
	return article.Child(identity, body)
}
