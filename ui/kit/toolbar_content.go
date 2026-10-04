package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"strconv"
)

func (v *ToolbarView) customWidth(it ToolbarItem) float32 {
	if it.Width > 0 && it.Width <= 4096 && finiteNumber(float64(it.Width)) {
		return it.Width
	}
	return 160
}

func (v *ToolbarView) renderCustomOverflow(cx *el.Context, id string, visible int) {
	index := v.customOpen
	if index < 0 {
		return
	}
	if v.disabled || index < visible || index >= len(v.items) || v.items[index].Disabled || v.items[index].Content == nil || v.items[index].Separator {
		v.customOpen = -1
		return
	}
	it := v.items[index]
	content := it.Content
	if it.OverflowContent != nil {
		content = it.OverflowContent
	}
	width, height := cx.ViewportSize()
	panel := floating(theme.ElevationMd).ID(id + "/custom-panel/" + strconv.Itoa(index)).Role("dialog").Name(it.Label).
		W(el.Dp(v.customWidth(it) + 2*theme.SpaceLg)).MaxW(el.Dp(max(0, width-16))).MaxH(el.Dp(max(0, height-16))).P(theme.SpaceLg).ScrollY().Items(el.Stretch)
	cx.Overlay(id+"/custom-overflow", el.Anchored(id, panel).Owner(id).Placement(el.Bottom, el.End).Modal().TrapFocus().OnDismiss(func() { v.customOpen = -1 }))
	panel.Child(content.Render(cx))
}
