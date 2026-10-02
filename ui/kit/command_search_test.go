package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestCommandGroupsDisabledAndOwnedItems(t *testing.T) {
	ran := ""
	source := []CommandItem{{Title: "disabled", Group: "Group", Disabled: true, Action: func() { ran = "bad" }}, {Title: "enabled", Group: "Group", Action: func() { ran = "ok" }}}
	cmd := Command(source...)
	source[1].Title = "external"
	h := page(cmd)
	cmd.SetValue(true)
	h.Frame()
	h.Frame()
	if !shown(h, "Group") || !shown(h, "enabled") || shown(h, "external") {
		t.Fatal("groups or item ownership")
	}
	click(t, h, "disabled")
	if ran != "" {
		t.Fatal("disabled item ran")
	}
	h.Key(key.NameHome, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran != "ok" || cmd.Value() {
		t.Fatal("keyboard failed to skip group/disabled item")
	}
}

func TestCommandAsyncRejectsStaleResultsAndRetries(t *testing.T) {
	var tokens []uint64
	cmd := Command().OnSearch(func(query string, token uint64) { tokens = append(tokens, token) })
	h := page(cmd)
	cmd.SetValue(true)
	h.Frame()
	h.Frame()
	if len(tokens) != 1 || !shown(h, "加载中") {
		t.Fatal("initial request")
	}
	h.Type("new")
	h.Frame()
	if len(tokens) != 2 {
		t.Fatalf("query requests %v", tokens)
	}
	if cmd.SetResults(tokens[0], CommandItem{Title: "stale"}) {
		t.Fatal("stale result accepted")
	}
	if !cmd.SetSearchError(tokens[1], "network failed") {
		t.Fatal("current error rejected")
	}
	h.Frame()
	click(t, h, "重试")
	h.Frame()
	if len(tokens) != 3 || !cmd.loading {
		t.Fatal("retry missing")
	}
	// Remote results need not fuzzy-match the input (semantic search).
	result := []CommandItem{{Title: "remote result"}}
	if !cmd.SetResults(tokens[2], result...) {
		t.Fatal("current results rejected")
	}
	result[0].Title = "external"
	h.Frame()
	h.Frame()
	if !shown(h, "remote result") || shown(h, "external") {
		t.Fatal("remote results filtered or aliased")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if cmd.SetResults(tokens[2], CommandItem{Title: "late"}) {
		t.Fatal("closed palette accepted results")
	}
	cmd.SetDisabled(true)
	cmd.SetValue(true)
	if cmd.Value() {
		t.Fatal("disabled palette opened")
	}
}

func TestCommandVirtualResultsAndFocusReturn(t *testing.T) {
	items := make([]CommandItem, 10000)
	for i := range items {
		items[i] = CommandItem{Title: fmt.Sprintf("command-%05d", i)}
	}
	cmd := Command(items...)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element {
		cx = c
		return el.Div().Child(Button("open", cmd.Toggle).Render(c), cmd.Render(c))
	})
	click(t, h, "open")
	h.Frame()
	h.Frame()
	first, last := cmd.list.visible(cx)
	if last-first > 40 {
		t.Fatalf("virtual range %d %d", first, last)
	}
	cmd.active = 9990
	h.Key(key.NamePageDown, 0)
	h.Frame()
	h.Frame()
	if cmd.active != 9999 || !shown(h, "command-09999") {
		t.Fatalf("PageDown reveal %d", cmd.active)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Key(key.NameSpace, 0)
	h.Frame()
	if !cmd.Value() {
		t.Fatal("focus did not return to trigger")
	}
	h.Frame()
	if cmd.active != 0 || !shown(h, "command-00000") {
		t.Fatal("reopening retained old scroll position")
	}
}
