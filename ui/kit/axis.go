package kit

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// axisFraction avoids overflowing the subtraction for opposite extreme values.
func axisFraction(x, lo, hi float64) float64 {
	scale := max(math.Abs(lo), math.Abs(hi))
	if scale == 0 || hi == lo {
		return 0.5
	}
	return (x/scale - lo/scale) / (hi/scale - lo/scale)
}

// niceTicks covers [lo, hi] with about n round steps (1, 2 or 5 × 10^k) and
// returns the ticks, first <= lo and last >= hi.
func niceTicks(lo, hi float64, n int) []float64 {
	if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		lo, hi = 0, 1
	}
	if hi < lo {
		lo, hi = hi, lo
	}
	if hi == lo {
		if lo == 0 {
			hi = 1
		} else {
			lo, hi = lo-math.Abs(lo)/2, hi+math.Abs(hi)/2
		}
	}
	lo = max(-math.MaxFloat64, lo)
	hi = min(math.MaxFloat64, hi)
	if lo == hi {
		if lo > 0 {
			lo = 0
		} else {
			hi = 1
		}
	}
	// Of the round steps (1, 2 or 5 × 10^k) near the even split, take the one
	// whose tick count is closest to n; ties go to the larger step.
	raw := (hi - lo) / float64(max(n, 1))
	if raw == 0 || !finiteNumber(raw) {
		return []float64{lo, hi}
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	if mag == 0 || !finiteNumber(mag) {
		return []float64{lo, hi}
	}
	step, best := mag, math.MaxFloat64
	for _, m := range []float64{0.5, 1, 2, 5, 10} {
		s := m * mag
		count := math.Ceil(hi/s-1e-9) - math.Floor(lo/s+1e-9)
		if d := math.Abs(count - float64(n)); d <= best {
			step, best = s, d
		}
	}
	var out []float64
	for v := math.Floor(lo/step) * step; v <= hi+step*1e-9; v += step {
		out = append(out, math.Round(v/step)*step)
		if len(out) > 50 {
			break
		}
	}
	if len(out) == 0 {
		return []float64{lo, hi}
	}
	if out[len(out)-1] < hi {
		out = append(out, out[len(out)-1]+step)
	}
	for _, t := range out {
		if !finiteNumber(t) {
			return []float64{lo, hi}
		}
	}
	if len(out) < 2 || out[0] >= out[len(out)-1] {
		return []float64{lo, hi}
	}
	return out
}

// formatNumber writes v with thousands separators and at most two decimals.
func formatNumber(v float64) string {
	if !finiteNumber(v) {
		return "—"
	}
	if a := math.Abs(v); a >= 1e12 || a > 0 && a < .01 {
		return strconv.FormatFloat(v, 'g', 4, 64)
	}
	neg := v < 0
	s := strconv.FormatFloat(math.Abs(v), 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	whole, frac, _ := strings.Cut(s, ".")
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	if frac != "" {
		whole += "." + frac
	}
	if neg && whole != "0" {
		whole = "-" + whole
	}
	return whole
}

func fillRect(gtx core.C, r image.Rectangle, c color.NRGBA) {
	paint.FillShape(gtx.Ops, c, clip.Rect(r).Op())
}

// strokePath draws a polyline of width w px with round joins.
func strokePath(gtx core.C, pts []f32.Point, w float32, c color.NRGBA) {
	if len(pts) < 2 {
		return
	}
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(pts[0])
	for _, q := range pts[1:] {
		p.LineTo(q)
	}
	paint.FillShape(gtx.Ops, c, clip.Stroke{Path: p.End(), Width: w}.Op())
}

// dot draws a filled circle of radius r with a ring of the surface color.
func dot(gtx core.C, at f32.Point, r, ring float32, fill, surface color.NRGBA) {
	circle := func(r float32, c color.NRGBA) {
		rect := image.Rect(int(at.X-r), int(at.Y-r), int(at.X+r+0.5), int(at.Y+r+0.5))
		paint.FillShape(gtx.Ops, c, clip.Ellipse(rect).Op(gtx.Ops))
	}
	circle(r+ring, surface)
	circle(r, fill)
}

// legendItem is a series key that toggles the series: swatch, then name in
// text color. A hidden series keeps an outlined swatch and a muted name, so
// the key stays readable without leaning on its color.
func legendItem(id, name string, c color.NRGBA, radius float32, shown bool, toggle func()) *el.DivEl {
	swatch := el.Div().Size(el.Dp(10)).Rounded(radius).NoShrink()
	fg := theme.Text
	if shown {
		swatch.Bg(c)
	} else {
		swatch.Border(1.5, c)
		fg = theme.Muted
	}
	return el.Div().ID(id).Role("toggle").Name(name).Selected(shown).
		Row().Items(el.Center).Gap(6).H(el.Dp(26)).Px(8).Rounded(theme.RadiusMd).Border(1, color.NRGBA{}).TextSize(theme.TextMd).TextColor(fg).
		Focusable(true).CursorPointer().OnClick(toggle).
		Hover(func(s *el.Style) { s.Bg(theme.Subtle) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		Child(swatch, el.Text(name))
}
