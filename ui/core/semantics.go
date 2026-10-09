package core

import (
	"gioui.org/io/semantic"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Semantic lays out w inside its own clip area and attaches semantic ops to
// it, so the component is one node in Gio's semantic tree with its real
// bounds. Agents (ui/window automation) and accessibility read that tree.
//
// Gio's classes cover buttons, checkboxes, editors, radios and switches. Other
// roles go in a Role description, e.g. Role("row") or Role("select", value).
func Semantic(gtx C, w func(gtx C) D, ops ...interface{ Add(*op.Ops) }) D {
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

// Role marks a node's role for agents, optionally with a value.
// The internal el-inert marker hides a background subtree from Agent snapshots
// while a modal el layer is active; it is not exposed as a component role. Automation
// reads it as "role" or "role:value" from the node's description. A button may
// carry "button:loading" while its action is unavailable. On a semantic.Button,
// automation keeps only link, tab, columnheader, select, image, disclosure and
// toggle; any other role is reported as given.
func Role(role string, value ...string) semantic.DescriptionOp {
	if len(value) > 0 && value[0] != "" {
		return semantic.DescriptionOp(role + ":" + value[0])
	}
	return semantic.DescriptionOp(role)
}
