package kit

import (
	"math"
	"strconv"

	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// ItemsPerView sets the default number of equal-size slots in the viewport.
// One restores single-slide presentation when no item overrides remain.
// Nonpositive values are ignored.
func (v *CarouselView) ItemsPerView(count int) *CarouselView {
	if count > 0 {
		if v.perView != count || v.basis != 0 {
			v.drag.active = false
		}
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
		if v.basis != fraction {
			v.drag.active = false
		}
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
	if v.itemBasis[index] != fraction {
		v.drag.active = false
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

// ItemSize sets the main-axis size in dp for one zero-based item, excluding
// gaps. It overrides ItemBasis and the default fraction, and may exceed the
// viewport. Zero restores the proportional size. Invalid inputs are ignored.
// Resizing or changing orientation preserves the dp value and selection.
func (v *CarouselView) ItemSize(index int, dp float32) *CarouselView {
	if index < 0 || index >= len(v.slides) || dp < 0 || !finiteNumber(float64(dp)) {
		return v
	}
	if v.itemSizes[index] != dp {
		v.drag.active = false
	}
	if dp == 0 {
		delete(v.itemSizes, index)
	} else {
		if v.itemSizes == nil {
			v.itemSizes = make(map[int]float32)
		}
		v.itemSizes[index] = dp
	}
	return v
}

// Gap sets the spacing in dp between items in multi-item mode. Invalid values
// are ignored. The effective gap shrinks in viewports too small to fit it.
func (v *CarouselView) Gap(dp float32) *CarouselView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		if v.gap != dp {
			v.drag.active = false
		}
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
		if fixed, ok := v.itemSizes[i]; ok {
			sizes[i] = fixed
		}
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

	scale := cx.PixelScale()
	points := make([]float32, len(sizes))
	total := float32(0)
	for i, size := range sizes {
		points[i] = total
		total += float32(math.Round(float64(size*scale))) / scale
		if i+1 < len(sizes) {
			total += float32(math.Round(float64(gap*scale))) / scale
		}
	}
	maximum := max(0, total-viewport)
	for i := range points {
		points[i] = min(points[i], maximum)
	}
	geometry := &carouselGeometry{points: points, maximum: maximum, vertical: v.vertical}
	if v.draggable {
		stage.OnDrag(func(e el.DragEvent) { v.handleDrag(e, geometry) })
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
		target := float32(0)
		if v.current < len(points) {
			target = points[v.current]
		}
		if v.drag.active {
			if !gtx.Enabled() || v.drag.vertical != v.vertical || v.drag.maximum != maximum {
				v.drag.active = false
			} else {
				target = v.drag.offset
			}
		}
		if v.vertical {
			stage.ScrollOffset(0, target)
		} else {
			stage.ScrollOffset(target, 0)
		}
		draw()
	})
}
