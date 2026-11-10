package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AppPage describes a destination. Create is called once, on first visit.
// Retain stateful controls in the returned view. IDs must be unique and nonempty.
type AppPage struct {
	ID, Title, Description string
	Icon                   IconName
	Create                 func() el.View
	Actions                el.View
	// Fill lets a page own its layout and scrolling instead of using padded scrolling content.
	Fill     bool
	Disabled bool
}

type shellPage struct {
	spec AppPage
	root *el.RootWidget
}

// AppShellView owns navigation, a page heading, actions and retained page roots.
// Keep the shell in application state rather than constructing it during Render.
type AppShellView struct {
	sidebarMin, sidebarMax, sidebarAvailable, resizeGrab, resizePainted float32
	resizing                                                            bool
	nav                                                                 *SidebarView
	chrome                                                              *TitleBarView
	title                                                               string
	sidebar, sidebarHeader, header, footer, contentFooter               el.View
	content                                                             *el.RootWidget
	pages                                                               []*shellPage
	selected, shown                                                     string
	actions                                                             el.View
	onChange                                                            func(string)
	fade                                                                valueMotion
	duration                                                            time.Duration
	breakpoint                                                          float32
	hideSidebarOnCollapse                                               bool
	compact, collapsed                                                  bool
}

// AppShell creates a full-window application layout with borderless navigation.
func AppShell(title string) *AppShellView {
	v := &AppShellView{title: title, chrome: TitleBar(title), nav: Sidebar().BorderWidth(0), duration: 150 * time.Millisecond, breakpoint: 640, sidebarMin: 200, sidebarMax: 420}
	v.nav.OnChange(func(id string) {
		v.SetValue(id)
		if v.onChange != nil {
			v.onChange(id)
		}
	})
	return v
}

// Page appends a destination. Invalid and duplicate IDs are ignored.
func (v *AppShellView) Page(page AppPage) *AppShellView {
	if page.ID == "" || v.page(page.ID) != nil {
		return v
	}
	v.pages = append(v.pages, &shellPage{spec: page})
	if page.Icon == IconNone {
		page.Icon = IconFile
	}
	item := SidebarItem{ID: page.ID, Label: page.Title, Icon: page.Icon, Disabled: page.Disabled}
	if len(v.nav.sections) == 0 {
		v.nav.Section("", item)
	} else {
		v.nav.sections[0].items = append(v.nav.sections[0].items, item)
	}
	if v.selected == "" && !page.Disabled {
		v.SetValue(page.ID)
	}
	return v
}
func (v *AppShellView) page(id string) *shellPage {
	for _, p := range v.pages {
		if p.spec.ID == id {
			return p
		}
	}
	return nil
}
func (v *AppShellView) Value() string { return v.selected }

// SetValue selects an enabled destination without invoking OnChange.
func (v *AppShellView) SetValue(id string) {
	if p := v.page(id); p != nil && !p.spec.Disabled {
		v.selected = id
		v.nav.SetValue(id)
	}
}
func (v *AppShellView) OnChange(fn func(string)) *AppShellView { v.onChange = fn; return v }

// Actions supplies persistent controls at the end of the page heading.
func (v *AppShellView) Actions(view el.View) *AppShellView { v.actions = view; return v }

// Footer supplies persistent content below navigation.
func (v *AppShellView) Footer(view el.View) *AppShellView     { v.footer = view; return v }
func (v *AppShellView) SidebarWidth(dp float32) *AppShellView { v.nav.Width(dp); return v }
func (v *AppShellView) SetCollapsed(on bool)                  { v.collapsed = on; v.nav.SetCollapsed(on || v.compact) }
func (v *AppShellView) Collapsed() bool                       { return v.nav.Collapsed() }

// CollapseBelow sets the automatic icon-rail breakpoint in dp, 640 by default.
// Zero disables responsive collapsing. The expanded-window preference is preserved.
func (v *AppShellView) CollapseBelow(dp float32) *AppShellView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.breakpoint = dp
	}
	return v
}

// Transition sets the page fade duration. Zero disables it. Reduced motion skips it.
func (v *AppShellView) Transition(d time.Duration) *AppShellView { v.duration = max(0, d); return v }

func (v *AppShellView) Render(cx *el.Context) el.Element {
	if !v.compact {
		v.collapsed = v.nav.Collapsed()
	}
	width, _ := cx.ViewportSize()
	v.compact = v.breakpoint > 0 && width > 0 && width < v.breakpoint
	v.nav.SetCollapsed(v.collapsed || v.compact)
	v.nav.Collapsible(false)
	v.sidebarAvailable = max(0, width-320)
	v.nav.width = v.clampSidebar(v.nav.width)
	body := el.Div().Grow().W(el.Dp(0)).MinH(el.Dp(0)).Items(el.Stretch)
	p := v.page(v.selected)
	if v.content != nil {
		body.Child(el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(el.Widget(v.content).Grow()))
	} else if p != nil {
		if p.root == nil {
			var view el.View
			if p.spec.Create != nil {
				view = p.spec.Create()
			}
			p.root = el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
				content := el.Div().ID("app-shell/content").WFull().HFull().Items(el.Stretch)
				if !p.spec.Fill {
					content.ScrollY().P(24).Gap(16)
				}
				if view != nil {
					content.Child(view.Render(cx))
				}
				return content
			}))
		}
		if v.shown != v.selected {
			// Remove the old page immediately and fade the
			// new page in; blending two text-heavy pages makes both hard to read.
			if v.shown != "" {
				v.fade = valueMotion{initialized: true, started: cx.Now()}
			}
			v.shown = v.selected
		}
		opacity := float32(1)
		if v.duration > 0 {
			opacity = v.fade.sample(cx, 1, v.duration)
		}
		body.Child(el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Opacity(opacity).Child(el.Widget(p.root).Grow()))
	}

	if v.contentFooter != nil {
		body.Child(el.Div().NoShrink().Items(el.Stretch).Child(v.contentFooter.Render(cx)))
	}
	headerPage := p
	if v.content != nil {
		headerPage = nil
	}
	row := el.Div().Row().Grow().MinH(el.Dp(0)).Items(el.Stretch)
	hidden := v.Collapsed() && v.hideSidebarOnCollapse
	if !hidden {
		row.Child(v.navigation(cx))
	}
	row.Child(body)
	dividerLeft := v.nav.width
	if v.Collapsed() {
		dividerLeft = 0
	}
	header := el.Div().Row().NoShrink().Items(el.Stretch).Child(
		v.sidebarTitlebar(cx), v.titlebar(cx, headerPage),
		el.Div().ID(autoID("app-shell-header-divider", v)).Absolute().Left(dividerLeft).Right(0).Bottom(0).H(el.Dp(1)).Bg(theme.Border),
	)
	root := el.Div().WFull().HFull().Items(el.Stretch).Bg(theme.Bg).Child(header, row)
	if !hidden {
		root.Child(v.sidebarDivider(cx))
	} else {
		v.resizing = false
	}
	return root
}
