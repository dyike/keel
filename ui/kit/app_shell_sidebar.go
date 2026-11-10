package kit

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"strconv"
)

// HideSidebarOnCollapse removes navigation and its footer when collapsed,
// letting content use the full width. The title-bar toggle stays available.
// False (the default) keeps the compact icon rail.
func (v *AppShellView) HideSidebarOnCollapse(hide bool) *AppShellView {
	v.hideSidebarOnCollapse = hide
	return v
}

// SidebarSize returns the expanded sidebar width, including after a drag.
// Custom sidebar views should use it rather than keeping a separate width.
func (v *AppShellView) SidebarSize() float32 { return v.nav.width }

// SidebarLimits sets the expanded width range in dp (200..420 by default).
func (v *AppShellView) SidebarLimits(minimum, maximum float32) *AppShellView {
	if finiteNumber(float64(minimum)) && finiteNumber(float64(maximum)) && minimum >= 120 && maximum >= minimum {
		v.sidebarMin, v.sidebarMax = minimum, maximum
	}
	return v
}
func (v *AppShellView) clampSidebar(width float32) float32 {
	upper := v.sidebarMax
	if v.sidebarAvailable > 0 {
		upper = min(upper, max(v.sidebarMin, v.sidebarAvailable))
	}
	return max(v.sidebarMin, min(upper, width))
}
func (v *AppShellView) sidebarToggle(cx *el.Context) el.Element {
	name := locale.Current().CollapseSidebar
	if v.Collapsed() {
		name = locale.Current().ExpandSidebar
	}
	// Standard sidebar glyph: a window outline with a narrow navigation pane.
	glyph := el.Div().W(el.Dp(18)).H(el.Dp(14)).Rounded(3).Border(1, theme.Text).Child(
		el.Div().Absolute().Left(5).Top(1).Bottom(1).W(el.Dp(1)).Bg(theme.Text))
	return el.Div().ID(autoID("app-shell-toggle", v)).Role("button").Name(name).Focusable(true).CursorPointer().
		Size(el.Dp(32)).Rounded(6).Center().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
		FocusStyle(func(s *el.Style) { s.Bg(theme.SubtleHover) }).OnClick(func() { v.SetCollapsed(!v.Collapsed()) }).Child(glyph)
}
func (v *AppShellView) sidebarDivider(cx *el.Context) el.Element {
	width := v.nav.width
	if v.Collapsed() {
		width = 56
		v.resizing = false
	}
	handle := el.Div().ID(autoID("app-shell-divider", v)).Absolute().Left(width - 3).Top(0).Bottom(0).W(el.Dp(6)).Items(el.Center).
		Child(el.Div().W(el.Dp(1)).HFull().Bg(theme.Border))
	if v.Collapsed() {
		return handle.Top(AppShellTitleBarHeight)
	}
	v.resizePainted = width
	return handle.Role("separator").Name(locale.Current().Resize).Value(strconv.Itoa(int(width))).Focusable(true).Cursor(pointer.CursorColResize).
		OnDrag(func(e el.DragEvent) {
			switch e.Kind {
			case el.DragStart:
				v.resizeGrab = e.X
				v.resizing = true
			case el.DragMove:
				if v.resizing {
					v.nav.width = v.clampSidebar(v.resizePainted + e.X - v.resizeGrab)
				}
			case el.DragEnd:
				v.resizing = false
			}
		}).OnKey(func(e el.KeyEvent) bool {
		if e.State != el.KeyPress {
			return false
		}
		n := v.nav.width
		switch key.Name(e.Name) {
		case key.NameLeftArrow:
			n -= 16
		case key.NameRightArrow:
			n += 16
		case key.NameHome:
			n = v.sidebarMin
		case key.NameEnd:
			n = v.sidebarMax
		default:
			return false
		}
		v.nav.width = v.clampSidebar(n)
		return true
	})
}
