// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

// semanticNode is one area of a test frame.
type semanticNode struct {
	label    string
	tag      event.Tag // nil: no handler
	selected bool
}

func semanticFrame(r *Router, nodes []semanticNode, ask bool) []SemanticID {
	for _, n := range nodes {
		if n.tag != nil {
			r.Event(pointer.Filter{Target: n.tag, Kinds: pointer.Press})
		}
	}
	var ops op.Ops
	for i, n := range nodes {
		area := clip.Rect(image.Rect(0, i*10, 100, i*10+10)).Push(&ops)
		if n.tag != nil {
			event.Op(&ops, n.tag)
		}
		semantic.CheckBox.Add(&ops)
		semantic.LabelOp(n.label).Add(&ops)
		semantic.SelectedOp(n.selected).Add(&ops)
		area.Pop()
	}
	r.Frame(&ops)
	if !ask {
		return nil
	}
	var ids []SemanticID
	for _, n := range r.AppendSemantics(nil) {
		if n.Desc.Label != "" {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

// Semantic IDs stay with their nodes (Keel patch): through a change of
// label or state, a frame nobody asked about, and a node gaining a handler;
// nodes that say the same thing keep distinct IDs.
func TestSemanticIDsStayWithNodes(t *testing.T) {
	var r Router
	box, save := new(int), new(int)
	nodes := []semanticNode{{label: "Paid", tag: box}, {label: "Save", tag: save}, {label: "Row"}, {label: "Row"}}
	before := semanticFrame(&r, nodes, true)
	if len(before) != 4 || before[2] == before[3] {
		t.Fatalf("ids %v", before)
	}
	nodes[0].selected = true        // a checkbox toggles
	nodes[1].label = "Saved 1"      // a label changes
	semanticFrame(&r, nodes, false) // nobody asks about this frame
	nodes[2].tag = new(int)         // a row gains a handler
	after := semanticFrame(&r, nodes, true)
	for i := range before {
		if before[i] != after[i] {
			t.Errorf("node %d (%s): ID %d became %d", i, nodes[i].label, before[i], after[i])
		}
	}
}
