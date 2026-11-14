package kit

import (
	"image/color"
	"strconv"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// sidebarHandleHit is the pointer target around the dividing line, in dp.
const sidebarHandleHit = 7

// Resizable lets the user drag the dividing line, or focus it and use the
// arrow keys (16dp a press, Home / End for the limits), to change the
// expanded width between min and max dp. A width chosen this way is explicit,
// like Width, so it turns off AutoWidth. The collapsed rail is not resizable.
func (v *SidebarView) Resizable(min, max float32) *SidebarView {
	if !finiteNumber(float64(min)) || !finiteNumber(float64(max)) || min <= 0 {
		return v
	}
	v.resizable, v.minWidth, v.maxWidth = true, min, max
	if v.maxWidth < v.minWidth {
		v.maxWidth = v.minWidth
	}
	return v
}

// OnResize is called with the new expanded width after the user resizes it.
func (v *SidebarView) OnResize(fn func(dp float32)) *SidebarView { v.onResize = fn; return v }

// CurrentWidth is the expanded width in dp, as last set by Width or a resize.
func (v *SidebarView) CurrentWidth() float32 { return v.width }

// SectionTitleColor sets the color of section headings, theme.Muted by
// default. Set it in Render to follow theme changes.
func (v *SidebarView) SectionTitleColor(c color.NRGBA) *SidebarView {
	v.titleColor = &c
	return v
}

func (v *SidebarView) resize(dp float32) {
	dp = min(max(dp, v.minWidth), v.maxWidth)
	if dp == v.width && !v.autoWidth {
		return
	}
	v.width, v.autoWidth = dp, false
	if v.onResize != nil {
		v.onResize(dp)
	}
}

// border is the dividing line, or with Resizable a wider separator that
// overlaps the line's neighbours without changing the layout.
func (v *SidebarView) border(cx *el.Context, base string, painted float32) el.Element {
	if !v.resizable || v.Collapsed() || v.disabled {
		v.resizeGrab, v.resizing = 0, false
		return el.Div().W(el.Dp(v.borderWidth)).NoShrink().Bg(theme.Border)
	}
	id := base + "/resize"
	sign := float32(1)
	if v.side == el.Right {
		sign = -1
	}
	line, c := v.borderWidth, theme.Border
	if v.resizing || cx.Hovered(id) || cx.Focused(id) {
		line, c = max(line, 2), theme.Primary
	}
	hit := max(float32(sidebarHandleHit), v.borderWidth)
	return el.Div().ID(id).Role("separator").Name(locale.Current().Resize).Value(strconv.Itoa(int(painted))).
		W(el.Dp(hit)).Mx(-(hit - v.borderWidth) / 2).NoShrink().Row().Justify(el.Center).Items(el.Stretch).
		Cursor(pointer.CursorColResize).Focusable(true).
		OnDrag(func(e el.DragEvent) {
			switch e.Kind {
			case el.DragStart:
				v.resizeGrab, v.resizing = e.X, true
				return
			case el.DragEnd:
				v.resizing = false
				if e.Canceled {
					return
				}
			}
			// The handle sits where the last paint put it, so the new width
			// is that width plus how far the pointer moved inside it.
			v.resize(painted + sign*(e.X-v.resizeGrab))
		}).
		OnKey(func(e el.KeyEvent) bool {
			d := float32(0)
			switch key.Name(e.Name) {
			case key.NameLeftArrow:
				d = -16 * sign
			case key.NameRightArrow:
				d = 16 * sign
			case key.NameHome:
				d = -1e6
			case key.NameEnd:
				d = 1e6
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.resize(painted + d)
			}
			return true
		}).
		Child(el.Div().W(el.Dp(line)).Bg(c))
}
