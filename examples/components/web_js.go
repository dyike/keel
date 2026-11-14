//go:build js

package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"syscall/js"

	"github.com/dyike/keel/ui/core"
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
		for _, name := range []string{"section", "theme", "lang"} {
			if v := q.Get(name); v != "" {
				os.Args = append(os.Args, "-"+name+"="+v)
			}
		}
	}
	if err := theme.FetchFonts("font.otf"); err != nil {
		log.Printf("Chinese will not render: %v", err)
	}
}

// One listener per gallery lifetime, with state mutations posted to the frame
// lock. Component previews (-section) don't change their host's history.
func setupGalleryNavigation(g *gallery) {
	location := js.Global().Get("location")
	history := js.Global().Get("history")
	read := func() string {
		q, _ := url.ParseQuery(strings.TrimPrefix(location.Get("hash").String(), "#"))
		return q.Get("component")
	}
	routeURL := func(name string) string {
		base := location.Get("pathname").String() + location.Get("search").String()
		if name != "" {
			base += "#component=" + url.QueryEscape(name)
		}
		return base
	}
	set := func(name string, push bool) {
		method := "replaceState"
		if push {
			method = "pushState"
		}
		history.Call(method, map[string]any{"keelGallery": true, "index": name == ""}, "", routeURL(name))
	}
	initial := read()
	set("", false)
	if _, ok := g.sections[initial]; ok {
		g.nav.SetValue(initial)
		g.mobileDetail = true
		set(initial, true)
	}
	g.onNavigate = func(name string) {
		if read() != name {
			set(name, true)
		}
	}
	g.onBack = func() {
		state := history.Get("state")
		if !state.IsNull() && !state.IsUndefined() && state.Get("keelGallery").Truthy() && !state.Get("index").Truthy() {
			history.Call("back")
			return
		}
		g.mobileDetail = false
		set("", false)
	}
	callback := js.FuncOf(func(js.Value, []js.Value) any {
		name := read()
		core.Update(func() {
			if _, ok := g.sections[name]; ok {
				g.nav.SetValue(name)
				g.mobileDetail = true
			} else {
				g.mobileDetail = false
			}
		})
		return nil
	})
	js.Global().Call("addEventListener", "popstate", callback)
}

func galleryReady() {
	// Let the browser present the operations submitted by the first layout.
	var callback js.Func
	callback = js.FuncOf(func(js.Value, []js.Value) any {
		js.Global().Call("dispatchEvent", js.Global().Get("Event").New("keel-ready"))
		callback.Release()
		return nil
	})
	js.Global().Call("requestAnimationFrame", callback)
}

func switchGalleryBundle(name, category string) {
	location := js.Global().Get("location")
	target := galleryBundleURL(location.Get("pathname").String(), location.Get("search").String(), name, category, locale.Current().Lang, theme.CurrentName())
	location.Call("assign", target)
}
