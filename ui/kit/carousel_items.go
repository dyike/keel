package kit

import (
	"math"
	"strconv"

	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// ItemsPerView sets the number of equal-size slots in the viewport. One restores
// the default single-slide presentation. Nonpositive values are ignored.
func (v *CarouselView) ItemsPerView(count int) *CarouselView {
	if count > 0 {
		v.perView = count
	}
	return v
}

// Gap sets the spacing in dp between items in multi-item mode. Invalid values
// are ignored. The effective gap shrinks in viewports too small to fit it.
func (v *CarouselView) Gap(dp float32) *CarouselView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = dp
	}
	return v
}

func (v *CarouselView) multiStage(cx *el.Context, stage *el.DivEl, id string) {
	viewport, height := cx.LastSize(id)
	if v.vertical {
		viewport = height
	}
	if viewport <= 0 {
		viewport, _ = cx.ViewportSize()
		if v.vertical {
			viewport = v.height
		}
	}
	count := float32(v.perView)
	gap := min(v.gap, viewport/count)
	itemSize := max(0, (viewport-gap*(count-1))/count)
	stage.Justify(el.Start).Gap(gap)
	if v.vertical {
		stage.Col().ScrollY()
	} else {
		stage.Row().ScrollX()
	}
	for i, slide := range v.slides {
		item := el.Div().ID(id + "/item/" + strconv.Itoa(i)).NoShrink().Items(el.Stretch).
			Role("group").Name(strconv.Itoa(i+1) + "/" + strconv.Itoa(len(v.slides)))
		if v.vertical {
			item.H(el.Dp(itemSize))
		} else {
			item.W(el.Dp(itemSize)).HFull()
		}
		if slide != nil {
			item.Child(slide.Render(cx))
		}
		stage.Child(item)
	}

	// Each cell and gap is independently rounded by the layout engine.
	stage.Decorate(func(gtx core.C, draw func()) {
		actual, height := cx.LayoutSize(stage)
		if v.vertical {
			actual = height
		}
		if math.Abs(float64(actual-viewport)) > 0.01 {
			gtx.Execute(op.InvalidateCmd{})
		}
		scale := gtx.Metric.PxPerDp
		if scale <= 0 {
			scale = 1
		}
		step := float32(gtx.Dp(unit.Dp(itemSize))+gtx.Dp(unit.Dp(gap))) / scale
		target := float32(v.current) * step
		if v.vertical {
			stage.ScrollOffset(0, target)
		} else {
			stage.ScrollOffset(target, 0)
		}
		draw()
	})
}
