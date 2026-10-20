package motion

import "math"

// Easing maps a normalized time to a normalized animation progress.
type Easing func(float32) float32

var (
	Linear    Easing = func(x float32) float32 { return clamp01(x) }
	EaseIn           = CubicBezier(.42, 0, 1, 1)
	EaseOut          = CubicBezier(0, 0, .58, 1)
	EaseInOut        = CubicBezier(.42, 0, .58, 1)
	CircIn    Easing = func(x float32) float32 {
		x = clamp01(x)
		return 1 - float32(math.Sqrt(float64(max(0, 1-x*x))))
	}
	CircOut Easing = func(x float32) float32 {
		x = clamp01(x) - 1
		return float32(math.Sqrt(float64(max(0, 1-x*x))))
	}
	CircInOut Easing = func(x float32) float32 {
		x = clamp01(x) * 2
		if x < 1 {
			return (1 - float32(math.Sqrt(float64(max(0, 1-x*x))))) / 2
		}
		x -= 2
		return (float32(math.Sqrt(float64(max(0, 1-x*x)))) + 1) / 2
	}
	BackIn Easing = func(x float32) float32 {
		x = clamp01(x)
		const c = float32(1.70158)
		return (c+1)*x*x*x - c*x*x
	}
	BackOut Easing = func(x float32) float32 {
		x = clamp01(x) - 1
		const c = float32(1.70158)
		return 1 + (c+1)*x*x*x + c*x*x
	}
	BackInOut Easing = func(x float32) float32 {
		x = clamp01(x) * 2
		const c = float32(1.70158 * 1.525)
		if x < 1 {
			return x * x * ((c+1)*x - c) / 2
		}
		x -= 2
		return (x*x*((c+1)*x+c) + 2) / 2
	}
	Anticipate Easing = func(x float32) float32 {
		x = clamp01(x)
		if x < .5 {
			return .5 * BackIn(x*2)
		}
		return .5 * (BackOut(x*2-1) + 1)
	}
)

// CubicBezier returns an easing curve with CSS-compatible control points.
// The x coordinates are clamped to 0..1 so the curve remains a function.
func CubicBezier(x1, y1, x2, y2 float32) Easing {
	x1, x2 = clamp01(x1), clamp01(x2)
	return func(x float32) float32 {
		x = clamp01(x)
		if x == 0 || x == 1 {
			return x
		}
		t := x
		for range 8 {
			err := bezier(t, x1, x2) - x
			d := bezierDerivative(t, x1, x2)
			if abs(d) < 1e-6 {
				break
			}
			next := t - err/d
			if next < 0 || next > 1 {
				break
			}
			t = next
		}
		lo, hi := float32(0), float32(1)
		for range 12 {
			bx := bezier(t, x1, x2)
			if abs(bx-x) < 1e-6 {
				break
			}
			if bx < x {
				lo = t
			} else {
				hi = t
			}
			t = (lo + hi) / 2
		}
		return bezier(t, y1, y2)
	}
}

func bezier(t, p1, p2 float32) float32 {
	u := 1 - t
	return 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t
}

func bezierDerivative(t, p1, p2 float32) float32 {
	u := 1 - t
	return 3*u*u*p1 + 6*u*t*(p2-p1) + 3*t*t*(1-p2)
}

func clamp01(x float32) float32 { return min(max(x, 0), 1) }
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
