package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"slices"
)

// MessageContentView mixes typed bubbles with arbitrary message content.
// Direct BubbleView items inherit message alignment and contribute Ghost metadata.
type MessageContentView struct {
	items    []el.View
	bubbles  map[*BubbleView]*BubbleView
	gap      float32
	disabled bool
	style    func(*el.DivEl)
}

// MessageContent copies items, ignoring nils. Reuse each item at most once.
// Custom views must provide stable identities to retain state when reordered.
func MessageContent(items ...el.View) *MessageContentView {
	v := &MessageContentView{gap: theme.SpaceMd}
	v.SetItems(items...)
	return v
}

// SetItems replaces the mixed sequence, retaining identities of surviving bubbles.
func (v *MessageContentView) SetItems(items ...el.View) {
	next := make(map[*BubbleView]*BubbleView)
	kept := make([]el.View, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if b, ok := item.(*BubbleView); ok {
			if b == nil {
				continue
			}
			copy := v.bubbles[b]
			if copy == nil {
				copy = &BubbleView{}
			}
			next[b] = copy
		}
		kept = append(kept, item)
	}
	v.items, v.bubbles = kept, next
}
func (v *MessageContentView) Items() []el.View { return slices.Clone(v.items) }
func (v *MessageContentView) Gap(dp float32) *MessageContentView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = dp
	}
	return v
}
func (v *MessageContentView) SetDisabled(on bool) { v.disabled = on }

// Style refines the content stack after defaults. Nil restores defaults.
// Do not retain the element or append children in this callback.
func (v *MessageContentView) Style(fn func(*el.DivEl)) *MessageContentView { v.style = fn; return v }
func (v *MessageContentView) hasGhost() bool {
	for b := range v.bubbles {
		if b.variant == BubbleGhost {
			return true
		}
	}
	return false
}

// Render lays out the stack independently with leading alignment. When installed
// directly in Message.Content, the message supplies alignment and default tone.
func (v *MessageContentView) Render(cx *el.Context) el.Element { return v.render(cx, false, false) }
func (v *MessageContentView) render(cx *el.Context, end, user bool) el.Element {
	box := el.Div().WFull().MinW(el.Dp(0)).MaxW(el.Full).Items(el.Stretch).Gap(v.gap)
	if end {
		box.Items(el.End)
	}
	if v.style != nil {
		v.style(box)
	}
	box.ID(autoID("message-content", v))
	if v.disabled {
		box.Disabled(true)
	}
	for _, item := range v.items {
		if source, ok := item.(*BubbleView); ok {
			copy := v.bubbles[source]
			*copy = *source
			copy.mine = end
			if copy.variant == BubbleAuto {
				copy.variant = BubbleSecondary
				if user {
					copy.variant = BubbleFilled
				}
			}
			box.Child(copy.Render(cx))
		} else {
			box.Child(item.Render(cx))
		}
	}
	return box
}
