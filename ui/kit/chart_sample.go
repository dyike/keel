package kit

import (
	"gioui.org/f32"
	"slices"
)

// chartSegments keeps gaps as separate paths and keeps first/last/min/max in
// each pixel bucket. Narrow peaks survive while dense runs emit O(width) points.
func chartSegments(values []float64, n int, band float32, y func(float64) float32, pixels int) [][]f32.Point {
	n = min(n, len(values))
	step := max(1, n/max(pixels, 1))
	var out [][]f32.Point
	for start := 0; start < n; {
		if !finiteNumber(values[start]) {
			start++
			continue
		}
		end := start + 1
		for end < n && finiteNumber(values[end]) {
			end++
		}
		points := make([]f32.Point, 0, min(end-start, pixels*4+4))
		for a := start; a < end; a += step {
			b := min(a+step, end)
			low, high := a, a
			for j := a + 1; j < b; j++ {
				if values[j] < values[low] {
					low = j
				}
				if values[j] > values[high] {
					high = j
				}
			}
			indices := []int{a, low, high, b - 1}
			slices.Sort(indices)
			for k, j := range indices {
				if k == 0 || j != indices[k-1] {
					points = append(points, f32.Pt((float32(j)+.5)*band, y(values[j])))
				}
			}
		}
		out = append(out, points)
		start = end
	}
	return out
}
