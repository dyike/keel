package kit

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AppShellTitleBarHeight is the shared height of the sidebar toolbar and title bar.
const AppShellTitleBarHeight = 56

// Sidebar replaces built-in page navigation. Nil restores it. The view owns its
// scrolling and can query Collapsed to adapt to the icon rail.
func (v *AppShellView) Sidebar(view el.View) *AppShellView { v.sidebar = view; return v }

// SidebarHeader supplies toolbar controls beside the macOS window buttons.
func (v *AppShellView) SidebarHeader(view el.View) *AppShellView { v.sidebarHeader = view; return v }

// Header replaces the title and description with custom interactive content.
// It is excluded from window dragging; the remaining blank area stays draggable.
func (v *AppShellView) Header(view el.View) *AppShellView { v.header = view; return v }

// Content replaces page content with a retained full-size view. The view owns
// padding and scrolling. Nil returns to the selected page.
func (v *AppShellView) Content(view el.View) *AppShellView {
	if view == nil {
		v.content = nil
	} else {
		v.content = el.Root(view)
	}
	return v
}

// ContentFooter pins custom content, such as a message composer, below the body.
func (v *AppShellView) ContentFooter(view el.View) *AppShellView { v.contentFooter = view; return v }

func (v *AppShellView) navigation(cx *el.Context) el.Element {
	width := v.nav.width
	if v.Collapsed() {
		width = 56
	}
	side := el.Div().W(el.Dp(width)).NoShrink().Items(el.Stretch).Bg(theme.Bg)
	nav := el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch)
	if v.sidebar != nil {
		nav.Child(v.sidebar.Render(cx))
	} else {
		nav.Row().Child(v.nav.Render(cx))
	}
	side.Child(nav)
	if v.footer != nil {
		side.Child(el.Div().NoShrink().Items(el.Stretch).Child(v.footer.Render(cx)))
	}
	return side
}

func (v *AppShellView) titlebar(cx *el.Context, p *shellPage) el.Element {
	title, description := v.title, ""
	if p != nil {
		title, description = p.spec.Title, p.spec.Description
	}
	bar := el.Div().Grow().W(el.Dp(0)).Role("banner").Name(title).MinH(el.Dp(AppShellTitleBarHeight)).NoShrink().Row().Items(el.Center).Gap(12).Px(16)
	w := core.CurrentWindow()
	controls := w != nil && w.Frameless()
	mac := v.chrome.goos == "darwin"
	drag := el.Div().Grow().W(el.Dp(0)).MinW(el.Dp(24)).MinH(el.Dp(AppShellTitleBarHeight)).Justify(el.Center).Gap(2)
	if v.header != nil {
		bar.Child(v.header.Render(cx))
	} else {
		drag.Child(el.Text(title).Bold().MaxLines(1))
		if description != "" {
			drag.Child(el.Text(description).TextSize(12).TextColor(theme.Muted).MaxLines(1))
		}
	}
	if controls {
		drag.OnDoubleClick(w.ToggleMaximize)
		dragArea(cx, w, drag)
	}
	bar.Child(drag)
	if p != nil && p.spec.Actions != nil {
		bar.Child(p.spec.Actions.Render(cx))
	}
	if v.actions != nil {
		bar.Child(v.actions.Render(cx))
	}
	if controls && !mac {
		bar.Pr(0).Child(v.chrome.buttons(cx, w))
	}
	return bar
}

func (v *AppShellView) sidebarTitlebar(cx *el.Context) el.Element {
	width := v.nav.width
	if v.Collapsed() {
		width = 56
	}
	top := el.Div().MinH(el.Dp(AppShellTitleBarHeight)).NoShrink().Row().Items(el.Center).Px(12).Gap(8)
	w := core.CurrentWindow()
	if w != nil && w.Frameless() && v.chrome.goos == "darwin" {
		if native, ok := w.(core.NativeTitleBarControls); ok && native.NativeTitleBarInset() > 0 {
			top.Child(el.Div().W(el.Dp(max(0, native.NativeTitleBarInset()-12))).NoShrink())
		} else {
			top.Child(v.chrome.lights(cx, w))
		}
	}
	top.Child(v.sidebarToggle(cx))
	if v.sidebarHeader != nil {
		top.Child(v.sidebarHeader.Render(cx))
	} else if !v.Collapsed() {
		top.Child(el.Text(v.title).Bold().MaxLines(1))
	}

	if w != nil && w.Frameless() && v.chrome.goos == "darwin" && v.Collapsed() {
		width = 124
		if native, ok := w.(core.NativeTitleBarControls); ok {
			width = max(width, native.NativeTitleBarInset()+48)
		}
		if v.sidebarHeader != nil {
			width += 36
		}
	}
	if v.Collapsed() {
		// Custom controls remain visible when the navigation collapses. Measure
		// their full width instead of squeezing them into the 56dp icon rail.
		return top.MinW(el.Dp(width))
	}
	return top.W(el.Dp(width))
}
