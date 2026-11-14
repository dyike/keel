//go:build js

package main

import (
	"flag"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
	"log"
)

// The browser entry excludes native screenshot/matrix export and filesystem
// theme watching. Query flags remain compatible with documentation embeds.
func main() {
	section := flag.String("section", "", "component or category")
	language := flag.String("lang", "en", "en or zh-CN")
	palette := flag.String("theme", "light", "theme name")
	flag.Parse()
	if *language == "zh-CN" {
		locale.Apply(locale.Chinese())
	} else {
		locale.Apply(locale.English())
	}
	if err := theme.Use(*palette); err != nil {
		log.Fatal(err)
	}
	g := newGallery()
	var content core.Widget
	if *section == "" {
		setupGalleryNavigation(g)
		content = el.Root(g)
	} else {
		var ok bool
		content, ok = sectionContent(*section)
		if !ok {
			log.Fatalf("unknown section %q", *section)
		}
	}
	mainWindow = window.Open(window.Options{Title: demoText("Keel · Components", "Keel · 组件"), Content: &galleryFirstFrame{Widget: content}})
	window.Main()
}
