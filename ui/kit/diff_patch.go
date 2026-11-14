package kit

import (
	"fmt"
	"strconv"
	"strings"
)

func diffPatch(a, b []string, changes []diffRange, oldName, newName string) (string, []int) {
	return diffPatchContext(a, b, changes, oldName, newName, 3)
}

func diffPatchContext(a, b []string, changes []diffRange, oldName, newName string, context int) (string, []int) {
	context = max(0, context)
	if len(changes) == 0 {
		return "", nil
	}
	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", strconv.Quote("a/"+oldName), strconv.Quote("b/"+newName))
	row := 2
	positions := make([]int, len(changes))
	emit := func(prefix byte, line string) {
		out.WriteByte(prefix)
		out.WriteString(line)
		row++
		if !strings.HasSuffix(line, "\n") {
			out.WriteString("\n\\ No newline at end of file\n")
			row++
		}
	}
	for i := 0; i < len(changes); {
		j := i + 1
		first := changes[i]
		oa, na := max(0, first.oldStart-context), max(0, first.newStart-context)
		oz, nz := min(len(a), first.oldEnd+context), min(len(b), first.newEnd+context)
		for j < len(changes) && changes[j].oldStart-context <= oz {
			oz, nz = min(len(a), changes[j].oldEnd+context), min(len(b), changes[j].newEnd+context)
			j++
		}
		start := func(offset, count int) int {
			if count == 0 {
				return offset
			}
			return offset + 1
		}
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", start(oa, oz-oa), oz-oa, start(na, nz-na), nz-na)
		row++
		cursor := oa
		for k := i; k < j; k++ {
			c := changes[k]
			for _, line := range a[cursor:c.oldStart] {
				emit(' ', line)
			}
			positions[k] = row
			for _, line := range a[c.oldStart:c.oldEnd] {
				emit('-', line)
			}
			for _, line := range b[c.newStart:c.newEnd] {
				emit('+', line)
			}
			cursor = c.oldEnd
		}
		for _, line := range a[cursor:oz] {
			emit(' ', line)
		}
		i = j
	}
	return out.String(), positions
}
