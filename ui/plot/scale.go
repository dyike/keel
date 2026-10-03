// Package plot provides scales, data layouts and immediate-mode chart drawing.
// Shapes use local pixel coordinates; applications own data, interaction and
// accessible descriptions. kit.Plot and kit.Chart provide complete chart views.
package plot

import (
	"math"
	"slices"
)

func finite(x float64) bool               { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func interpolate(a, b, t float64) float64 { return a*(1-t) + b*t }
func fraction(x, a, b float64) float64 {
	if a != b {
		delta := b - a
		if finite(delta) && delta != 0 {
			t := (x - a) / delta
			if finite(t) {
				return t
			}
		}
	}
	scale := max(math.Abs(x), math.Abs(a), math.Abs(b))
	if scale == 0 || a == b {
		return .5
	}
	return (x/scale - a/scale) / (b/scale - a/scale)
}

// ScaleLinear supports reversed domains/ranges. Equal domains map to the range
// midpoint. Invalid input or an unrepresentable extrapolation returns false.
type ScaleLinear struct {
	Domain, Range [2]float64
	Clamp         bool
}

func (s ScaleLinear) Map(x float64) (float64, bool) {
	a, b := s.Domain[0], s.Domain[1]
	c, d := s.Range[0], s.Range[1]
	if !finite(x) || !finite(a) || !finite(b) || !finite(c) || !finite(d) {
		return 0, false
	}
	t := fraction(x, a, b)
	if s.Clamp {
		t = min(1, max(0, t))
	}
	y := interpolate(c, d, t)
	return y, finite(y)
}
func (s ScaleLinear) Invert(y float64) (float64, bool) {
	return (ScaleLinear{Domain: s.Range, Range: s.Domain, Clamp: s.Clamp}).Map(y)
}

// Ticks returns evenly spaced domain values, including endpoints (2..1000).
// It preserves descending order and returns one tick for a constant domain.
func (s ScaleLinear) Ticks(count int) []float64 {
	a, b := s.Domain[0], s.Domain[1]
	if !finite(a) || !finite(b) {
		return nil
	}
	if a == b {
		return []float64{a}
	}
	count = min(1000, max(2, count))
	out := make([]float64, count)
	for i := range out {
		out[i] = interpolate(a, b, float64(i)/float64(count-1))
	}
	return out
}

// ScaleBand owns a unique, insertion-ordered categorical domain.
// Map returns the low coordinate of each band even for descending ranges.
type ScaleBand[T comparable] struct {
	domain              []T
	index               map[T]int
	extent              [2]float64
	inner, outer, align float64
}

func NewBand[T comparable](domain []T, extent [2]float64) ScaleBand[T] {
	s := ScaleBand[T]{index: map[T]int{}, extent: extent, align: .5}
	for _, x := range domain {
		if _, ok := s.index[x]; !ok {
			s.index[x] = len(s.domain)
			s.domain = append(s.domain, x)
		}
	}
	return s
}
func (s ScaleBand[T]) Domain() []T { return slices.Clone(s.domain) }
func (s ScaleBand[T]) Padding(inner, outer float64) ScaleBand[T] {
	if finite(inner) {
		s.inner = min(1, max(0, inner))
	}
	if finite(outer) {
		s.outer = max(0, outer)
	}
	return s
}
func (s ScaleBand[T]) Align(f float64) ScaleBand[T] {
	if finite(f) {
		s.align = min(1, max(0, f))
	}
	return s
}
func (s ScaleBand[T]) metrics() (start, step, width float64, ok bool) {
	a, b := min(s.extent[0], s.extent[1]), max(s.extent[0], s.extent[1])
	n := len(s.domain)
	if n == 0 || !finite(a) || !finite(b) || !finite(b-a) {
		return 0, 0, 0, false
	}
	step = (b - a) / max(1, float64(n)-s.inner+2*s.outer)
	width = step * (1 - s.inner)
	start = a + ((b-a)-step*(float64(n)-s.inner))*s.align
	return start, step, width, finite(start) && finite(step) && finite(width)
}
func (s ScaleBand[T]) Bandwidth() float64 { _, _, w, _ := s.metrics(); return w }
func (s ScaleBand[T]) Map(x T) (float64, bool) {
	i, exists := s.index[x]
	start, step, _, ok := s.metrics()
	if !exists || !ok {
		return 0, false
	}
	if s.extent[1] < s.extent[0] {
		i = len(s.domain) - 1 - i
	}
	return start + float64(i)*step, true
}

// ScalePoint places categories on points; a singleton is centered.
type ScalePoint[T comparable] struct{ band ScaleBand[T] }

func NewPoint[T comparable](domain []T, extent [2]float64) ScalePoint[T] {
	return ScalePoint[T]{NewBand(domain, extent).Padding(1, 0)}
}
func (s ScalePoint[T]) Padding(p float64) ScalePoint[T] { s.band = s.band.Padding(1, p); return s }
func (s ScalePoint[T]) Map(x T) (float64, bool)         { return s.band.Map(x) }
func (s ScalePoint[T]) Domain() []T                     { return s.band.Domain() }

// ScaleOrdinal cycles a discrete range across a fixed domain. Missing keys and
// empty ranges return false; Map never extends the domain.
type ScaleOrdinal[K comparable, V any] struct {
	index  map[K]int
	values []V
}

func NewOrdinal[K comparable, V any](domain []K, values []V) ScaleOrdinal[K, V] {
	s := ScaleOrdinal[K, V]{index: map[K]int{}, values: slices.Clone(values)}
	for _, x := range domain {
		if _, ok := s.index[x]; !ok {
			s.index[x] = len(s.index)
		}
	}
	return s
}
func (s ScaleOrdinal[K, V]) Map(x K) (V, bool) {
	i, ok := s.index[x]
	if !ok || len(s.values) == 0 {
		var zero V
		return zero, false
	}
	return s.values[i%len(s.values)], true
}
