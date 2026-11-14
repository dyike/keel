//go:build js

package main

import (
	"strconv"
	"strings"
	"syscall/js"

	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
)

var gallerySafeArea layout.Inset

func init() {
	read := func() layout.Inset {
		doc := js.Global().Get("document")
		probe := doc.Call("createElement", "div")
		probe.Get("style").Set("cssText", "position:fixed;visibility:hidden;pointer-events:none;padding:env(safe-area-inset-top,0px) env(safe-area-inset-right,0px) env(safe-area-inset-bottom,0px) env(safe-area-inset-left,0px)")
		doc.Get("body").Call("appendChild", probe)
		style := js.Global().Call("getComputedStyle", probe)
		dp := func(name string) unit.Dp {
			n, err := strconv.ParseFloat(strings.TrimSuffix(style.Get(name).String(), "px"), 32)
			if err != nil || n < 0 {
				return 0
			}
			return unit.Dp(n)
		}
		result := layout.Inset{Top: dp("paddingTop"), Right: dp("paddingRight"), Bottom: dp("paddingBottom"), Left: dp("paddingLeft")}
		probe.Call("remove")
		return result
	}
	gallerySafeArea = read()
	listener := js.FuncOf(func(js.Value, []js.Value) any {
		insets := read()
		core.UpdateIf(func() bool {
			if gallerySafeArea == insets {
				return false
			}
			gallerySafeArea = insets
			return true
		})
		return nil
	})
	js.Global().Call("addEventListener", "resize", listener)
}

func galleryLayout(w core.Widget, gtx core.C) core.D {
	return gallerySafeArea.Layout(gtx, w.Layout)
}
