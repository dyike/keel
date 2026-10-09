package core

import "gioui.org/io/event"

// EditAction is a standard application-menu operation. Custom editors may
// consume these with NextEditAction while they retain keyboard focus.
type EditAction string

const (
	EditCopy      EditAction = "copy"
	EditCut       EditAction = "cut"
	EditPaste     EditAction = "paste"
	EditSelectAll EditAction = "select-all"
	EditUndo      EditAction = "undo"
	EditRedo      EditAction = "redo"
)

// NextEditAction returns a queued menu action only for the focused, enabled
// editor. Call during Layout; menu clicks preserve the editor's focus.
func NextEditAction(gtx C, tag event.Tag) (EditAction, bool) {
	if !gtx.Enabled() || !gtx.Focused(tag) {
		return "", false
	}
	if host, ok := CurrentWindow().(interface{ TakeEditAction() (EditAction, bool) }); ok {
		return host.TakeEditAction()
	}
	return "", false
}
