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

// A browser gives a WebAssembly app no system fonts: fetch a CJK font served
// beside index.html (copy one there as font.ttf, e.g. Noto Sans SC).
func init() {
	search := js.Global().Get("location").Get("search").String()
	if len(search) > 0 {
		search = search[1:]
	}
	if q, err := url.ParseQuery(search); err == nil {
		language := q.Get("lang")
		if language == "zh-CN" {
			locale.Apply(locale.Chinese())
		} else {
			locale.Apply(locale.English())
		}
		if language != "" {
			os.Args = append(os.Args, "-lang="+language)
		}
	}
	if err := theme.FetchFonts("font.ttf"); err != nil {
		log.Printf(text("Chinese text will not render: %v", "中文文字无法显示：%v"), err)
	}
}
