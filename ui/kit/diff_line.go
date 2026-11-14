package kit

// DiffSide identifies the source of a line, independent of display mode.
type DiffSide uint8

const (
	DiffSideBefore DiffSide = iota
	DiffSideAfter
)

// DiffLine identifies a zero-based source line. Context in a unified patch
// belongs to After; file/hunk headers and missing-newline markers have no line.
type DiffLine struct {
	Side DiffSide
	Line int
}

// OnLineNumberPress reports the source identity of a pressed gutter number.
// Normal line selection is preserved; fold arrows and patch headers do not notify.
func (v *DiffView) OnLineNumberPress(fn func(DiffLine)) *DiffView { v.onLineNumberPress = fn; return v }
func (v *DiffView) sourceLine(pane, row int) (DiffLine, bool) {
	if row < 0 {
		return DiffLine{}, false
	}
	if v.patch != nil {
		line, ok := v.patch.lineRefs[pane][row]
		return line, ok
	}
	if pane == 0 {
		return DiffLine{DiffSideBefore, row}, row < v.old.Lines()
	}
	if pane == 1 || len(v.changes) == 0 {
		return DiffLine{DiffSideAfter, row}, row < v.new.Lines()
	}
	line, ok := v.unifiedLines[row]
	return line, ok
}
func (v *DiffView) pressLine(pane, row int) {
	if v.disabled || v.onLineNumberPress == nil {
		return
	}
	if line, ok := v.sourceLine(pane, row); ok {
		v.onLineNumberPress(line)
	}
}
