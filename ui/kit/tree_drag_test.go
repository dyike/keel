package kit

import (
	"fmt"
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
)

// dragRow drives a drag of row from, with the pointer at y dp in list
// content coordinates, as the row element reports it.
func dragRow(tree *TreeView, from int, ys ...float32) {
	rowTop := float32(from) * tree.list.rowH
	tree.dragNode(from, el.DragEvent{Kind: el.DragStart, Y: tree.list.rowH / 2})
	for i, y := range ys {
		kind := el.DragMove
		if i == len(ys)-1 {
			kind = el.DragEnd
		}
		tree.dragNode(from, el.DragEvent{Kind: kind, Y: y - rowTop})
	}
}

func folderTree() *TreeView {
	return Tree(
		&TreeNode{ID: "src", Label: "src", Children: []*TreeNode{{ID: "main", Label: "main.go"}}},
		&TreeNode{ID: "docs", Label: "docs"},
		&TreeNode{ID: "readme", Label: "README.md"},
	)
}

func TestTreeDropIntoFolderAndInvalidDrops(t *testing.T) {
	var moves []string
	tree := folderTree().Reorderable(func(id, parent string, index int) { moves = append(moves, fmt.Sprint(id, ">", parent, ":", index)) })
	h := sized(300, tree)
	h.Frame()
	// README (row 2) onto the middle of src (row 0): into the folder, appended.
	dragRow(tree, 2, 40, 14)
	if len(moves) != 1 || moves[0] != "readme>src:1" || !tree.Expanded("src") {
		t.Fatalf("drop into folder: %v, src open %v", moves, tree.Expanded("src"))
	}
	h.Frame()
	// src into its own child is refused.
	dragRow(tree, 0, 30, 28*1+14)
	if len(moves) != 1 {
		t.Fatalf("a folder moved into its own child: %v", moves)
	}
	// The top quarter of a row inserts before it.
	h.Frame()
	docs := tree.index("docs")
	dragRow(tree, docs, 20, 2)
	if roots := tree.Roots(); len(moves) != 2 || roots[0].ID != "docs" {
		t.Fatalf("drop before src: %v %v", moves, roots)
	}
}

func TestTreeDragExpandsAndScrolls(t *testing.T) {
	tree := folderTree().Reorderable(func(string, string, int) {})
	h := sized(300, tree)
	h.Frame()
	readme := tree.index("readme")
	tree.dragNode(readme, el.DragEvent{Kind: el.DragStart, Y: 14})
	tree.dragNode(readme, el.DragEvent{Kind: el.DragMove, Y: 14 - float32(readme)*28})
	deadline := time.Now().Add(2 * time.Second)
	for !tree.Expanded("src") && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		h.Frame()
	}
	if !tree.Expanded("src") {
		t.Fatal("resting over a collapsed folder did not expand it")
	}
	tree.dragNode(tree.index("readme"), el.DragEvent{Kind: el.DragEnd, Canceled: true})

	var roots []*TreeNode
	for i := range 60 {
		roots = append(roots, &TreeNode{ID: fmt.Sprint("n", i), Label: fmt.Sprint("节点 ", i)})
	}
	long := Tree(roots...).Height(140).Reorderable(func(string, string, int) {})
	lh := sized(300, long)
	lh.Frame()
	long.dragNode(0, el.DragEvent{Kind: el.DragStart, Y: 14})
	long.dragNode(0, el.DragEvent{Kind: el.DragMove, Y: 135}) // near the bottom edge
	first := long.drag.target
	deadline = time.Now().Add(2 * time.Second)
	for long.drag != nil && long.drag.target < first+5 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		lh.Frame()
	}
	if long.drag == nil || long.drag.target < first+5 {
		t.Fatal("dragging at the bottom edge did not scroll")
	}
}
