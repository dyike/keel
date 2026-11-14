package kit

import (
	"sort"
	"strings"
)

// diffRange uses zero-based, half-open line ranges. Line tokens retain their
// terminators, so a final newline or CRLF change is not silently discarded.
type diffRange struct{ oldStart, oldEnd, newStart, newEnd int }
type diffMatch struct{ old, new int }

func diffLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffText uses unique-line anchors and bounded LCS for repeated-line regions.
// The budget bounds quadratic work across the entire comparison. When no
// anchors remain and the budget is insufficient, that region is one change;
// reconstruction is exact, but the result need not be a minimal edit script.
func diffText(before, after string) ([]string, []string, []diffRange) {
	a, b := diffLines(before), diffLines(after)
	budget := 1_000_000
	return a, b, diffTokens(a, b, &budget)
}

func diffTokens(a, b []string, budget *int) []diffRange {
	var matches []diffMatch
	var walk func(int, int, int, int)
	walk = func(ai, ae, bi, be int) {
		for ai < ae && bi < be && a[ai] == b[bi] {
			matches = append(matches, diffMatch{ai, bi})
			ai++
			bi++
		}
		suffix := 0
		for ai < ae && bi < be && a[ae-1] == b[be-1] {
			ae--
			be--
			suffix++
		}
		n, m := ae-ai, be-bi
		if n > 0 && m > 0 {
			if n <= *budget/m {
				*budget -= n * m
				stride := m + 1
				table := make([]int, (n+1)*stride)
				for i := n - 1; i >= 0; i-- {
					for j := m - 1; j >= 0; j-- {
						if a[ai+i] == b[bi+j] {
							table[i*stride+j] = 1 + table[(i+1)*stride+j+1]
						} else {
							table[i*stride+j] = max(table[(i+1)*stride+j], table[i*stride+j+1])
						}
					}
				}
				for i, j := 0, 0; i < n && j < m; {
					if a[ai+i] == b[bi+j] {
						matches = append(matches, diffMatch{ai + i, bi + j})
						i++
						j++
					} else if table[(i+1)*stride+j] >= table[i*stride+j+1] {
						i++
					} else {
						j++
					}
				}
			} else {
				anchors := diffAnchors(a[ai:ae], b[bi:be])
				leftA, leftB := ai, bi
				for _, p := range anchors {
					x, y := ai+p.old, bi+p.new
					walk(leftA, x, leftB, y)
					matches = append(matches, diffMatch{x, y})
					leftA, leftB = x+1, y+1
				}
				if len(anchors) > 0 {
					walk(leftA, ae, leftB, be)
				}
			}
		}
		for i := 0; i < suffix; i++ {
			matches = append(matches, diffMatch{ae + i, be + i})
		}
	}
	walk(0, len(a), 0, len(b))
	var changes []diffRange
	ai, bi := 0, 0
	for _, p := range matches {
		if ai != p.old || bi != p.new {
			changes = append(changes, diffRange{ai, p.old, bi, p.new})
		}
		ai, bi = p.old+1, p.new+1
	}
	if ai != len(a) || bi != len(b) {
		changes = append(changes, diffRange{ai, len(a), bi, len(b)})
	}
	return changes
}

func diffAnchors(a, b []string) []diffMatch {
	unique := func(lines []string) map[string]int {
		out := make(map[string]int, len(lines))
		for i, s := range lines {
			if _, ok := out[s]; ok {
				out[s] = -1
			} else {
				out[s] = i
			}
		}
		return out
	}
	am, bm := unique(a), unique(b)
	var candidates []diffMatch
	for i, s := range a {
		if am[s] == i {
			if j, ok := bm[s]; ok && j >= 0 {
				candidates = append(candidates, diffMatch{i, j})
			}
		}
	}
	tails := make([]int, 0, len(candidates))
	prev := make([]int, len(candidates))
	for i, p := range candidates {
		at := sort.Search(len(tails), func(j int) bool { return candidates[tails[j]].new >= p.new })
		prev[i] = -1
		if at > 0 {
			prev[i] = tails[at-1]
		}
		if at == len(tails) {
			tails = append(tails, i)
		} else {
			tails[at] = i
		}
	}
	result := make([]diffMatch, len(tails))
	if len(tails) > 0 {
		i := tails[len(tails)-1]
		for j := len(result) - 1; j >= 0; j-- {
			result[j] = candidates[i]
			i = prev[i]
		}
	}
	return result
}
