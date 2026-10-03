package kit

import "github.com/dyike/keel/ui/el"

// MessagePart identifies a styleable region of a complete message.
type MessagePart uint8

const (
	MessagePartRoot MessagePart = iota
	MessagePartStack
	MessagePartAvatar
	MessagePartHeader
	MessagePartContent
	MessagePartFooter
	MessagePartStatus
	MessagePartActions
	MessagePartReactions
)

// PartStyle refines a fresh region after defaults, each frame. Nil restores
// defaults; invalid parts are ignored. Do not retain the element or append
// children here. The component owns identity, root semantics and disabled state.
func (v *MessageView) PartStyle(part MessagePart, fn func(*el.DivEl)) *MessageView {
	if int(part) < len(v.styles) {
		v.styles[part] = fn
	}
	return v
}

func (v *MessageView) part(part MessagePart, id string, box *el.DivEl) *el.DivEl {
	if fn := v.styles[part]; fn != nil {
		fn(box)
	}
	return box.ID(id)
}
