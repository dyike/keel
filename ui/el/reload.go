package el

import (
	"encoding/json"
	"strconv"
	"strings"
)

// reloadEntry is the state of one element carried across a keel run reload.
// Path is the element's place in the tree, as child indices; IDs are left
// out, since components generate theirs per process. Who names the element
// by role, accessible name and placeholder: state returns only to the same
// kind of element at the same place, so an edited layout never puts text
// into a different input.
type reloadEntry struct {
	Path string  `json:"p"`
	Who  string  `json:"w"`
	X    int     `json:"x,omitempty"`
	Y    int     `json:"y,omitempty"`
	Text *string `json:"t,omitempty"`
}

func reloadIdentity(n *Node) string {
	placeholder := ""
	if n.input != nil {
		placeholder = n.input.placeholder
	}
	return strings.Join([]string{n.role, n.name, placeholder}, "\x1f")
}

// SaveReloadState implements core.ReloadState: scroll offsets and the text
// of inputs in the main tree. Password inputs and inputs holding tokens are
// left out; a password must never reach a file.
func (r *RootWidget) SaveReloadState() []byte {
	var out []reloadEntry
	var walk func(n *Node, path string)
	walk = func(n *Node, path string) {
		if n == nil {
			return
		}
		e := reloadEntry{Path: path}
		keep := false
		if st := r.store.states[n.key]; st != nil && (st.scrollX != 0 || st.scrollY != 0) {
			e.X, e.Y, keep = st.scrollX, st.scrollY, true
		}
		if in := n.input; in != nil && !in.password {
			text, ok := "", false
			switch {
			case in.document != nil:
				if c := in.document.Content(); c.TokenCount() == 0 {
					text, ok = c.Text(), true
				}
			case in.bind != nil:
				text, ok = *in.bind, true
			}
			if ok && text != "" {
				e.Text, keep = &text, true
			}
		}
		if keep {
			e.Who = reloadIdentity(n)
			out = append(out, e)
		}
		for i, c := range n.children {
			walk(c.node(), childPath(path, i))
		}
	}
	walk(r.mainTree, "")
	if len(out) == 0 {
		return nil
	}
	b, _ := json.Marshal(out)
	return b
}

// RestoreReloadState implements core.ReloadState.
func (r *RootWidget) RestoreReloadState(b []byte) {
	var entries []reloadEntry
	if json.Unmarshal(b, &entries) != nil || len(entries) == 0 {
		return
	}
	r.reload = make(map[string]reloadEntry, len(entries))
	for _, e := range entries {
		r.reload[e.Path] = e
	}
}

func childPath(parent string, i int) string {
	return parent + "." + strconv.Itoa(i)
}

// applyReload gives the first frame's elements their saved state, once.
// Elements not drawn in that frame, such as a hidden tab's, start afresh.
func (r *RootWidget) applyReload(n *Node, path string) {
	if n == nil {
		return
	}
	if e, ok := r.reload[path]; ok && e.Who == reloadIdentity(n) {
		if e.X != 0 || e.Y != 0 {
			st := r.store.get(n.key)
			st.scrollX, st.scrollY = e.X, e.Y
		}
		if in := n.input; in != nil && e.Text != nil && !in.password {
			if in.bind != nil {
				*in.bind = *e.Text
			}
			if in.document != nil {
				_ = in.document.SetText(*e.Text)
			}
		}
	}
	for i, c := range n.children {
		r.applyReload(c.node(), childPath(path, i))
	}
}
