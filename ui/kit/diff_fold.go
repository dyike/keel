package kit

// Keep three unchanged context lines on either side of each change. Each
// omitted interval retains its first source line as the editor's fold anchor;
// the editor keeps source coordinates, selection and copy over the full text.
func (v *DiffView) foldContext() {
	for side, ed := range []*CodeEditorView{v.old, v.new} {
		ed.foldRanges = make(map[int]int)
		clear(ed.folds)
		if len(v.changes) > 0 {
			previous := 0
			add := func(end int) {
				if end > previous+1 {
					ed.foldRanges[previous] = end - 1
					if !v.expanded {
						ed.folds[previous] = true
					}
				}
			}
			for _, c := range v.changes {
				start, end := c.oldStart, c.oldEnd
				if side == 1 {
					start, end = c.newStart, c.newEnd
				}
				add(max(0, start-3))
				previous = max(previous, end+3)
			}
			add(ed.Lines())
		}
		ed.invalidateRows()
	}
	v.scrollKnown = false
}
