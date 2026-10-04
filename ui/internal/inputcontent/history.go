package inputcontent

// Snapshot stores references alongside text and selection, including edits
// that change a reference ID without changing any characters.
type Snapshot struct {
	Content   Content
	Selection Range
}
type History struct {
	Current    Snapshot
	undo, redo []Snapshot
}

func (h *History) Set(value Snapshot) { h.Current = value; h.undo = nil; h.redo = nil }
func (h *History) Commit(value Snapshot) {
	h.undo = append(h.undo, h.Current)
	if len(h.undo) > 100 {
		copy(h.undo, h.undo[len(h.undo)-100:])
		h.undo = h.undo[:100]
	}
	h.Current = value
	h.redo = nil
}
func (h *History) Undo() bool {
	if len(h.undo) == 0 {
		return false
	}
	h.redo = append(h.redo, h.Current)
	h.Current = h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	return true
}
func (h *History) Redo() bool {
	if len(h.redo) == 0 {
		return false
	}
	h.undo = append(h.undo, h.Current)
	h.Current = h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	return true
}
