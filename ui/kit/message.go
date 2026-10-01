package kit

import (
	"github.com/dyike/keel/ui/el"
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

// MessageView is one turn of a conversation: an avatar, the content and a row
// of actions (copy, retry…) under it. User messages show as a Bubble on the
// right without an avatar; others show full width beside the author's avatar,
// which suits long Markdown answers.
type MessageView struct {
	author  string
	content el.View
	user    bool
	actions []el.View
}

// Message creates a message from author; its avatar shows author's initials.
func Message(author string, content el.View) *MessageView {
	return &MessageView{author: author, content: content}
}

// User marks the current user's message.
func (v *MessageView) User() *MessageView { v.user = true; return v }

// Actions sets the views under the content, shown once it is complete.
func (v *MessageView) Actions(views ...el.View) *MessageView { v.actions = views; return v }

func (v *MessageView) Render(cx *el.Context) el.Element {
	if v.user {
		return el.Div().Role("article").Name(v.author).Items(el.Stretch).Child(Bubble(v.content).Mine().Render(cx))
	}
	body := el.Div().Grow().W(el.Dp(0)).Pt(4).Gap(8).Items(el.Stretch)
	if v.content != nil {
		body.Child(v.content.Render(cx))
	}
	if len(v.actions) > 0 {
		row := el.Div().Row().Gap(4)
		for _, a := range v.actions {
			row.Child(a.Render(cx))
		}
		body.Child(row)
	}
	return el.Div().Role("article").Name(v.author).Row().Gap(10).Items(el.Start).
		Child(Avatar(v.author).Size(28).Render(cx), body)
}
