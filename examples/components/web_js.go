//go:build js

package main

import (
	"log"
	"net/url"
	"os"
	"syscall/js"

	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// In a browser the gallery takes its flags from the page address, so the
// documentation site can embed one component: demo/?section=dock&theme=dark.
// It also fetches the CJK font published beside the page, since a browser
// gives WebAssembly no system fonts.
func init() {
	search := js.Global().Get("location").Get("search").String() // "?section=dock"
	if len(search) > 0 {
		search = search[1:]
	}
	q, err := url.ParseQuery(search)
	if err == nil {
		if q.Get("lang") == "zh-CN" {
			locale.Apply(locale.Chinese())
		} else {
			locale.Apply(locale.English())
		}
		for _, name := range []string{"section", "theme"} {
			if v := q.Get(name); v != "" {
				os.Args = append(os.Args, "-"+name+"="+v)
			}
		}
	}
	if err := theme.FetchFonts("font.otf"); err != nil {
		log.Printf("Chinese will not render: %v", err)
	}
}
