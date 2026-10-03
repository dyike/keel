package kit

import (
	"math"
	"strconv"

	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// ItemsPerView sets the default number of equal-size slots in the viewport.
// One restores single-slide presentation when no item overrides remain.
// Nonpositive values are ignored.
func (v *CarouselView) ItemsPerView(count int) *CarouselView {
	if count > 0 {
		v.perView = count
		v.basis = 0
	}
	return v
}

// Basis sets the default fraction of the viewport occupied by each slot,
// including its share of spacing. Values in (0, 1] are accepted; zero restores
// ItemsPerView. ItemsPerView clears this default but preserves item overrides.
func (v *CarouselView) Basis(fraction float32) *CarouselView {
	if fraction >= 0 && fraction <= 1 && finiteNumber(float64(fraction)) {
		v.basis = fraction
	}
	return v
}

// ItemBasis overrides the viewport fraction for one zero-based item. Zero
// restores the default; invalid indices/fractions are ignored. Configuration
// changes preserve selection and do not emit OnChange.
func (v *CarouselView) ItemBasis(index int, fraction float32) *CarouselView {
	if index < 0 || index >= len(v.slides) || fraction < 0 || fraction > 1 || !finiteNumber(float64(fraction)) {
		return v
	}
	if fraction == 0 {
		delete(v.itemBasis, index)
	} else {
		if v.itemBasis == nil {
			v.itemBasis = make(map[int]float32)
		}
		v.itemBasis[index] = fraction
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
	defaultBasis := v.basis
	if defaultBasis == 0 {
		defaultBasis = 1 / float32(max(v.perView, 1))
	}
	minimumBasis := defaultBasis
	for _, fraction := range v.itemBasis {
		minimumBasis = min(minimumBasis, fraction)
	}
	gap := min(v.gap, viewport*minimumBasis)
	sizes := make([]float32, len(v.slides))
	for i := range sizes {
		fraction := defaultBasis
		if override, ok := v.itemBasis[i]; ok {
			fraction = override
		}
		sizes[i] = max(0, (viewport+gap)*fraction-gap)
	}
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
			item.H(el.Dp(sizes[i]))
		} else {
			item.W(el.Dp(sizes[i])).HFull()
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
		target := float32(0)
		for i := 0; i < v.current && i < len(sizes); i++ {
			target += float32(gtx.Dp(unit.Dp(sizes[i]))+gtx.Dp(unit.Dp(gap))) / scale
		}
		if v.vertical {
			stage.ScrollOffset(0, target)
		} else {
			stage.ScrollOffset(target, 0)
		}
		draw()
	})
}
