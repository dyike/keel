package plot

import (
	"fmt"
	"math"
)

// StackValue retains the source series/sample index. Missing samples are invalid.
type StackValue struct {
	Series, Index int
	Low, High     float64
	Valid         bool
}

// Stack lays out series-major values. Positive and negative totals accumulate
// separately from zero. NaN represents a missing sample; infinities and total
// overflow fail atomically. Input slices are never retained.
func Stack(series [][]float64) ([][]StackValue, error) {
	n := 0
	for _, s := range series {
		n = max(n, len(s))
	}
	positive, negative := make([]float64, n), make([]float64, n)
	out := make([][]StackValue, len(series))
	for i, s := range series {
		out[i] = make([]StackValue, n)
		for j := 0; j < n; j++ {
			v := StackValue{Series: i, Index: j}
			if j >= len(s) || math.IsNaN(s[j]) {
				out[i][j] = v
				continue
			}
			x := s[j]
			if !finite(x) {
				return nil, fmt.Errorf("plot stack: non-finite value at %d/%d", i, j)
			}
			base := positive[j]
			if x < 0 {
				base = negative[j]
			}
			end := base + x
			if !finite(end) {
				return nil, fmt.Errorf("plot stack: total overflow at %d/%d", i, j)
			}
			v.Low, v.High, v.Valid = base, end, true
			out[i][j] = v
			if x < 0 {
				negative[j] = end
			} else {
				positive[j] = end
			}
		}
	}
	return out, nil
}

// PieSlice retains the input index and value. Angles are radians, with zero on
// positive X; positive sweeps turn clockwise in screen coordinates.
type PieSlice struct {
	Index             int
	Value, Start, End float64
}

// Pie lays out positive values over a signed sweep of at most one revolution.
// Zeros are omitted. Negative/non-finite values fail atomically. Padding is a
// nonnegative angular gap, capped at the available angle per positive item.
func Pie(values []float64, start, sweep, padding float64) ([]PieSlice, error) {
	if !finite(start) || !finite(sweep) || !finite(padding) || math.Abs(sweep) > 2*math.Pi || padding < 0 {
		return nil, fmt.Errorf("plot pie: invalid angles")
	}
	peak := 0.0
	n := 0
	for i, v := range values {
		if !finite(v) || v < 0 {
			return nil, fmt.Errorf("plot pie: invalid value at %d", i)
		}
		if v > 0 {
			n++
			peak = max(peak, v)
		}
	}
	if n == 0 || sweep == 0 {
		return nil, nil
	}
	total := 0.0
	for _, v := range values {
		total += v / peak
	}
	gap := min(padding, math.Abs(sweep)/float64(n))
	available := max(0, math.Abs(sweep)-float64(n)*gap)
	sign := math.Copysign(1, sweep)
	at := math.Mod(start, 2*math.Pi)
	out := make([]PieSlice, 0, n)
	for i, v := range values {
		if v == 0 {
			continue
		}
		a := at + sign*gap/2
		span := available * (v / peak) / total
		z := a + sign*span
		out = append(out, PieSlice{i, v, a, z})
		at += sign * (span + gap)
	}
	return out, nil
}
