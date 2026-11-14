package kit

import (
	"strings"

	"github.com/dyike/keel/typesetting/segmenter"
)

type diffInline struct {
	oldLine, newLine, oldUnified, newUnified int
	ranges                                   []diffRange // rune offsets rather than line offsets
}

func diffGraphemes(text string) ([]string, []int) {
	if strings.HasSuffix(text, "\n") {
		text = strings.TrimSuffix(strings.TrimSuffix(text, "\n"), "\r")
	}
	runes := []rune(text)
	var seg segmenter.Segmenter
	seg.Init(runes)
	it := seg.GraphemeIterator()
	var tokens []string
	var offsets []int
	for it.Next() {
		g := it.Grapheme()
		tokens = append(tokens, string(g.Text))
		offsets = append(offsets, g.Offset)
	}
	offsets = append(offsets, len(runes))
	return tokens, offsets
}

func diffIntraline(before, after string, budget *int) []diffRange {
	a, ao := diffGraphemes(before)
	b, bo := diffGraphemes(after)
	changes := diffTokens(a, b, budget)
	for i, c := range changes {
		changes[i] = diffRange{ao[c.oldStart], ao[c.oldEnd], bo[c.newStart], bo[c.newEnd]}
	}
	return changes
}

func (v *DiffView) buildInline() {
	v.inline = nil
	budget := 1_000_000
	for k, c := range v.changes {
		row := v.positions[k]
		oldRows := make([]int, c.oldEnd-c.oldStart)
		for i, line := range v.a[c.oldStart:c.oldEnd] {
			oldRows[i] = row
			row++
			if !strings.HasSuffix(line, "\n") {
				row++
			}
		}
		newRows := make([]int, c.newEnd-c.newStart)
		for i, line := range v.b[c.newStart:c.newEnd] {
			newRows[i] = row
			row++
			if !strings.HasSuffix(line, "\n") {
				row++
			}
		}
		for i := 0; i < min(len(oldRows), len(newRows)); i++ {
			ranges := diffIntraline(v.a[c.oldStart+i], v.b[c.newStart+i], &budget)
			if len(ranges) > 0 {
				v.inline = append(v.inline, diffInline{c.oldStart + i, c.newStart + i, oldRows[i], newRows[i], ranges})
			}
		}
	}
}
