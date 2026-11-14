package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// NavPage describes a registered navigation destination. Create is lazy and is
// called once per stack entry, so repeated pushes can have independent state.
// Fill lets content own its scrolling; otherwise NavStack supplies a scroll area.
type NavPage struct {
	ID, Title string
	Create    func() el.View
	Fill      bool
	Disabled  bool
}

type navEntry struct {
	page NavPage
	root *el.RootWidget
}

// NavStackView retains earlier pages while showing only the top entry. Popping
// discards the removed entry; pushing it again creates a fresh page instance.
// It does not own browser history, OS safe areas or application persistence.
type NavStackView struct {
	pages    []NavPage
	entries  []*navEntry
	disabled bool
	onChange func([]string)
}

func NavStack() *NavStackView { return &NavStackView{} }

// Page registers a destination. Empty/duplicate IDs are ignored. The first
// enabled page becomes the initial root without invoking OnChange.
func (v *NavStackView) Page(page NavPage) *NavStackView {
	if page.ID == "" || v.page(page.ID) != nil {
		return v
	}
	v.pages = append(v.pages, page)
	if len(v.entries) == 0 && !page.Disabled {
		v.entries = append(v.entries, &navEntry{page: page})
	}
	return v
}
func (v *NavStackView) page(id string) *NavPage {
	for i := range v.pages {
		if v.pages[i].ID == id {
			return &v.pages[i]
		}
	}
	return nil
}

// Value returns a copy of the route IDs from root to current page.
func (v *NavStackView) Value() []string {
	ids := make([]string, len(v.entries))
	for i, e := range v.entries {
		ids[i] = e.page.ID
	}
	return ids
}

// SetValue restores a route path silently, retaining the existing common prefix.
// An empty path clears the stack. Unknown/disabled IDs reject the entire path.
// Assignment remains available while disabled, unlike Push/Pop user commands.
func (v *NavStackView) SetValue(ids []string) {
	pages := make([]NavPage, len(ids))
	for i, id := range ids {
		p := v.page(id)
		if p == nil || p.Disabled {
			return
		}
		pages[i] = *p
	}
	common := 0
	for common < len(ids) && common < len(v.entries) && ids[common] == v.entries[common].page.ID {
		common++
	}
	entries := make([]*navEntry, len(ids))
	copy(entries, v.entries[:common])
	for i := common; i < len(ids); i++ {
		entries[i] = &navEntry{page: pages[i]}
	}
	v.entries = entries
}
func (v *NavStackView) OnChange(fn func([]string)) *NavStackView { v.onChange = fn; return v }
func (v *NavStackView) SetDisabled(on bool)                      { v.disabled = on }

// Push is a navigation command: successful changes notify OnChange. Disabled or
// unknown pages are ignored. Use SetValue for silent programmatic restoration.
func (v *NavStackView) Push(id string) bool {
	p := v.page(id)
	if v.disabled || p == nil || p.Disabled {
		return false
	}
	v.entries = append(v.entries, &navEntry{page: *p})
	v.changed()
	return true
}

// Pop navigates back one page and notifies OnChange. The root is never popped.
func (v *NavStackView) Pop() bool {
	if v.disabled || len(v.entries) < 2 {
		return false
	}
	v.entries = slices.Delete(v.entries, len(v.entries)-1, len(v.entries))
	v.changed()
	return true
}
func (v *NavStackView) changed() {
	if v.onChange != nil {
		v.onChange(v.Value())
	}
}

func (v *NavStackView) Render(cx *el.Context) el.Element {
	id := autoID("nav-stack", v)
	outer := el.Div().ID(id).Disabled(v.disabled).Items(el.Stretch)
	if len(v.entries) == 0 {
		return outer
	}
	entry := v.entries[len(v.entries)-1]
	if entry.root == nil {
		var view el.View
		if entry.page.Create != nil {
			view = entry.page.Create()
		}
		entry.root = el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			body := el.Div().ID("nav-stack/content").WFull().HFull().Items(el.Stretch)
			if !entry.page.Fill {
				body.ScrollY().P(theme.Space2xl).Gap(theme.SpaceXl)
			}
			if view != nil {
				body.Child(view.Render(cx))
			}
			return body
		}))
	}
	header := el.Div().NoShrink().Row().Items(el.Center).Gap(theme.SpaceMd).P(theme.SpaceLg)
	if len(v.entries) > 1 {
		header.Child(Button("", func() { v.Pop() }).ID(id + "/back").Name(locale.Current().PrevPage).
			Icon(IconChevronLeft).Variant(ButtonGhost).Size(float32(theme.ControlHeight) + theme.SpaceMd).Render(cx))
	}
	header.Child(el.Div().Grow().W(el.Dp(0)).Child(el.Text(entry.page.Title).TextSize(theme.TextLg).Bold()))
	return outer.Role("group").Name(entry.page.Title).Child(header,
		el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border),
		el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(el.Widget(entry.root).Grow()),
	)
}
