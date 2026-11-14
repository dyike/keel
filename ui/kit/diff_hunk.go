package kit

import "strings"

// DiffHunk is a snapshot of one review region. Source ranges are zero-based,
// half-open; an insertion or deletion has an empty range on one side.
// In patch inspection, a region is an imported hunk including its context,
// and its text contains only known source lines.
type DiffHunk struct {
	Index                  int
	BeforeStart, BeforeEnd int
	AfterStart, AfterEnd   int
	Before, After          string
	PatchOnly              bool
}

// Hunk returns a snapshot, independent of the currently selected region.
// Invalid indices return false. A snapshot does not track subsequent edits.
func (v *DiffView) Hunk(index int) (DiffHunk, bool) {
	if index < 0 || index >= len(v.changes) {
		return DiffHunk{}, false
	}
	if v.patch != nil {
		return v.patch.hunks[index], true
	}
	c := v.changes[index]
	return DiffHunk{Index: index, BeforeStart: c.oldStart, BeforeEnd: c.oldEnd,
		AfterStart: c.newStart, AfterEnd: c.newEnd,
		Before: strings.Join(v.a[c.oldStart:c.oldEnd], ""), After: strings.Join(v.b[c.newStart:c.newEnd], "")}, true
}

// KeepAt accepts a specific region without requiring SetCurrent first.
// Indices refer to the current comparison and may change after a mutation.
func (v *DiffView) KeepAt(index int) bool { return v.resolveAt(index, true) }

// RevertAt restores a specific region without requiring SetCurrent first.
func (v *DiffView) RevertAt(index int) bool { return v.resolveAt(index, false) }
