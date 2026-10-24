// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

// The node holding keyboard focus is marked (Keel patch), also when the
// focused handler sits in a plain area inside it.
func TestSemanticFocus(t *testing.T) {
	var r Router
	first, second := new(int), new(int)
	frame := func() map[string]bool {
		for _, tag := range []*int{first, second} {
			r.Event(key.FocusFilter{Target: tag})
		}
		var ops op.Ops
		a := clip.Rect(image.Rect(0, 0, 100, 20)).Push(&ops)
		semantic.Button.Add(&ops)
		semantic.LabelOp("First").Add(&ops)
		event.Op(&ops, first)
		a.Pop()
		b := clip.Rect(image.Rect(0, 20, 100, 40)).Push(&ops)
		semantic.Button.Add(&ops)
		semantic.LabelOp("Second").Add(&ops)
		inner := clip.Rect(image.Rect(10, 25, 90, 35)).Push(&ops) // no semantics of its own
		event.Op(&ops, second)
		inner.Pop()
		b.Pop()
		r.Frame(&ops)
		focused := map[string]bool{}
		for _, n := range r.AppendSemantics(nil) {
			if n.Desc.Focused {
				focused[n.Desc.Label] = true
			}
		}
		return focused
	}
	if f := frame(); len(f) != 0 {
		t.Fatalf("focused before focusing: %v", f)
	}
	r.Source().Execute(key.FocusCmd{Tag: first})
	if f := frame(); len(f) != 1 || !f["First"] {
		t.Fatalf("focus on First marks %v", f)
	}
	r.Source().Execute(key.FocusCmd{Tag: second})
	if f := frame(); len(f) != 1 || !f["Second"] {
		t.Fatalf("focus in Second's inner area marks %v", f)
	}
	r.Source().Execute(key.FocusCmd{})
	if f := frame(); len(f) != 0 {
		t.Fatalf("focused after clearing focus: %v", f)
	}
	// Focus by node: Second's handler sits in an area inside it.
	ids := map[string]SemanticID{}
	for _, n := range r.AppendSemantics(nil) {
		ids[n.Desc.Label] = n.ID
	}
	if !r.FocusSemantic(ids["Second"]) {
		t.Fatal("FocusSemantic(Second) found no handler")
	}
	if f := frame(); len(f) != 1 || !f["Second"] {
		t.Fatalf("after FocusSemantic(Second), focused %v", f)
	}
	if !r.FocusSemantic(ids["First"]) {
		t.Fatal("FocusSemantic(First) found no handler")
	}
	if f := frame(); len(f) != 1 || !f["First"] {
		t.Fatalf("after FocusSemantic(First), focused %v", f)
	}
	if r.FocusSemantic(9999) {
		t.Fatal("FocusSemantic of an unknown node succeeded")
	}
}
