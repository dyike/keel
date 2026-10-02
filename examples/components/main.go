// Components is the kit gallery: a sidebar of every component, the selected
// one beside it. -section shows a single component on its own:
//
//	go run ./examples/components
//	go run ./examples/components -section menu -theme dark
package main

import (
	"flag"
	"log"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func main() {
	section := flag.String("section", "", "show one component, or a category (controls, inputs, overlays, data, shell), instead of the app")
	screenshot := flag.String("screenshot", "", "render a PNG and exit")
	width := flag.Int("width", 680, "viewport width in dp")
	height := flag.Int("height", 1040, "screenshot height in dp")
	scale := flag.Float64("scale", 2, "screenshot pixels per dp")
	matrix := flag.String("matrix", "", "write all component light/dark, narrow/standard, 1x/2x screenshots and an HTML index")
	palette := flag.String("theme", "light", "light or dark palette")
	flag.Parse()
	switch *palette {
	case "light":
		theme.Apply(theme.Light())
	case "dark":
		theme.Apply(theme.Dark())
	default:
		log.Fatalf("unknown theme %q", *palette)
	}
	if *matrix != "" {
		if err := renderMatrix(*matrix); err != nil {
			log.Fatal(err)
		}
		return
	}
	var content core.Widget = el.Root(newGallery())
	if *section != "" {
		var ok bool
		if content, ok = sectionContent(*section); !ok {
			log.Fatalf("unknown section %q", *section)
		}
	}
	if *screenshot != "" {
		if err := window.ScreenshotAtScale(content, *width, *height, float32(*scale), *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *section == "" && *width == 680 {
		*width = 1200 // the app needs room for its sidebar
	}
	window.Open(window.Options{Title: "Keel · 组件", Width: *width, Height: 860, Content: content})
	window.Main()
}
