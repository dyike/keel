package el

import "slices"

// Mounting attaches views to a root without placing them in its view tree,
// the way GPUI's root view hosts dialogs, sheets and notifications. Each
// frame the root renders its own view, then every mounted view in mount
// order, and adds what they return as children of the root element. Mounted
// views should return overlays (Modal, Anchored) or absolute or hidden
// elements: they are laid out in the root element, outside its flow.

type mount struct {
	key  any
	view View
}

// Mount renders view with this root every frame until Unmount(key).
// Mounting a key again replaces its view in place. Call it from Render or a
// callback; the view shows from the next render.
func (cx *Context) Mount(key any, view View) {
	r := cx.root
	if i := slices.IndexFunc(r.mounts, func(m mount) bool { return m.key == key }); i >= 0 {
		r.mounts[i].view = view
		return
	}
	r.mounts = append(r.mounts, mount{key, view})
}

// Unmount stops rendering the view mounted at key.
func (cx *Context) Unmount(key any) {
	r := cx.root
	r.mounts = slices.DeleteFunc(r.mounts, func(m mount) bool { return m.key == key })
}

// Mounted reports whether a view is mounted at key.
func (cx *Context) Mounted(key any) bool {
	return slices.ContainsFunc(cx.root.mounts, func(m mount) bool { return m.key == key })
}

// renderTree renders the root view and then the mounted ones.
func (r *RootWidget) renderTree(cx *Context) *Node {
	tree := r.view.Render(cx).node()
	if len(r.mounts) == 0 {
		return tree
	}
	if tree.isText || tree.input != nil || tree.widget != nil {
		tree = Div().Child(tree).node() // a leaf cannot hold children
	}
	for _, m := range slices.Clone(r.mounts) { // a view may unmount itself
		if e := m.view.Render(cx); e != nil {
			tree.children = append(tree.children, e)
		}
	}
	return tree
}

// MountedView returns the view mounted at key, or nil.
func (cx *Context) MountedView(key any) View {
	for _, m := range cx.root.mounts {
		if m.key == key {
			return m.view
		}
	}
	return nil
}
