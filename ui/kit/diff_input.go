package kit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type diffPatchMark struct{ pane, row, shade int }
type diffPatchInput struct {
	lineRefs      [3]map[int]DiffLine
	hunks         []DiffHunk
	raw, old, new string
	changes       []diffRange // display-row ranges; no omitted source lines are synthesized
	positions     []int
	marks         []diffPatchMark
	stats         DiffStats
}

var diffHunkHeader = regexp.MustCompile(`^@@ -([0-9]+)(?:,([0-9]+))? \+([0-9]+)(?:,([0-9]+))? @@(?:.*)$`)

func parseDiffPatch(raw string) (*diffPatchInput, error) {
	p := &diffPatchInput{raw: raw}
	for i := range p.lineRefs {
		p.lineRefs[i] = make(map[int]DiffLine)
	}
	if raw == "" {
		return p, nil
	}
	lines := diffLines(raw)
	var old, new strings.Builder
	oldRow, newRow := 0, 0
	headers := 0
	lastOld, lastNew := 0, 0
	oldEOF, newEOF := false, false
	for i := 0; i < len(lines); {
		line := strings.TrimSuffix(strings.TrimSuffix(lines[i], "\n"), "\r")
		if headers == 0 && (strings.HasPrefix(line, "diff --git ") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "new file mode ") || strings.HasPrefix(line, "deleted file mode ")) {
			i++
			continue
		}
		if headers == 0 && strings.HasPrefix(line, "--- ") {
			if i+1 >= len(lines) || !strings.HasPrefix(lines[i+1], "+++ ") {
				return nil, fmt.Errorf("diff: missing new-file header at line %d", i+1)
			}
			headers = 1
			i += 2
			continue
		}
		match := diffHunkHeader.FindStringSubmatch(line)
		if headers == 0 || match == nil {
			return nil, fmt.Errorf("diff: unsupported or malformed patch at line %d", i+1)
		}
		numbers := [4]int{}
		for j, s := range match[1:] {
			if j >= 4 {
				break
			}
			if s == "" {
				numbers[j] = 1
				continue
			}
			n, err := strconv.Atoi(s)
			if err != nil {
				return nil, fmt.Errorf("diff: invalid hunk range at line %d", i+1)
			}
			numbers[j] = n
		}
		os, oc, ns, nc := numbers[0], numbers[1], numbers[2], numbers[3]
		offset := func(start, count int) int {
			if count == 0 {
				return start
			}
			return start - 1
		}
		oa, na := offset(os, oc), offset(ns, nc)
		maxInt := int(^uint(0) >> 1)
		if oa < 0 || na < 0 || oc > maxInt-oa || nc > maxInt-na || oa < lastOld || na < lastNew || oc == 0 && nc == 0 {
			return nil, fmt.Errorf("diff: invalid or overlapping hunk at line %d", i+1)
		}
		lastOld, lastNew = oa+oc, na+nc
		p.positions = append(p.positions, i)
		startOld, startNew := oldRow, newRow
		old.WriteString(line + "\n")
		new.WriteString(line + "\n")
		oldRow++
		newRow++
		var hunkOld, hunkNew strings.Builder
		oldLeft, newLeft := oc, nc
		oldLine, newLine := os, ns
		i++
		canMark := false
		var lastPrefix byte
		trimLastNewline := func(b *strings.Builder) { s := b.String(); b.Reset(); b.WriteString(strings.TrimSuffix(s, "\n")) }
		for i < len(lines) {
			body := lines[i]
			if strings.TrimSuffix(strings.TrimSuffix(body, "\n"), "\r") == "\\ No newline at end of file" {
				if !canMark {
					return nil, fmt.Errorf("diff: misplaced newline marker at line %d", i+1)
				}
				if lastPrefix == '-' || lastPrefix == ' ' {
					trimLastNewline(&hunkOld)
					oldEOF = true
				}
				if lastPrefix == '+' || lastPrefix == ' ' {
					trimLastNewline(&hunkNew)
					newEOF = true
				}
				canMark = false
				i++
				continue
			}
			if oldLeft == 0 && newLeft == 0 {
				break
			}
			if body == "" || !strings.HasSuffix(body, "\n") {
				return nil, fmt.Errorf("diff: unterminated body at line %d", i+1)
			}
			if oldEOF && (body[0] == '-' || body[0] == ' ') || newEOF && (body[0] == '+' || body[0] == ' ') {
				return nil, fmt.Errorf("diff: source continues after missing-newline marker at line %d", i+1)
			}
			emitOld := func() {
				p.lineRefs[0][oldRow] = DiffLine{Side: DiffSideBefore, Line: oldLine - 1}
				fmt.Fprintf(&old, "%d %s", oldLine, body[1:])
				hunkOld.WriteString(body[1:])
				oldLine++
				oldRow++
				oldLeft--
			}
			emitNew := func() {
				p.lineRefs[1][newRow] = DiffLine{Side: DiffSideAfter, Line: newLine - 1}
				fmt.Fprintf(&new, "%d %s", newLine, body[1:])
				hunkNew.WriteString(body[1:])
				newLine++
				newRow++
				newLeft--
			}
			switch body[0] {
			case ' ':
				p.lineRefs[2][i] = DiffLine{Side: DiffSideAfter, Line: newLine - 1}
				emitOld()
				emitNew()
			case '-':
				p.lineRefs[2][i] = DiffLine{Side: DiffSideBefore, Line: oldLine - 1}
				p.marks = append(p.marks, diffPatchMark{0, oldRow, 0}, diffPatchMark{2, i, 0})
				emitOld()
				p.stats.Deleted++
			case '+':
				p.lineRefs[2][i] = DiffLine{Side: DiffSideAfter, Line: newLine - 1}
				p.marks = append(p.marks, diffPatchMark{1, newRow, 1}, diffPatchMark{2, i, 1})
				emitNew()
				p.stats.Added++
			default:
				return nil, fmt.Errorf("diff: invalid body at line %d", i+1)
			}
			if oldLeft < 0 || newLeft < 0 {
				return nil, fmt.Errorf("diff: hunk count exceeded at line %d", i+1)
			}
			canMark = true
			lastPrefix = body[0]
			i++
		}
		if oldLeft != 0 || newLeft != 0 {
			return nil, fmt.Errorf("diff: incomplete hunk")
		}
		p.hunks = append(p.hunks, DiffHunk{Index: len(p.hunks), BeforeStart: oa, BeforeEnd: oa + oc, AfterStart: na, AfterEnd: na + nc, Before: hunkOld.String(), After: hunkNew.String(), PatchOnly: true})
		p.changes = append(p.changes, diffRange{startOld, oldRow, startNew, newRow})
	}
	if len(p.changes) == 0 {
		return nil, fmt.Errorf("diff: patch has no hunks")
	}
	p.stats.Changes = len(p.changes)
	p.old, p.new = old.String(), new.String()
	return p, nil
}

// SetPatch atomically switches to inspection of a single-file unified patch.
// Missing source ranges are not reconstructed. Review mutations are unavailable.
// Invalid patches leave the current comparison untouched; no callback is fired.
func (v *DiffView) SetPatch(text string) error {
	p, err := parseDiffPatch(text)
	if err != nil {
		return err
	}
	v.patch = p
	v.value = DiffValue{}
	v.a, v.b = nil, nil
	v.inline = nil
	v.changes, v.positions, v.stats = p.changes, p.positions, p.stats
	v.old.SetValue(p.old)
	v.new.SetValue(p.new)
	v.unified.SetValue(text)
	v.current = -1
	if len(p.changes) > 0 {
		v.current = 0
	}
	v.rebuildActions()
	v.dirty = true
	v.reveal()
	return nil
}

// PatchInput reports the original patch and whether patch-inspection mode is active.
func (v *DiffView) PatchInput() (string, bool) {
	if v.patch == nil {
		return "", false
	}
	return v.patch.raw, true
}
