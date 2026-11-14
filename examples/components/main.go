//go:build !js

// Components is the kit gallery: a sidebar of every component, the selected
// one beside it. -section shows a single component on its own:
//
//	go run ./examples/components
//	go run ./examples/components -section menu -theme dark
package main

import (
	"flag"
	"log"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func main() {
	language := flag.String("lang", "en", "example language: en or zh-CN")
	section := flag.String("section", "", "show one component, or a category (general, inputs, display, data, charts, chat, feedback, navigation, overlays, shell, motion, foundations), instead of the app")
	screenshot := flag.String("screenshot", "", "render a PNG and exit")
	width := flag.Int("width", 680, "viewport width in dp")
	height := flag.Int("height", 1040, "screenshot height in dp")
	scale := flag.Float64("scale", 2, "screenshot pixels per dp")
	matrix := flag.String("matrix", "", "write all component light/dark, narrow/standard, 1x/2x screenshots and an HTML index")
	palette := flag.String("theme", "light", "a registered theme: light, dark, nord, aurora, ...")
	themeDir := flag.String("themes", "", "also load *.json themes from this directory and reload them as they change")
	flag.Parse()
	switch *language {
	case "en":
		locale.Apply(locale.English())
	case "zh-CN":
		locale.Apply(locale.Chinese())
	default:
		log.Fatalf("unknown language %q", *language)
	}
	if *themeDir != "" {
		w, err := theme.WatchThemes(*themeDir, time.Second, func(names []string, err error) {
			if err != nil {
				log.Print(err)
			}
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Stop()
	}
	if err := theme.Use(*palette); err != nil {
		log.Fatal(err)
	}
	if *matrix != "" {
		if err := renderMatrix(*matrix); err != nil {
			log.Fatal(err)
		}
		return
	}
	g := newGallery()
	if *section == "" {
		setupGalleryNavigation(g)
	}
	var content core.Widget = el.Root(g)
	if *section != "" {
		var ok bool
		if content, ok = sectionContent(*section); !ok {
			log.Fatalf("unknown section %q", *section)
		}
	}
	if *screenshot != "" {
		theme.SetReducedMotion(true) // Capture settled content, including entry animations.
		if err := window.ScreenshotAtScale(content, *width, *height, float32(*scale), *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *section == "" && *width == 680 {
		*width = 1200 // the app needs room for its sidebar
	}
	content = &galleryFirstFrame{Widget: content}
	mainWindow = window.Open(window.Options{Title: demoText("Keel · Components", "Keel · 组件"), Width: *width, Height: 860, Content: content})
	window.Main()
}
