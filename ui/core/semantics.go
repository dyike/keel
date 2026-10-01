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

// Role marks a node's role for agents, optionally with a value. Automation
// reads it as "role" or "role:value" from the node's description. Roles:
// link, tab, columnheader, select, disclosure, toggle (clickable); image, row, option,
// table, progressbar, slider, dialog, code, footnotes, accordion, badge, avatar, alert, tag, group, status, marker (others).
func Role(role string, value ...string) semantic.DescriptionOp {
	if len(value) > 0 && value[0] != "" {
		return semantic.DescriptionOp(role + ":" + value[0])
	}
	return semantic.DescriptionOp(role)
}
