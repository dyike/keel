//go:build js

package main

import (
	"log"
	"syscall/js"

	"github.com/dyike/keel/ui/theme"
)

// A browser gives a WebAssembly app no system fonts, so Chinese would show
// as boxes. Fetch a CJK font served beside index.html (copy one there as
// font.ttf, e.g. Noto Sans SC) and add it before the first frame.
func init() {
	data, err := fetch("font.ttf")
	if err != nil {
		log.Printf("no font.ttf beside index.html, Chinese will not render: %v", err)
		return
	}
	if err := theme.LoadFonts(data); err != nil {
		log.Print(err)
	}
}

// fetch downloads url with the browser's fetch and waits for the bytes.
func fetch(url string) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	fail := js.FuncOf(func(_ js.Value, args []js.Value) any {
		done <- result{err: js.Error{Value: args[0]}}
		return nil
	})
	defer fail.Release()
	got := js.FuncOf(func(_ js.Value, args []js.Value) any {
		buf := js.Global().Get("Uint8Array").New(args[0])
		data := make([]byte, buf.Get("length").Int())
		js.CopyBytesToGo(data, buf)
		done <- result{data: data}
		return nil
	})
	defer got.Release()
	check := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			done <- result{err: js.Error{Value: js.ValueOf("HTTP " + resp.Get("status").String())}}
			return nil
		}
		return resp.Call("arrayBuffer").Call("then", got, fail)
	})
	defer check.Release()
	js.Global().Call("fetch", url).Call("then", check).Call("catch", fail)
	r := <-done
	return r.data, r.err
}
