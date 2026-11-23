package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall/js"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/window"
)

func main() {
	s := newStudio()
	loc := js.Global().Get("location")
	read := func() string {
		switch loc.Get("hash").String() {
		case "#studio":
			return "Studio"
		case "#templates":
			return "Templates"
		default:
			return "Home"
		}
	}
	s.page = read()
	s.onNavigate = func(page string) {
		if page == read() {
			return
		}
		url := loc.Get("pathname").String() + loc.Get("search").String()
		if page != "Home" {
			url += "#" + strings.ToLower(page)
		}
		js.Global().Get("history").Call("pushState", nil, "", url)
	}
	restore := js.FuncOf(func(js.Value, []js.Value) any {
		page := read()
		core.UpdateIf(func() bool {
			if s.page == page {
				return false
			}
			s.page = page
			s.playing = false
			return true
		})
		return nil
	})
	js.Global().Call("addEventListener", "hashchange", restore)
	opts := options(s)
	opts.CustomBrowserHistory = true
	window.Open(opts)
	window.Main()
}

func saveMovie(c composition) (string, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	resp, err := http.Post("/render", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
			return "", fmt.Errorf("start the local Keel Motion server with -web to enable MP4 exports")
		}
		return "", fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(message)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	array := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(array, data)
	blob := js.Global().Get("Blob").New([]any{array}, map[string]any{"type": "video/mp4"})
	urlAPI := js.Global().Get("URL")
	url := urlAPI.Call("createObjectURL", blob)
	doc := js.Global().Get("document")
	link := doc.Call("createElement", "a")
	link.Set("href", url)
	link.Set("download", templates[c.Scene].name+".mp4")
	doc.Get("body").Call("appendChild", link)
	link.Call("click")
	link.Call("remove")
	time.AfterFunc(time.Minute, func() { urlAPI.Call("revokeObjectURL", url) })
	return "Downloaded " + templates[c.Scene].name + ".mp4", nil
}
