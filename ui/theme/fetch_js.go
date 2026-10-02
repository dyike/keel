//go:build js

package theme

import (
	"fmt"
	"syscall/js"
)

// FetchFonts downloads font files with the browser's fetch, relative to the
// page, and adds them with LoadFonts. A WebAssembly build calls it before
// window.Main: the browser gives the app no system fonts, so without a CJK
// font Chinese shows as boxes. It blocks until every file arrived.
func FetchFonts(urls ...string) error {
	var files [][]byte
	for _, u := range urls {
		data, err := fetch(u)
		if err != nil {
			return fmt.Errorf("fetch %s: %w", u, err)
		}
		files = append(files, data)
	}
	return LoadFonts(files...)
}

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
			done <- result{err: fmt.Errorf("HTTP %s", resp.Get("status").String())}
			return nil
		}
		return resp.Call("arrayBuffer").Call("then", got, fail)
	})
	defer check.Release()
	js.Global().Call("fetch", url).Call("then", check).Call("catch", fail)
	r := <-done
	return r.data, r.err
}
