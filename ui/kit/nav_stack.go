package kit

import (
	"math"
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
	// Modal starts a modal presentation; subsequent ordinary pushes stay inside it.
	Modal bool
}

type navEntry struct {
	page                            NavPage
	root                            *el.RootWidget
	swipeStart, swipeY, swipeOffset float32
	swiping                         bool
}

// NavStackView retains earlier pages and shows each presentation's top entry. Popping
// discards the removed entry; pushing it again creates a fresh page instance.
// It does not own browser history, OS safe areas or application persistence.
type NavStackView struct {
	pages    []NavPage
	entries  []*navEntry
	disabled bool
	onChange func([]string)
	swipeOff bool
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
	v.cancelSwipe()
	v.entries = entries
}
func (v *NavStackView) OnChange(fn func([]string)) *NavStackView { v.onChange = fn; return v }
func (v *NavStackView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.cancelSwipe()
	}
}

// SwipeBack enables left-edge drag navigation (enabled by default).
func (v *NavStackView) SwipeBack(on bool) *NavStackView {
	v.swipeOff = !on
	if !on {
		v.cancelSwipe()
	}
	return v
}
func (v *NavStackView) cancelSwipe() {
	for _, e := range v.entries {
		e.swiping = false
		e.swipeOffset = 0
	}
}

// Dismiss removes the latest modal presentation, including pages pushed inside
// it. It notifies once and returns false when there is no modal to dismiss.
func (v *NavStackView) Dismiss() bool {
	if v.disabled {
		return false
	}
	for i := len(v.entries) - 1; i > 0; i-- {
		if v.entries[i].page.Modal {
			v.cancelSwipe()
			v.entries = slices.Delete(v.entries, i, len(v.entries))
			v.changed()
			return true
		}
	}
	return false
}

// Push is a navigation command: successful changes notify OnChange. Disabled or
// unknown pages are ignored. Use SetValue for silent programmatic restoration.
func (v *NavStackView) Push(id string) bool {
	p := v.page(id)
	if v.disabled || p == nil || p.Disabled {
		return false
	}
	v.cancelSwipe()
	v.entries = append(v.entries, &navEntry{page: *p})
	v.changed()
	return true
}

// Pop navigates back one page and notifies OnChange. The root is never popped.
func (v *NavStackView) Pop() bool {
	if v.disabled || len(v.entries) < 2 {
		return false
	}
	v.cancelSwipe()
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
	// Each modal starts a presentation. Only its last page is rendered;
	// earlier pages retain their roots without producing frames or input.
	start := 0
	for i := 1; i <= len(v.entries); i++ {
		if i < len(v.entries) && !v.entries[i].page.Modal {
			continue
		}
		page := v.renderEntry(cx, v.entries[i-1], i-1, start > 0)
		if start == 0 {
			outer.Child(page.Grow())
		} else {
			modalID := autoID("nav-stack/modal", v.entries[start])
			panel := page.FocusTrap(true).W(el.Dp(640)).MaxW(el.Full).HFull().Bg(theme.Surface).Role("dialog").Name(v.entries[i-1].page.Title)
			cx.Overlay(modalID, el.Modal(panel).Owner(id).OnDismiss(func() { v.Dismiss() }))
		}
		start = i
	}
	return outer
}

func (v *NavStackView) renderEntry(cx *el.Context, entry *navEntry, index int, modal bool) *el.DivEl {
	id := autoID("nav-stack/page", entry)
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
				content := el.Div().ID("nav-stack/page-body").WFull().Items(el.Stretch).Child(view.Render(cx))
				if entry.page.Fill {
					content.HFull()
				}
				if entry.swiping {
					content.AnimatePose(el.Pose{X: entry.swipeOffset}, el.Instant)
				} else {
					content.AnimatePose(el.Pose{}, el.Snappy)
				}
				body.Child(content)
			}
			if !v.swipeOff && len(v.entries) > 1 && v.entries[len(v.entries)-1] == entry {
				body.DragAccept(func(dx, dy float32) bool {
					return entry.swiping && entry.swipeStart <= 24 && dx > 0 && dx > float32(math.Abs(float64(dy)))*1.5
				}).OnDrag(func(e el.DragEvent) {
					if v.disabled || len(v.entries) < 2 || v.entries[len(v.entries)-1] != entry {
						return
					}
					switch e.Kind {
					case el.DragStart:
						entry.swiping = true
						entry.swipeStart, entry.swipeY = e.X, e.Y
						entry.swipeOffset = 0
					case el.DragMove:
						if entry.swiping {
							entry.swipeOffset = max(0, min(e.W, e.X-entry.swipeStart))
						}
					case el.DragEnd:
						commit := entry.swiping && !e.Canceled && entry.swipeStart <= 24 && e.X-entry.swipeStart >= max(float32(64), e.W*.3) && e.X-entry.swipeStart > float32(math.Abs(float64(e.Y-entry.swipeY)))*1.5
						entry.swiping = false
						entry.swipeOffset = 0
						if commit {
							v.Pop()
						}
					}
				})
			}
			return body
		}))
	}
	header := el.Div().NoShrink().Row().Items(el.Center).Gap(theme.SpaceMd).P(theme.SpaceLg)
	if index > 0 {
		header.Child(Button("", func() { v.Pop() }).ID(id + "/back").Name(locale.Current().PrevPage).
			Icon(IconChevronLeft).Variant(ButtonGhost).Size(float32(theme.ControlHeight) + theme.SpaceMd).Render(cx))
	}
	header.Child(el.Div().Grow().W(el.Dp(0)).Child(el.Text(entry.page.Title).TextSize(theme.TextLg).Bold()))
	if modal {
		header.Child(Button("", func() { v.Dismiss() }).Name(locale.Current().Close).Icon(IconClose).Variant(ButtonGhost).Size(44).Render(cx))
	}
	return el.Div().Items(el.Stretch).Reveal(1).Disabled(v.disabled).Role("group").Name(entry.page.Title).Child(header,
		el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border),
		el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(el.Widget(entry.root).Grow()),
	)
}
