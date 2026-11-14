package kit

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// DiffMode selects the presentation. Auto uses a unified view on narrow screens.
type DiffMode uint8

const (
	DiffModeAuto DiffMode = iota
	DiffModeUnified
	DiffModeSplit
)

// DiffValue contains the reference text and proposed text, including line endings.
type DiffValue struct{ Before, After string }

// DiffStats counts changed source lines, excluding unchanged context.
type DiffStats struct{ Added, Deleted, Changes int }

// DiffView compares text and lets the user review changes. CodeEditor supplies
// selection, copy, keyboard navigation and viewport-based text rendering.
type DiffView struct {
	hunkActions       func(DiffHunk) []DiffAction
	actionGeneration  uint64
	onLineNumberPress func(DiffLine)
	unifiedLines      map[int]DiffLine
	patch             *diffPatchInput
	expanded          bool
	stats             DiffStats
	scrollKnown       bool
	lastScroll        [2]float32
	inline            []diffInline
	decorations       [3]*CodeDecorationCollection
	colors            [2]color.NRGBA
	dirty             bool
	value             DiffValue
	mode              DiffMode
	name              string
	height            float32
	disabled          bool
	current           int
	a, b              []string
	changes           []diffRange
	positions         []int
	old, new, unified *CodeEditorView
	onChange          func(DiffValue)
}

func Diff(before, after string) *DiffView {
	v := &DiffView{height: 360, current: -1, old: CodeEditor(""), new: CodeEditor(""), unified: CodeEditor("")}
	for _, ed := range []*CodeEditorView{v.old, v.new, v.unified} {
		ed.SetReadOnly(true)
	}
	for pane, ed := range []*CodeEditorView{v.old, v.new, v.unified} {
		ed.OnLineNumberPress(func(line int) { v.pressLine(pane, line) })
	}
	v.decorations = [3]*CodeDecorationCollection{v.old.Decorations(), v.new.Decorations(), v.unified.Decorations()}
	v.SetValue(DiffValue{before, after})
	return v
}
func (v *DiffView) Name(name string) *DiffView { v.name = name; return v }
func (v *DiffView) Height(dp float32) *DiffView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}
func (v *DiffView) Mode(mode DiffMode) *DiffView {
	if mode <= DiffModeSplit {
		v.mode = mode
	}
	return v
}
func (v *DiffView) Value() DiffValue                      { return v.value }
func (v *DiffView) OnChange(fn func(DiffValue)) *DiffView { v.onChange = fn; return v }
func (v *DiffView) SetDisabled(on bool)                   { v.disabled = on }
func (v *DiffView) Changes() int                          { return len(v.changes) }
func (v *DiffView) Current() int                          { return v.current }

// SetValue replaces both texts without notifying OnChange.
func (v *DiffView) SetValue(value DiffValue) {
	if v.patch == nil && v.value == value && v.a != nil {
		return
	}
	v.patch = nil
	v.value = value
	v.dirty = true
	v.a, v.b, v.changes = diffText(value.Before, value.After)
	v.stats = DiffStats{Changes: len(v.changes)}
	for _, c := range v.changes {
		v.stats.Added += c.newEnd - c.newStart
		v.stats.Deleted += c.oldEnd - c.oldStart
	}
	v.old.SetValue(value.Before)
	v.new.SetValue(value.After)
	if len(v.changes) == 0 {
		v.current = -1
	} else {
		v.current = min(max(v.current, 0), len(v.changes)-1)
	}
	v.refreshUnified()

}

// Stats returns source-line counts, independent of display mode or context.
func (v *DiffView) Stats() DiffStats { return v.stats }

// Language applies the existing optional syntax provider to the source panes.
// The unified view remains a patch, with diff decorations rather than source syntax.
func (v *DiffView) Language(name string) *DiffView {
	v.old.Language(name)
	v.new.Language(name)
	return v
}

// Expanded controls unchanged context in unified and split text comparisons.
// Split folding preserves full source text for selection/copy. Patch export
// always uses three context lines; imported patches cannot reveal missing text.
func (v *DiffView) Expanded(on bool) *DiffView {
	if v.expanded != on {
		v.expanded = on
		if v.patch == nil {
			v.refreshUnified()
		}
	}
	return v
}
func (v *DiffView) refreshUnified() {
	context := 3
	if v.expanded {
		context = max(len(v.a), len(v.b))
	}
	patch, positions := diffPatchContext(v.a, v.b, v.changes, "before", "after", context)
	v.positions = positions
	v.unifiedLines = nil
	if len(v.changes) > 0 {
		if parsed, err := parseDiffPatch(patch); err == nil {
			v.unifiedLines = parsed.lineRefs[2]
		}
	}
	v.buildInline()
	if len(v.changes) == 0 {
		patch = v.value.Before
	}
	v.unified.SetValue(patch)
	v.foldContext()
	v.rebuildActions()
	v.dirty = true
	v.reveal()
}

// SetCurrent selects a zero-based change without changing the texts or notifying.
func (v *DiffView) SetCurrent(index int) bool {
	if index < 0 || index >= len(v.changes) {
		return false
	}
	v.current = index
	v.reveal()
	return true
}
func (v *DiffView) reveal() {
	v.scrollKnown = false
	if v.current < 0 || v.current >= len(v.changes) {
		return
	}
	c := v.changes[v.current]
	v.old.SetCursor(c.oldStart, 0)
	v.new.SetCursor(c.newStart, 0)
	v.unified.SetCursor(v.positions[v.current], 0)
}

// Revert restores the reference lines of the current change in After.
func (v *DiffView) Revert() bool { return v.resolveAt(v.current, false) }

// Keep accepts the proposed lines of the current change into Before.
func (v *DiffView) Keep() bool { return v.resolveAt(v.current, true) }
func (v *DiffView) resolveAt(index int, keep bool) bool {
	if v.patch != nil || v.disabled || index < 0 || index >= len(v.changes) {
		return false
	}
	c := v.changes[index]
	value := v.value
	if keep {
		value.Before = strings.Join(v.a[:c.oldStart], "") + strings.Join(v.b[c.newStart:c.newEnd], "") + strings.Join(v.a[c.oldEnd:], "")
	} else {
		value.After = strings.Join(v.b[:c.newStart], "") + strings.Join(v.a[c.oldStart:c.oldEnd], "") + strings.Join(v.b[c.newEnd:], "")
	}
	v.SetValue(value)
	if v.onChange != nil {
		v.onChange(v.value)
	}
	return true
}

// Patch returns a unified patch with three context lines. Names are quoted.
func (v *DiffView) Patch(oldName, newName string) string {
	if v.patch != nil {
		return v.patch.raw
	}
	p, _ := diffPatch(v.a, v.b, v.changes, oldName, newName)
	return p
}
func (v *DiffView) Render(cx *el.Context) el.Element {
	v.decorate()
	text := locale.Current()
	width, _ := cx.ViewportSize()
	split := v.mode == DiffModeSplit || v.mode == DiffModeAuto && width >= 720*max(float32(1), cx.TextScale())
	button := func(label string, enabled bool, fn func()) el.Element {
		b := Button(label, fn).Variant(ButtonSecondary).Size(44)
		b.SetDisabled(!enabled)
		return b.Render(cx)
	}
	head := el.Div().Row().Wrap().NoShrink().Gap(theme.SpaceSm).Child(
		button(text.DiffPrevious, v.current > 0, func() { v.SetCurrent(v.current - 1) }), button(text.DiffNext, v.current >= 0 && v.current+1 < len(v.changes), func() { v.SetCurrent(v.current + 1) }),
		button(text.DiffKeep, v.patch == nil && v.current >= 0, func() { v.Keep() }), button(text.DiffRevert, v.patch == nil && v.current >= 0, func() { v.Revert() }),
		button(text.Copy, len(v.changes) > 0, func() { el.WriteClipboard(v.Patch("before", "after")) }),
	)
	if v.patch == nil {
		label := text.DiffExpand
		if v.expanded {
			label = text.DiffCollapse
		}
		head.Child(button(label, len(v.changes) > 0, func() { v.Expanded(!v.expanded) }))
	}
	count := fmt.Sprintf("%d / %d", v.current+1, len(v.changes))
	root := el.Div().ID(autoID("diff", v)).Role("group").Name(v.name).Value(count).Disabled(v.disabled).Items(el.Stretch).Gap(theme.SpaceSm).Child(head, el.Text(fmt.Sprintf("+%d −%d · %s", v.stats.Added, v.stats.Deleted, count)).TextColor(theme.Muted))
	v.old.Name(text.DiffBefore).Height(v.height)
	v.new.Name(text.DiffAfter).Height(v.height)
	v.unified.Name(text.DiffUnified).Height(v.height)
	if split {
		root.Child(el.Div().Row().Gap(theme.SpaceSm).Decorate(v.syncSplitScroll).Child(el.Div().Grow().W(el.Dp(0)).Child(el.Text(text.DiffBefore).Bold(), v.old.Render(cx)), el.Div().Grow().W(el.Dp(0)).Child(el.Text(text.DiffAfter).Bold(), v.new.Render(cx))))
	} else {
		v.scrollKnown = false
		root.Child(v.unified.Render(cx))
	}
	return root
}

func (v *DiffView) decorate() {
	colors := [2]color.NRGBA{theme.Danger, theme.Success}
	for i := range colors {
		colors[i].A = 40
	}
	if !v.dirty && v.colors == colors {
		return
	}
	v.dirty, v.colors = false, colors
	var entries [3][]CodeDecoration
	add := func(which, line, end, shade int) {
		if line == end {
			return
		}
		entries[which] = append(entries[which], CodeDecoration{Range: CodeRange{Line: line, EndLine: end}, Style: CodeDecorationFill, Color: &colors[shade]})
	}
	if v.patch != nil {
		for _, mark := range v.patch.marks {
			add(mark.pane, mark.row, mark.row+1, mark.shade)
		}
		for i, c := range v.decorations {
			c.Set(entries[i]...)
		}
		return
	}
	for _, c := range v.changes {
		add(0, c.oldStart, c.oldEnd, 0)
		add(1, c.newStart, c.newEnd, 1)
	}
	if len(v.changes) > 0 {
		for i, line := range strings.Split(v.unified.Value(), "\n") {
			if i < 2 || line == "" {
				continue
			}
			if line[0] == '-' {
				add(2, i, i+1, 0)
			}
			if line[0] == '+' {
				add(2, i, i+1, 1)
			}
		}
	}
	strong := colors
	for i := range strong {
		strong[i].A = 100
	}
	span := func(which, line, start, end, shade int) {
		if start == end {
			return
		}
		entries[which] = append(entries[which], CodeDecoration{Range: CodeRange{Line: line, Col: start, EndLine: line, EndCol: end}, Style: CodeDecorationFill, Color: &strong[shade]})
	}
	for _, pair := range v.inline {
		for _, r := range pair.ranges {
			span(0, pair.oldLine, r.oldStart, r.oldEnd, 0)
			span(1, pair.newLine, r.newStart, r.newEnd, 1)
			span(2, pair.oldUnified, r.oldStart+1, r.oldEnd+1, 0)
			span(2, pair.newUnified, r.newStart+1, r.newEnd+1, 1)
		}
	}
	for i, c := range v.decorations {
		c.Set(entries[i]...)
	}
}
