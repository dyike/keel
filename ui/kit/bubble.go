package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// BubbleVariant selects a semantic surface independently of alignment.
type BubbleVariant uint8

const (
	BubbleAuto BubbleVariant = iota // primary for Mine, secondary otherwise
	BubbleFilled
	BubbleSecondary
	BubbleMuted
	BubbleTinted
	BubbleOutline
	BubbleGhost
	BubbleDestructive
)

type BubblePart uint8

const (
	BubblePartRoot BubblePart = iota
	BubblePartContent
	BubblePartReactions
)

type BubbleReactionSide uint8

const (
	BubbleReactionBottom BubbleReactionSide = iota
	BubbleReactionTop
)

// BubbleView lays out a chat surface and an optional application-owned reaction region.
type BubbleView struct {
	content       el.View
	reactions     el.View
	mine          bool
	variant       BubbleVariant
	reactionSide  BubbleReactionSide
	reactionAlign el.Align
	styles        [3]func(*el.DivEl)
}

func Bubble(content el.View) *BubbleView { return &BubbleView{content: content, reactionAlign: el.End} }
func (v *BubbleView) Mine() *BubbleView  { v.mine = true; return v }

// Alignment changes placement without replacing content or reaction state.
func (v *BubbleView) Alignment(align el.Align) *BubbleView {
	if align == el.Start || align == el.End {
		v.mine = align == el.End
	}
	return v
}
func (v *BubbleView) Variant(variant BubbleVariant) *BubbleView {
	if variant <= BubbleDestructive {
		v.variant = variant
	}
	return v
}
func (v *BubbleView) Content(content el.View) *BubbleView { v.content = content; return v }

// Reactions supplies arbitrary controls or content; nil clears the region.
// Counts, selection and callbacks belong to the supplied view.
func (v *BubbleView) Reactions(view el.View) *BubbleView { v.reactions = view; return v }
func (v *BubbleView) ReactionSide(side BubbleReactionSide) *BubbleView {
	if side <= BubbleReactionTop {
		v.reactionSide = side
	}
	return v
}
func (v *BubbleView) ReactionAlignment(align el.Align) *BubbleView {
	if align == el.Start || align == el.End {
		v.reactionAlign = align
	}
	return v
}

// PartStyle refines fresh elements after defaults; nil restores the defaults.
func (v *BubbleView) PartStyle(part BubblePart, fn func(*el.DivEl)) *BubbleView {
	if int(part) < len(v.styles) {
		v.styles[part] = fn
	}
	return v
}
func (v *BubbleView) part(part BubblePart, id string, box *el.DivEl) *el.DivEl {
	if fn := v.styles[part]; fn != nil {
		fn(box)
	}
	return box.ID(id)
}
func (v *BubbleView) Render(cx *el.Context) el.Element {
	id := autoID("bubble", v)
	variant := v.variant
	if variant == BubbleAuto {
		variant = BubbleSecondary
		if v.mine {
			variant = BubbleFilled
		}
	}
	box := el.Div().Rounded(theme.RadiusXl).Px(theme.SpaceLg).Py(theme.SpaceMd).TextColor(theme.Text)
	switch variant {
	case BubbleFilled:
		box.Bg(theme.Primary).BgGradient(theme.PrimaryGradient).TextColor(theme.OnColor)
	case BubbleSecondary:
		box.Bg(theme.Subtle)
	case BubbleMuted:
		box.Bg(theme.Subtle).TextColor(theme.Muted)
	case BubbleTinted:
		box.Bg(theme.Highlight).TextColor(theme.PrimaryText)
	case BubbleOutline:
		box.Border(1, theme.Border)
	case BubbleGhost:
		box.P(0).Rounded(0).WFull()
	case BubbleDestructive:
		c := theme.DangerText
		c.A = 24
		box.Bg(c).TextColor(theme.DangerText)
	}
	box = v.part(BubblePartContent, id+"/content", box)
	if v.content != nil {
		box.Child(v.content.Render(cx))
	}
	reaction := el.Div().Bg(theme.Surface).TextColor(theme.Text).Border(1, theme.Border).Rounded(theme.RadiusFull).Px(theme.SpaceSm).Py(theme.SpaceXs)
	reaction = v.part(BubblePartReactions, id+"/reactions", reaction)
	if v.reactions != nil {
		reaction.Child(v.reactions.Render(cx))
	} else {
		reaction.Hidden(true)
	}
	reactionRow := el.Div().ID(id + "/reaction-row").Row().Justify(v.reactionAlign).Child(reaction)
	if v.reactions == nil {
		reactionRow.Hidden(true)
	}
	stack := el.Div().ID(id + "/stack").MaxW(el.Frac(.75)).Items(el.Stretch)
	if variant == BubbleGhost {
		stack.MaxW(el.Full).WFull()
	}
	if v.reactionSide == BubbleReactionTop {
		stack.Child(reactionRow, box)
	} else {
		stack.Child(box, reactionRow)
	}
	row := el.Div().WFull().Row().Child(stack)
	if v.mine {
		row.Justify(el.End)
	}
	return v.part(BubblePartRoot, id, row)
}
