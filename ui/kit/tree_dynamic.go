package kit

import "fmt"

type treeLoad struct {
	token   uint64
	loading bool
	err     string
}

func cloneTree(roots []*TreeNode) ([]*TreeNode, map[string]*TreeNode, error) {
	nodes := make(map[string]*TreeNode)
	var clone func([]*TreeNode) ([]*TreeNode, error)
	clone = func(source []*TreeNode) ([]*TreeNode, error) {
		var out []*TreeNode
		for _, n := range source {
			if n == nil {
				continue
			}
			if n.ID == "" {
				return nil, fmt.Errorf("kit.Tree: empty node ID")
			}
			if nodes[n.ID] != nil {
				return nil, fmt.Errorf("kit.Tree: duplicate node ID %s", n.ID)
			}
			copy := *n
			nodes[n.ID] = &copy
			children, err := clone(n.Children)
			if err != nil {
				return nil, err
			}
			copy.Children = children
			out = append(out, &copy)
		}
		return out, nil
	}
	owned, err := clone(roots)
	return owned, nodes, err
}
func (v *TreeView) install(roots []*TreeNode, nodes map[string]*TreeNode) {
	v.roots, v.nodes = roots, nodes
	for id := range v.open {
		if nodes[id] == nil {
			delete(v.open, id)
		}
	}
	for id := range v.loads {
		if nodes[id] == nil {
			delete(v.loads, id)
		}
	}
	selection := v.selection
	selection.Keep(func(id string) bool { return nodes[id] != nil })
	// Refreshing children must not reopen a branch the user just collapsed.
	if nodes[v.selected] == nil {
		v.selected = ""
	}
	v.selection = selection
	v.reveal = v.selected != ""
}

// SetChildren atomically replaces one node's children, retaining other branches
// and surviving selections/expansion. Invalid IDs/cycles leave the tree intact.
// Applied nodes are deep-copied. An empty result turns a Lazy node into a leaf.
// Pending requests for this subtree are invalidated, not unrelated requests.
func (v *TreeView) SetChildren(id string, children ...*TreeNode) error {
	if v.nodes[id] == nil {
		return fmt.Errorf("kit.Tree: unknown node %q", id)
	}
	roots, nodes, _ := cloneTree(v.roots)
	nodes[id].Children = children
	nodes[id].Lazy = false
	owned, index, err := cloneTree(roots)
	if err != nil {
		return err
	}
	var invalidate func(*TreeNode)
	invalidate = func(n *TreeNode) {
		delete(v.loads, n.ID)
		for _, c := range n.Children {
			invalidate(c)
		}
	}
	invalidate(v.nodes[id])
	v.install(owned, index)
	return nil
}

// Node returns a deep snapshot of a node; nil means no such ID.
func (v *TreeView) Node(id string) *TreeNode {
	n := v.nodes[id]
	if n == nil {
		return nil
	}
	roots, _, _ := cloneTree([]*TreeNode{n})
	return roots[0]
}

// SetNodeLabel updates a label without replacing children or selection.
func (v *TreeView) SetNodeLabel(id, label string) bool {
	if n := v.nodes[id]; n != nil {
		n.Label = label
		return true
	}
	return false
}

// OnExpand observes user expansion/collapse after internal state changes.
// SetExpanded and automatic ancestor expansion are silent.
func (v *TreeView) OnExpand(fn func(string, bool)) *TreeView { v.onExpand = fn; return v }

// OnLoad supplies lazy children. Deliver asynchronous results on the UI thread
// using core.Update and SetChildResults/SetChildError. Replacing the provider
// invalidates outstanding requests; explicitly reopen or ReloadNode to retry.
func (v *TreeView) OnLoad(fn func(string, uint64)) *TreeView { v.onLoad = fn; clear(v.loads); return v }
func (v *TreeView) expand(id string, on, user bool) {
	n := v.nodes[id]
	if n == nil || (user && len(n.Children) == 0 && !n.Lazy) {
		return
	}
	changed := v.open[id] != on
	v.open[id] = on
	if !on {
		delete(v.loads, id)
	}
	if changed && user && v.onExpand != nil {
		v.onExpand(id, on)
	}
	if on && v.open[id] && v.nodes[id] != nil && v.nodes[id].Lazy {
		v.requestChildren(id, false)
	}
}
func (v *TreeView) requestChildren(id string, force bool) bool {
	n := v.nodes[id]
	if n == nil || n.Disabled || v.disabled || !v.open[id] || v.onLoad == nil {
		return false
	}
	if !force && (!n.Lazy || v.loads[id].loading || v.loads[id].err != "") {
		return false
	}
	v.loadToken++
	token := v.loadToken
	v.loads[id] = treeLoad{token: token, loading: true}
	v.onLoad(id, token)
	return true
}

// ReloadNode starts a new request for an expanded enabled node; earlier results
// for it become stale. Existing children remain visible until replaced.
func (v *TreeView) ReloadNode(id string) bool { return v.requestChildren(id, true) }
func (v *TreeView) accepts(id string, token uint64) bool {
	n := v.nodes[id]
	state := v.loads[id]
	return n != nil && !n.Disabled && !v.disabled && v.open[id] && state.loading && state.token == token
}

// SetChildResults rejects stale, collapsed, removed or disabled requests.
// A valid token with invalid children returns an error without changing data.
func (v *TreeView) SetChildResults(id string, token uint64, children ...*TreeNode) (bool, error) {
	if !v.accepts(id, token) {
		return false, nil
	}
	if err := v.SetChildren(id, children...); err != nil {
		return false, err
	}
	return true, nil
}

// SetChildError completes the current request with a retryable error.
func (v *TreeView) SetChildError(id string, token uint64, message string) bool {
	if !v.accepts(id, token) {
		return false
	}
	v.loads[id] = treeLoad{token: token, err: message}
	return true
}

// NodeLoading and NodeError expose per-node loading state for custom rows.
func (v *TreeView) NodeLoading(id string) bool { return v.loads[id].loading }
func (v *TreeView) NodeError(id string) string { return v.loads[id].err }
