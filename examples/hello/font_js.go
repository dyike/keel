//go:build js

package main

import (
	"log"

	"github.com/dyike/keel/ui/theme"
)

// A browser gives a WebAssembly app no system fonts: fetch a CJK font served
// beside index.html (copy one there as font.ttf, e.g. Noto Sans SC).
func init() {
	if err := theme.FetchFonts("font.ttf"); err != nil {
		log.Printf("Chinese will not render: %v", err)
	}
}
