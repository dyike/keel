package motion

import "github.com/dyike/keel/ui/internal/animation"

// Easing maps normalized time to animation progress.
type Easing = animation.Easing

var (
	Linear     = animation.Linear
	EaseIn     = animation.EaseIn
	EaseOut    = animation.EaseOut
	EaseInOut  = animation.EaseInOut
	CircIn     = animation.CircIn
	CircOut    = animation.CircOut
	CircInOut  = animation.CircInOut
	BackIn     = animation.BackIn
	BackOut    = animation.BackOut
	BackInOut  = animation.BackInOut
	Anticipate = animation.Anticipate
)

func CubicBezier(x1, y1, x2, y2 float32) Easing { return animation.CubicBezier(x1, y1, x2, y2) }
func clamp01(x float32) float32                 { return min(max(x, 0), 1) }
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
