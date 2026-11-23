package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

var (
	ink   = theme.RGB(0x152139)
	muted = theme.RGB(0x55637b)
	rule  = theme.RGB(0xd9e0ea)
	blue  = theme.RGB(0x225bff)
	white = theme.RGB(0xffffff)
)

type studio struct {
	page       string
	onNavigate func(string)
	comp       composition
	frame      int
	playing    bool
	start      time.Time
	timeline   *kit.SliderView
	exporting  bool
	message    string
}

func newStudio() *studio {
	s := &studio{page: "Home"}
	s.timeline = kit.Slider("", 0, float64(s.comp.frames()-1)).Step(1).OnChange(func(f float64) { s.seek(int(f)) })
	return s
}

func (s *studio) navigate(page string) {
	s.page = page
	s.playing = false
	if s.onNavigate != nil {
		s.onNavigate(page)
	}
}

func (s *studio) seek(frame int) {
	s.playing = false
	s.frame = min(max(frame, 0), s.comp.frames()-1)
}

func (s *studio) openTemplate(i int) {
	s.navigate("Studio")
	s.comp = composition{i, templates[i].defaultPalette}
	s.timeline.SetRange(0, float64(s.comp.frames()-1))
	s.seek(0)
}

func (s *studio) play(now time.Time) {
	if !s.playing {
		if s.frame == s.comp.frames()-1 {
			s.frame = 0
		}
		s.start = now.Add(-time.Duration(s.frame) * time.Second / fps)
	}
	s.playing = !s.playing
}

func (s *studio) export() {
	if s.exporting {
		return
	}
	s.playing = false
	s.exporting = true
	s.message = "Rendering " + templates[s.comp.Scene].name + "…"
	go func(c composition) {
		message, err := saveMovie(c)
		if err != nil {
			message = fmt.Sprintf("Export failed: %v", err)
		}
		core.Update(func() {
			s.exporting = false
			s.message = message
		})
	}(s.comp)
}

func label(str string, size float32, c color.NRGBA) *el.TextEl {
	return el.Text(str).TextSize(size).TextColor(c)
}
func mono(str string) *el.TextEl { return label(str, 11, muted).Mono() }
func action(id, str string, primary bool, fn func()) *el.DivEl {
	bg, fg, hover := white, ink, theme.RGB(0xedf1f9)
	if primary {
		bg, fg, hover = blue, white, theme.RGB(0x164cd9)
	}
	b := el.Div().ID(id).Name(str).Focusable(true).CursorPointer().OnClick(fn).Rounded(7).Px(16).H(el.Dp(40)).Center().Bg(bg).
		Hover(func(st *el.Style) { st.Bg(hover) }).Child(label(str, 14, fg).Medium())
	if !primary {
		b.Border(1, rule)
	}
	return b
}

func artwork(c composition, frame int) *el.DivEl {
	return el.Div().WFull().AspectRatio(16.0 / 9).Decorate(func(gtx core.C, children func()) {
		c.draw(gtx.Ops, gtx.Constraints.Max, frame)
		children()
	})
}

func (s *studio) Render(cx *el.Context) el.Element {
	width, _ := cx.ViewportSize()
	narrow := width < 820
	pad := float32(48)
	if narrow {
		pad = 20
	}
	if s.playing {
		now := cx.Now()
		s.frame = min(s.comp.frames()-1, max(0, int(now.Sub(s.start)*fps/time.Second)))
		s.playing = s.frame < s.comp.frames()-1
	}
	s.timeline.SetValue(float64(s.frame))
	brand := el.Div().ID("home").Focusable(true).Name("Keel Motion home").CursorPointer().OnClick(func() { s.navigate("Home") }).Row().Gap(9).
		Child(el.Div().Size(el.Dp(29)).Bg(blue).Rounded(7).Center().Child(label("K", 21, white).Bold().Italic()), label("Keel Motion", 20, ink).Bold())
	nav := el.Div().Row().Gap(16)
	if width >= 560 {
		nav.Child(el.Div().ID("templates").Name("Templates").CursorPointer().Focusable(true).OnClick(func() { s.navigate("Templates") }).P(8).Child(label("Templates", 14, ink)))
	}
	if s.page == "Studio" {
		nav.Child(action("back-home", "Back to home", false, func() { s.navigate("Home") }))
	} else {
		nav.Child(action("launch", "Open studio", true, func() { s.navigate("Studio") }))
	}
	header := el.Div().WFull().MaxW(el.Dp(1280)).Px(pad).H(el.Dp(68)).Row().Justify(el.SpaceBetween).Child(brand, nav)
	body := el.Div().WFull().MaxW(el.Dp(1280)).Px(pad)
	switch s.page {
	case "Home":
		body.Child(s.hero(cx, narrow), s.library(narrow, false))
	case "Templates":
		body.Child(s.library(narrow, true))
	case "Studio":
		body.Child(s.editor(cx, narrow))
	}
	root := el.Div().WFull().HFull().Bg(white).Items(el.Center)
	if s.playing {
		next := s.start.Add(time.Duration(s.frame+1)*time.Second/fps + time.Nanosecond)
		root.Decorate(func(gtx core.C, draw func()) {
			gtx.Execute(op.InvalidateCmd{At: next})
			draw()
		})
	}
	return root.Child(header, el.Div().WFull().H(el.Dp(1)).Bg(rule),
		el.Div().ID("scroll-"+s.page).WFull().Grow().ScrollY().Items(el.Center).Child(body))
}

func (s *studio) hero(cx *el.Context, narrow bool) el.Element {
	group := el.Div().Py(54).Gap(40)
	left := el.Div().Justify(el.Center).Items(el.Start)
	right := el.Div()
	titleSize := float32(66)
	if narrow {
		group.Py(30).Gap(24)
		titleSize = 44
	} else {
		group.Row().Items(el.Center)
		left.Flex(9).MinW(el.Dp(0))
		right.Flex(11).MinW(el.Dp(0))
	}
	left.Child(label("Make videos", titleSize, ink).Bold(), label("with Go.", titleSize, blue).Bold(),
		label("Build motion graphics in Go. Preview each frame in Keel and render your composition to video.", 17, muted).MaxW(el.Dp(404)).Mt(12),
		el.Div().Row().Gap(8).Mt(24).Child(action("create", "Create a video", true, func() { s.navigate("Studio") }),
			action("browse", "Browse templates", false, func() { s.navigate("Templates") })))
	right.Child(s.player(cx, false))
	return group.Child(left, right)
}

func (s *studio) player(cx *el.Context, editing bool) el.Element {
	c := s.comp
	top := el.Div().Row().Px(12).H(el.Dp(36)).Justify(el.SpaceBetween).Child(
		el.Div().Row().Gap(8).Child(el.Div().Size(el.Dp(7)).Rounded(4).Bg(palettes[c.Palette].base), mono(templates[c.Scene].name+" / preview")), mono("16:9"))
	playLabel := "Play"
	if s.playing {
		playLabel = "Pause"
	}
	control := el.Div().Row().Px(10).Gap(10).H(el.Dp(48)).Child(
		el.Div().ID("play").Name(playLabel).Focusable(true).CursorPointer().OnClick(func() { s.play(time.Now()) }).W(el.Dp(55)).H(el.Dp(30)).Center().Bg(theme.RGB(0xeaf0ff)).Rounded(5).Child(label(playLabel, 13, blue).Bold()),
		el.Div().Grow().Child(s.timeline.Render(cx)), mono(fmt.Sprintf("%03d / %d", s.frame+1, s.comp.frames())))
	box := el.Div().P(1).Bg(white).Border(1, rule).Rounded(9).Child(top, artwork(c, s.frame), control)
	if editing {
		return box
	}
	return el.Div().Bg(theme.RGB(0xf2f5fb)).Rounded(14).P(8).Child(box)
}

func (s *studio) library(narrow, full bool) el.Element {
	heading := "Starter templates"
	if full {
		heading = "Template library"
	}
	head := label(heading, 29, ink).Bold()
	grid := el.Div().Gap(18)
	if !narrow {
		grid.Row().Items(el.Start)
	}
	for i, t := range templates {
		card := el.Div().ID("template-"+t.name).Name("Use "+t.name).Focusable(true).CursorPointer().OnClick(func() { s.openTemplate(i) }).P(1).Border(1, rule).Rounded(8).
			Hover(func(st *el.Style) { st.BorderColor(blue) })
		if !narrow {
			card.Flex(1).MinW(el.Dp(0))
		}
		card.Child(artwork(composition{i, t.defaultPalette}, 0), el.Div().P(14).Gap(4).Child(
			el.Div().Row().Justify(el.SpaceBetween).Child(label(t.name, 19, ink).Bold(), mono(fmt.Sprintf("%ds / %d fps", t.seconds, fps))), label(t.description, 13, muted)))
		grid.Child(card)
	}
	section := el.Div().Gap(20).Pb(36)
	if full {
		section.Pt(32)
	} else {
		section.Child(el.Div().WFull().H(el.Dp(1)).Bg(rule))
	}
	return section.Child(head, grid)
}

func (s *studio) editor(cx *el.Context, narrow bool) el.Element {
	heading := el.Div().Gap(5).Child(label("Motion studio", 34, ink).Bold(), label("Preview, adjust and render your composition.", 15, muted))
	row := el.Div().Gap(26)
	preview := el.Div()
	panel := el.Div().Gap(14).P(18).Bg(theme.RGB(0xf6f8fc)).Rounded(9)
	if !narrow {
		row.Row().Items(el.Start)
		preview.Grow().MinW(el.Dp(0))
		panel.W(el.Dp(275))
	}
	preview.Child(s.player(cx, true), el.Div().Row().Mt(8).Gap(7).Child(
		action("prev", "Previous frame", false, func() { s.seek(s.frame - 1) }), action("next", "Next frame", false, func() { s.seek(s.frame + 1) })))
	panel.Child(label("Composition", 18, ink).Bold())
	for i, t := range templates {
		panel.Child(action("scene-"+t.name, t.name, s.comp.Scene == i, func() { s.openTemplate(i) }).Selected(s.comp.Scene == i))
	}
	swatches := el.Div().Row().Gap(8)
	for i, p := range palettes {
		selected := s.comp.Palette == i
		button := el.Div().ID("color-" + p.name).Name(p.name).Selected(selected).Size(el.Dp(34)).Rounded(17).Center().Focusable(true).CursorPointer().OnClick(func() { s.comp.Palette = i }).
			Child(el.Div().Size(el.Dp(24)).Rounded(12).Bg(p.base))
		if selected {
			button.Border(2, ink)
		}
		swatches.Child(button)
	}
	panel.Child(label("Palette · "+palettes[s.comp.Palette].name, 14, ink).Medium(), swatches,
		el.Div().H(el.Dp(1)).Bg(rule), mono(fmt.Sprintf("%d × %d   /   %d FPS   /   %d SEC", videoWidth, videoHeight, fps, templates[s.comp.Scene].seconds)),
		action("export", "Export MP4", true, s.export).Disabled(s.exporting))
	if s.message != "" {
		panel.Child(label(s.message, 13, muted))
	}
	return el.Div().Py(30).Gap(24).Child(heading, row.Child(preview, panel))
}

func options(s *studio) window.Options {
	return window.Options{Title: "Keel Motion", Width: 1190, Height: 850, Content: el.Root(s)}
}
