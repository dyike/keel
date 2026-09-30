package widget

import (
	"gioui.org/io/semantic"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// area lays out w inside its own clip area and attaches semantic ops to it, so
// the component shows up as one node in Gio's semantic tree. Automation
// (ui/window's snapshot) and accessibility read that tree.
func area(gtx C, w func(gtx C) D, ops ...interface{ Add(*op.Ops) }) D {
	m := op.Record(gtx.Ops)
	d := w(gtx)
	call := m.Stop()
	defer clip.Rect{Max: d.Size}.Push(gtx.Ops).Pop()
	for _, o := range ops {
		o.Add(gtx.Ops)
	}
	call.Add(gtx.Ops)
	return d
}

// linkDesc marks a clickable node as a link rather than a button.
const linkDesc = semantic.DescriptionOp("link")
