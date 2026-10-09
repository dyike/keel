package kit

import (
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Footer sets persistent content below the scrolling candidates, including
// loading, error and empty states. Nil removes it. Actions manage selection
// through the public setters; interacting with the footer does not choose a row.
func (v *ComboboxView) Footer(view el.View) *ComboboxView {
	v.footer = view
	return v
}

func (v *ComboboxView) renderFooter(cx *el.Context, id string, height float32) (el.Element, float32) {
	if v.footer == nil {
		v.footerMeasured = false
		return nil, 0
	}
	limit := max(1, (height-80)/3)
	footer := el.Div().ID(id + "/footer").MaxH(el.Dp(limit)).ScrollY().P(theme.SpaceSm).Child(v.footer.Render(cx))
	footer.Decorate(func(gtx core.C, draw func()) {
		_, actual := cx.LayoutSize(footer)
		if !v.footerMeasured || v.footerHeight != actual {
			v.footerHeight = actual
			v.footerMeasured = true
			gtx.Execute(op.InvalidateCmd{})
		}
		draw()
	})
	reserve := limit
	if v.footerMeasured {
		reserve = min(limit, v.footerHeight)
	}
	return footer, reserve
}
