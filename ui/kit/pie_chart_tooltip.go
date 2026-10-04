package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

// PieChartTooltip is a value snapshot of the hovered visible sector. Share is
// a fraction in [0,1], recalculated after legend visibility changes.
type PieChartTooltip struct {
	Index                          int
	Slice                          PieSlice
	Share                          float64
	FormattedValue, FormattedShare string
	Color                          color.NRGBA
}

// TooltipContent adds a display-only floating tooltip. Nil restores the default
// bottom readout alone; a callback returning nil suppresses the floating panel.
// The bottom readout and accessible table remain available. Do not mutate chart
// data from this render callback.
func (v *PieChartView) TooltipContent(fn func(*el.Context, PieChartTooltip) el.Element) *PieChartView {
	v.tooltipContent = fn
	return v
}
func (v *PieChartView) renderTooltip(cx *el.Context, id string, parts []float64, plot *el.DivEl) {
	i := v.hover
	if v.tooltipContent == nil || v.disabled || i < 0 || i >= len(v.data) || parts[i] <= 0 {
		return
	}
	content := v.tooltipContent(cx, PieChartTooltip{Index: i, Slice: v.data[i], Share: parts[i], FormattedValue: v.valueText(i), FormattedShare: pieShare(parts[i]), Color: theme.Chart[i%len(theme.Chart)]})
	if content == nil {
		return
	}
	plot.Child(el.Div().ID(id + "/point").Absolute().Left(v.pointerX).Top(v.pointerY).Size(el.Dp(1)))
	width, height := cx.ViewportSize()
	panel := floating(theme.ElevationSm).Role("tooltip").Disabled(true).P(theme.SpaceMd).MaxW(el.Dp(max(0, min(320, width-16)))).MaxH(el.Dp(max(0, height-16))).Items(el.Stretch).Child(content)
	cx.Overlay(id+"/tooltip", el.Anchored(id+"/point", panel).Owner(id).Placement(el.Bottom, el.Start).Offset(12))
}
