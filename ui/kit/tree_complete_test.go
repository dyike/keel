package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"reflect"
	"testing"
)

func TestTreeCustomRowsKeepActionsIndependent(t *testing.T) {
	expanded, actions := 0, 0
	tr := Tree(&TreeNode{ID: "p", Label: "Parent", Children: []*TreeNode{{ID: "c", Label: "Child"}}}).RowHeight(48).Indent(24).
		OnExpand(func(id string, on bool) {
			if id == "p" && on {
				expanded++
			}
		})
	button := Button("Inspect", func() { actions++ })
	var seen TreeItemContext
	tr.RenderItem(func(state TreeItemContext) el.View {
		if state.ID == "c" {
			seen = state
		}
		return el.ViewFunc(func(cx *el.Context) el.Element {
			d := el.Div().Row().Child(el.Text("Custom " + state.Label))
			if state.ID == "c" {
				d.Child(button.Render(cx))
			}
			return d
		})
	})
	h := page(tr)
	click(t, h, locale.Current().MoreOptions+" Parent")
	h.Frame()
	if expanded != 1 || tr.Value() != "" || !shown(h, "Custom Child") {
		t.Fatal("twisty selected or failed", expanded, tr.Value())
	}
	click(t, h, "Inspect")
	h.Frame()
	if actions != 1 || tr.Value() != "" {
		t.Fatal("child action selected parent")
	}
	click(t, h, "Custom Child")
	h.Frame()
	if tr.Value() != "c" || seen.Depth != 1 || !seen.Selected || seen.HasChildren {
		t.Fatal("row metadata", tr.Value(), seen)
	}
	tr.SetNodeDisabled("c", true)
	h.Frame()
	click(t, h, "Inspect")
	if actions != 1 {
		t.Fatal("disabled nested action")
	}
	tr.RenderItem(nil)
	h.Frame()
	if shown(h, "Custom Child") || !shown(h, "Child") {
		t.Fatal("default restore")
	}
}

func TestTreeChildrenUpdateAtomicAndPreservesState(t *testing.T) {
	tr := Tree(&TreeNode{ID: "a", Children: []*TreeNode{{ID: "keep"}, {ID: "remove"}}}, &TreeNode{ID: "b", Children: []*TreeNode{{ID: "other"}}}).MultiSelect()
	tr.SetSelectedIDs([]string{"keep", "remove", "other"})
	tr.SetExpanded("a", false)
	before := tr.Roots()
	for _, children := range [][]*TreeNode{{{ID: "other"}}, {{ID: ""}}, {{ID: "a"}}} {
		if tr.SetChildren("a", children...) == nil || !reflect.DeepEqual(before, tr.Roots()) {
			t.Fatal("invalid update changed tree")
		}
	}
	cycle := &TreeNode{ID: "cycle"}
	cycle.Children = []*TreeNode{cycle}
	if tr.SetChildren("a", cycle) == nil {
		t.Fatal("cycle accepted")
	}
	source := &TreeNode{ID: "keep", Label: "New label"}
	if err := tr.SetChildren("a", source, &TreeNode{ID: "new"}); err != nil {
		t.Fatal(err)
	}
	source.Label = "external"
	if tr.Expanded("a") || !reflect.DeepEqual(tr.SelectedIDs(), []string{"keep", "other"}) || tr.Node("keep").Label != "New label" {
		t.Fatal("state/ownership lost", tr.SelectedIDs())
	}
	snapshot := tr.Node("a")
	snapshot.Children[0].Label = "external2"
	if tr.Node("keep").Label != "New label" || !tr.SetNodeLabel("keep", "Updated") || tr.SetNodeLabel("missing", "x") {
		t.Fatal("node snapshot/label")
	}
}

func TestTreeLazyRequestsAndCollapseInvalidation(t *testing.T) {
	requests := map[string][]uint64{}
	tr := Tree(&TreeNode{ID: "a", Label: "A", Lazy: true}, &TreeNode{ID: "b", Label: "B", Lazy: true}).OnLoad(func(id string, token uint64) { requests[id] = append(requests[id], token) })
	h := page(tr)
	// Click the label: a narrow row's center can fall on its disclosure.
	clickRole(t, h, "", "A")
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if !tr.NodeLoading("a") || len(requests["a"]) != 1 {
		t.Fatal("keyboard lazy open")
	}
	tr.SetExpanded("a", true)
	tr.SetExpanded("b", true)
	if len(requests["a"]) != 1 {
		t.Fatal("duplicate in-flight request")
	}
	first := requests["a"][0]
	h.Key(key.NameLeftArrow, 0)
	h.Frame()
	if ok, _ := tr.SetChildResults("a", first, &TreeNode{ID: "stale"}); ok {
		t.Fatal("collapsed result accepted")
	}
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	token := requests["a"][1]
	if !tr.SetChildError("a", token, "network failed") {
		t.Fatal("error rejected")
	}
	h.Frame()
	click(t, h, locale.Current().Retry)
	h.Frame()
	if len(requests["a"]) != 3 || !tr.NodeLoading("a") {
		t.Fatal("retry missing")
	}
	token = requests["a"][2]
	if ok, err := tr.SetChildResults("a", token, &TreeNode{ID: "b"}); ok || err == nil || !tr.NodeLoading("a") {
		t.Fatal("invalid response applied")
	}
	if ok, err := tr.SetChildResults("a", token, &TreeNode{ID: "loaded", Label: "Loaded"}); !ok || err != nil {
		t.Fatal("valid response", ok, err)
	}
	if ok, err := tr.SetChildResults("b", requests["b"][0]); !ok || err != nil {
		t.Fatal("other branch request invalidated", ok, err)
	}
	h.Frame()
	if !shown(h, "Loaded") || tr.Node("b").Lazy || tr.NodeLoading("a") {
		t.Fatal("loaded state")
	}
	tr.ReloadNode("a")
	token = requests["a"][3]
	tr.SetNodeDisabled("a", true)
	if ok, _ := tr.SetChildResults("a", token); ok {
		t.Fatal("disabled response accepted")
	}
}

func TestTreeLoadReentrancyAndModelReplacement(t *testing.T) {
	var tr *TreeView
	tr = Tree(&TreeNode{ID: "p", Label: "Parent", Lazy: true}).OnLoad(func(id string, token uint64) {
		if ok, err := tr.SetChildResults(id, token, &TreeNode{ID: "c", Label: "Child"}); !ok || err != nil {
			t.Fatal(ok, err)
		}
	})
	h := page(tr)
	click(t, h, "Parent")
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if tr.NodeLoading("p") || !shown(h, "Child") {
		t.Fatal("synchronous load overwritten")
	}
	var token uint64
	tr.OnLoad(func(_ string, tok uint64) { token = tok })
	tr.ReloadNode("p")
	tr.SetRoots(&TreeNode{ID: "p", Label: "Replaced", Lazy: true})
	if ok, _ := tr.SetChildResults("p", token, &TreeNode{ID: "stale"}); ok {
		t.Fatal("old model request accepted")
	}
	tr.OnExpand(func(string, bool) { tr.SetRoots(&TreeNode{ID: "new", Label: "New"}) })
	tr.SetExpanded("p", false)
	h.Frame()
	click(t, h, "Replaced")
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if tr.Node("p") != nil || !shown(h, "New") {
		t.Fatal("expand callback state overwritten")
	}
}

func TestTreeCustomRowsVirtualizeAndScrollWithoutSelection(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			children := make([]*TreeNode, 10000)
			for i := range children {
				children[i] = &TreeNode{ID: fmt.Sprint(i), Label: fmt.Sprintf("Node %d", i)}
			}
			built := 0
			tr := Tree(&TreeNode{ID: "p", Label: "Parent", Children: children}).Height(160).RowHeight(40).RenderItem(func(state TreeItemContext) el.View {
				built++
				return el.ViewFunc(func(*el.Context) el.Element { return el.Text(state.Label) })
			})
			var cx *el.Context
			h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return tr.Render(ctx) }), 300, scale)
			tr.SetValue("0")
			settle(h)
			if built > 200 {
				t.Fatal("eager render", built)
			}
			if !tr.ScrollTo(cx, "9999") {
				t.Fatal("missing scroll target")
			}
			settle(h)
			if !shown(h, "Node 9999") || tr.Value() != "0" || built > 500 {
				t.Fatal("scroll selected or failed", tr.Value(), built)
			}
			if tr.ScrollTo(cx, "missing") {
				t.Fatal("accepted unknown scroll target")
			}
		})
	}
}
