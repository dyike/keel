package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// EmptyPart identifies a region for per-frame style refinement.
type EmptyPart uint8

const (
	EmptyPartRoot EmptyPart = iota
	EmptyPartHeader
	EmptyPartMedia
	EmptyPartTitle
	EmptyPartDescription
	EmptyPartContent
	EmptyPartFooter
)

// EmptyView explains why a region contains no results.
type EmptyView struct {
	title, description                  string
	icon                                IconName
	action                              el.View
	media                               el.View
	heading, descriptionContent, footer el.View
	styles                              [7]func(*el.DivEl)
}

func Empty(title string) *EmptyView                  { return &EmptyView{title: title, icon: IconInbox} }
func (v *EmptyView) Description(s string) *EmptyView { v.description = s; return v }
func (v *EmptyView) Icon(i IconName) *EmptyView      { v.icon = i; return v }
func (v *EmptyView) Action(e el.View) *EmptyView     { v.action = e; return v }

// Media replaces the icon with arbitrary content, preserving its size and semantics.
// Nil restores the configured icon; IconNone hides that fallback.
func (v *EmptyView) Media(view el.View) *EmptyView { v.media = view; return v }
func (v *EmptyView) SetTitle(s string)             { v.title = s }
func (v *EmptyView) SetDescription(s string)       { v.description = s }

// Heading replaces the title text; nil restores it.
func (v *EmptyView) Heading(view el.View) *EmptyView { v.heading = view; return v }

// DescriptionContent replaces the description text; nil restores it.
func (v *EmptyView) DescriptionContent(view el.View) *EmptyView {
	v.descriptionContent = view
	return v
}

// Footer adds supporting content after Action; nil removes it.
func (v *EmptyView) Footer(view el.View) *EmptyView { v.footer = view; return v }

// PartStyle refines a fresh region each frame, after defaults. Do not retain it.
// Nil restores defaults. Invalid parts are ignored; slot identities are preserved.
func (v *EmptyView) PartStyle(part EmptyPart, fn func(*el.DivEl)) *EmptyView {
	if int(part) < len(v.styles) {
		v.styles[part] = fn
	}
	return v
}
func (v *EmptyView) part(part EmptyPart, id string, box *el.DivEl) *el.DivEl {
	if fn := v.styles[part]; fn != nil {
		fn(box)
	}
	if part == EmptyPartRoot {
		return box
	}
	return box.ID(id)
}
func (v *EmptyView) Render(cx *el.Context) el.Element {
	box := v.part(EmptyPartRoot, "empty", el.Div().P(theme.Space2xl).Gap(theme.SpaceMd).Items(el.Center).Bg(theme.Surface))
	header := v.part(EmptyPartHeader, "header", el.Div().Gap(theme.SpaceMd).Items(el.Center).MaxW(el.Full))
	hasHeader := false
	if v.media != nil || v.icon != IconNone {
		media := v.part(EmptyPartMedia, "media", el.Div().Items(el.Center).MaxW(el.Full))
		if v.media != nil {
			media.Child(v.media.Render(cx))
		} else {
			media.Child(Icon(v.icon).Size(40).Color(theme.Muted).Render(cx))
		}
		header.Child(media)
		hasHeader = true
	}
	if v.heading != nil || v.title != "" {
		title := v.part(EmptyPartTitle, "title", el.Div().MaxW(el.Full).TextColor(theme.Text))
		if v.heading != nil {
			title.Child(v.heading.Render(cx))
		} else {
			title.Child(el.Text(v.title))
		}
		header.Child(title)
		hasHeader = true
	}
	if v.descriptionContent != nil || v.description != "" {
		description := v.part(EmptyPartDescription, "description", el.Div().MaxW(el.Full).TextColor(theme.Muted))
		if v.descriptionContent != nil {
			description.Child(v.descriptionContent.Render(cx))
		} else {
			description.Child(el.Text(v.description))
		}
		header.Child(description)
		hasHeader = true
	}
	if hasHeader {
		box.Child(header)
	}
	if v.action != nil {
		box.Child(v.part(EmptyPartContent, "action", el.Div().Items(el.Center).MaxW(el.Full)).Child(v.action.Render(cx)))
	}
	if v.footer != nil {
		box.Child(v.part(EmptyPartFooter, "footer", el.Div().Items(el.Center).MaxW(el.Full).TextColor(theme.Muted)).Child(v.footer.Render(cx)))
	}
	return box
}
