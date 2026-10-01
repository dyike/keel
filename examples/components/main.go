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
	content, ok := sectionContent(*section)
	if !ok {
		log.Fatalf("unknown section %q", *section)
	}
	if *screenshot != "" {
		if err := window.Screenshot(content, 680, 1040, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "Keel · 组件", Width: 680, Height: 860, Content: content})
	window.Main()
}
