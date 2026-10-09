package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestInputTokensEditAndUndo(t *testing.T) {
	for _, multi := range []bool{false, true} {
		t.Run(map[bool]string{false: "Input", true: "TextArea"}[multi], func(t *testing.T) {
			v := Input("tokens")
			if multi {
				v = TextArea("tokens")
			}
			c, err := NewInputContent("a/ref/z", InputTokenSpan{Range: InputRange{Start: 1, End: 6}, Token: InputToken{ID: "resource", Text: "/ref/", Label: "引用"}})
			if err != nil {
				t.Fatal(err)
			}
			if err = v.SetContent(c); err != nil {
				t.Fatal(err)
			}
			changes := 0
			v.OnChange(func(string) { changes++ })
			var cx *el.Context
			h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
			clickClass(t, h, "Editor", "tokens")
			cx.SelectInput(v.FocusID(), 6, 6)
			h.Frame()
			h.Key(key.NameDeleteBackward, 0)
			h.Frame()
			if v.Value() != "az" || len(v.Content().Tokens()) != 0 {
				t.Fatal("partial token delete", v.Value(), v.Content().Tokens())
			}
			h.Key("Z", key.ModShortcut)
			h.Frame()
			if v.Value() != "a/ref/z" || len(v.Content().Tokens()) != 1 {
				t.Fatal("undo", v.Value())
			}
			cx.SelectInput(v.FocusID(), 1, 6)
			h.Frame()
			// An explicit same-visible-text edit must remove the reference metadata.
			h.Router.Queue(key.EditEvent{Range: key.Range{Start: 1, End: 3}, Text: "引用"})
			h.Frame()
			if v.Value() != "a引用z" || len(v.Content().Tokens()) != 0 {
				t.Fatal("same visible text edit", v.Value())
			}
			h.Key("Z", key.ModShortcut)
			h.Frame()
			if v.Value() != "a/ref/z" || len(v.Content().Tokens()) != 1 {
				t.Fatal("reference undo", v.Value())
			}
			if changes != 4 {
				t.Fatal("change count", changes)
			}
		})
	}
}

func TestInputTokenCaretAndProgramReset(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("a/path/z", InputTokenSpan{Range: InputRange{Start: 1, End: 7}, Token: InputToken{ID: "ref", Text: "/path/", Label: "LongLabel"}})
	v.SetContent(c)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	cx.SelectInput(v.FocusID(), 1, 1)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	selection, ok := cx.InputSelection(v.FocusID())
	if !ok || selection.Start != 7 || selection.End != 7 {
		t.Fatal("right arrow entered label", selection)
	}
	h.Key(key.NameLeftArrow, 0)
	h.Frame()
	selection, _ = cx.InputSelection(v.FocusID())
	if selection.Start != 1 || selection.End != 1 {
		t.Fatal("left arrow entered label", selection)
	}
	v.SetValue(v.Value())
	h.Frame()
	if len(v.Content().Tokens()) != 0 {
		t.Fatal("same-value reset retained references")
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if len(v.Content().Tokens()) != 0 || v.Value() != "a/path/z" {
		t.Fatal("reset retained undo")
	}
}

func TestInputTokenCompositionAndReadOnly(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("ref", InputTokenSpan{Range: InputRange{Start: 0, End: 3}, Token: InputToken{ID: "ref", Text: "ref", Label: "引用"}})
	v.SetContent(c)
	h := render(func(c *el.Context) el.Element { return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	h.Router.Queue(key.CompositionEvent{Start: 0, End: 2}, key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "ni"})
	h.Frame()
	if err := v.ReplaceWithToken(InputToken{ID: "new", Text: "new"}); err == nil {
		t.Fatal("token inserted during composition")
	}
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "你"}, key.CompositionEvent{Start: -1, End: -1})
	h.Frame()
	if v.Value() != "你" {
		t.Fatal(v.Value())
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if v.Value() != "ref" || len(v.Content().Tokens()) != 1 {
		t.Fatal("IME undo", v.Value())
	}
	v.SetReadOnly(true)
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "bad"})
	h.Frame()
	h.Key(key.NameDeleteBackward, 0)
	h.Frame()
	if v.Value() != "ref" {
		t.Fatal("read-only changed")
	}
	if err := v.ReplaceWithToken(InputToken{ID: "new", Text: "new"}); err == nil {
		t.Fatal("read-only token inserted")
	}
	v.Password()
	h.Frame()
	if v.document != nil {
		t.Fatal("password retained token display mode")
	}
}

func TestInputTokenRichPasteAndStaleResult(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("ref", InputTokenSpan{Range: InputRange{Start: 0, End: 3}, Token: InputToken{ID: "ref", Text: "ref", Label: "引用"}})
	v.SetContent(c)
	var done func(core.ClipboardData, error)
	calls := 0
	v.PasteReader(func(fn func(core.ClipboardData, error)) { done = fn }).OnPaste(func(core.ClipboardData) bool { calls++; return false })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	cx.SelectInput(v.FocusID(), 0, 3)
	h.Frame()
	h.Key("V", key.ModShortcut)
	h.Frame()
	if done == nil {
		t.Fatal("rich reader not invoked")
	}
	done(core.ClipboardData{Text: "引用"}, nil)
	h.Frame()
	if v.Value() != "引用" || len(v.Content().Tokens()) != 0 || calls != 1 {
		t.Fatal("paste", v.Value(), calls)
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if v.Value() != "ref" || len(v.Content().Tokens()) != 1 {
		t.Fatal("paste undo")
	}
	h.Key("V", key.ModShortcut)
	h.Frame()
	late := done
	v.SetContent(c)
	h.Frame() // Same characters and references, new draft generation.
	late(core.ClipboardData{Text: "stale"}, nil)
	h.Frame()
	if v.Value() != "ref" || calls != 1 {
		t.Fatal("stale result", v.Value(), calls)
	}
}

func TestInputTokenWordDeletion(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("hello path suffix", InputTokenSpan{Range: InputRange{Start: 6, End: 10}, Token: InputToken{ID: "ref", Text: "path", Label: "long reference name"}})
	v.SetContent(c)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	cx.SelectInput(v.FocusID(), 10, 10)
	h.Frame()
	h.Key(key.NameDeleteBackward, key.ModShortcutAlt)
	h.Frame()
	if v.Value() != "hello  suffix" || len(v.Content().Tokens()) != 0 {
		t.Fatal("word deletion split reference", v.Value())
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	cx.SelectInput(v.FocusID(), 6, 6)
	h.Frame()
	h.Key(key.NameRightArrow, key.ModShortcutAlt)
	h.Frame()
	sel, _ := cx.InputSelection(v.FocusID())
	if sel.Start != 10 || sel.End != 10 {
		t.Fatal("word movement entered label", sel)
	}
}

func TestInputTokenPasteProgramChangeIsSilent(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("draft")
	v.SetContent(c)
	changes := 0
	v.OnChange(func(string) { changes++ }).OnPaste(func(core.ClipboardData) bool { v.SetValue("program"); return true })
	h := render(func(c *el.Context) el.Element { return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	h.Key("V", key.ModShortcut)
	clipboardText(h, "ignored")
	if v.Value() != "program" || changes != 0 {
		t.Fatal("programmatic change notified", v.Value(), changes)
	}
}

func TestInputTokenQueuedCutRespectsDisable(t *testing.T) {
	v := Input("tokens")
	c, _ := NewInputContent("reference", InputTokenSpan{Range: InputRange{Start: 0, End: 9}, Token: InputToken{ID: "ref", Text: "reference", Label: "引用"}})
	v.SetContent(c)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "tokens")
	cx.SelectInput(v.FocusID(), 0, 9)
	h.Frame()
	cx.InputAction(v.FocusID(), el.InputCut)
	v.SetDisabled(true)
	h.Frame()
	if v.Value() != "reference" || len(v.Content().Tokens()) != 1 {
		t.Fatal("queued cut bypassed disabled state", v.Value())
	}
}
