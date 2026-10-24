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

// A node keeps its semantic ID when it gains an input handler (Keel patch),
// and nodes that say the same thing keep distinct IDs in order.
func TestSemanticIDIgnoresHandler(t *testing.T) {
	var r Router
	tag := new(int)
	frame := func(handler bool) map[string][]SemanticID {
		if handler {
			r.Event(pointer.Filter{Target: tag, Kinds: pointer.Press})
		}
		var ops op.Ops
		for i, label := range []string{"Row", "Row", "Other"} {
			area := clip.Rect(image.Rect(0, i*10, 100, i*10+10)).Push(&ops)
			if handler && i == 0 {
				event.Op(&ops, tag)
			}
			semantic.Button.Add(&ops)
			semantic.LabelOp(label).Add(&ops)
			area.Pop()
		}
		r.Frame(&ops)
		ids := map[string][]SemanticID{}
		for _, n := range r.AppendSemantics(nil) {
			if n.Desc.Label != "" {
				ids[n.Desc.Label] = append(ids[n.Desc.Label], n.ID)
			}
		}
		return ids
	}
	before := frame(false)
	after := frame(true)
	if len(before["Row"]) != 2 || before["Row"][0] == before["Row"][1] {
		t.Fatalf("identical rows share an ID: %v", before)
	}
	for _, label := range []string{"Row", "Other"} {
		for i := range before[label] {
			if before[label][i] != after[label][i] {
				t.Errorf("%s %d: ID %d became %d when the node gained a handler", label, i, before[label][i], after[label][i])
			}
		}
	}
}
