package kit

import (
	"gioui.org/io/input"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func renderView(v el.View, width, scale int) *uitest.Harness {
	root := el.Embed(v)
	return uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
		gtx.Constraints.Max = image.Pt(width*scale, 1000*scale)
		root.Layout(gtx)
	})
}
func semanticNode(h *uitest.Harness, role string) (input.SemanticNode, bool) {
	var found input.SemanticNode
	ok := false
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Description == role {
			found = n
			ok = true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	return found, ok
}

type viewFunc func(*el.Context) el.Element

func (f viewFunc) Render(cx *el.Context) el.Element { return f(cx) }
func render(fn viewFunc) *uitest.Harness            { return uitest.New(el.Embed(fn)) }
func node(h *uitest.Harness, name string) (input.SemanticNode, bool) {
	var found input.SemanticNode
	ok := false
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if ok {
			return
		}
		if n.Desc.Label == name { // the first, outermost match: the control, not its label text
			found = n
			ok = true
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	return found, ok
}
func bounds(h *uitest.Harness, name string) image.Rectangle {
	n, _ := node(h, name)
	return n.Desc.Bounds
}
func click(t *testing.T, h *uitest.Harness, name string) {
	t.Helper()
	n, ok := node(h, name)
	if !ok {
		t.Fatalf("missing node %q", name)
	}
	p := n.Desc.Bounds.Min.Add(n.Desc.Bounds.Size().Div(2))
	h.Click(float32(p.X), float32(p.Y))
}

// clickRole clicks the first element with role (role or role:value) and name.
func clickRole(t *testing.T, h *uitest.Harness, role, name string) {
	t.Helper()
	var found image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		d := n.Desc.Description
		if found.Empty() && n.Desc.Label == name && (d == role || len(d) > len(role) && d[:len(role)+1] == role+":") {
			found = n.Desc.Bounds
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	if found.Empty() {
		t.Fatalf("missing %s %q", role, name)
	}
	h.Click(center(found))
}

// clickClass clicks the first element of a Gio semantic class (Button,
// Editor, …) with name, skipping label text of the same name.
func clickClass(t *testing.T, h *uitest.Harness, class, name string) {
	t.Helper()
	var found image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if found.Empty() && n.Desc.Label == name && n.Desc.Class.String() == class {
			found = n.Desc.Bounds
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	if found.Empty() {
		t.Fatalf("missing %s %q", class, name)
	}
	h.Click(center(found))
}
