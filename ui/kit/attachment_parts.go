package kit

import (
	"github.com/dyike/keel/ui/el"
	"slices"
)

// AttachmentPart identifies a styleable attachment region.
type AttachmentPart uint8

const (
	AttachmentPartRoot AttachmentPart = iota
	AttachmentPartMedia
	AttachmentPartContent
	AttachmentPartTitle
	AttachmentPartDescription
	AttachmentPartActions
)

// Content replaces the default metadata and progress bar. Nil restores them.
// With OnOpen configured, use display content here and interactive controls in Actions.
func (v *AttachmentView) Content(view el.View) *AttachmentView { v.content = view; return v }

// Actions adds custom controls before the built-in cancel/retry/remove controls.
// The slice is copied; nil entries are ignored. Empty arguments clear custom controls.
func (v *AttachmentView) Actions(views ...el.View) *AttachmentView {
	v.actions = slices.Clone(views)
	return v
}

// PartStyle refines a fresh region after defaults, each frame. Nil clears the
// refinement. Do not retain the element or use this callback to append children.
// Identity and root semantics are owned by the component.
func (v *AttachmentView) PartStyle(part AttachmentPart, fn func(*el.DivEl)) *AttachmentView {
	if int(part) < len(v.styles) {
		v.styles[part] = fn
	}
	return v
}
func (v *AttachmentView) part(part AttachmentPart, id string, box *el.DivEl) *el.DivEl {
	if fn := v.styles[part]; fn != nil {
		fn(box)
	}
	return box.ID(id)
}
