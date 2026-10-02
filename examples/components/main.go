// Components shows every kit component, one section each:
//
//	go run ./examples/components -section menu -theme dark
package main

import (
	"flag"
	"log"

	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func main() {
	section := flag.String("section", "all", "component name, a category (controls, inputs, overlays, data), or all")
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
	content, ok := sectionContent(*section)
	if !ok {
		log.Fatalf("unknown section %q", *section)
	}
	if *screenshot != "" {
		if err := window.ScreenshotAtScale(content, *width, *height, float32(*scale), *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "Keel · 组件", Width: *width, Height: 860, Content: content})
	window.Main()
}
