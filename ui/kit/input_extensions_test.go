package kit

import (
	"errors"
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/io/transfer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
	"io"
	"strings"
	"testing"
)

func TestInputMaskValues(t *testing.T) {
	prefix := Input("").Mask("+1 (###)")
	prefix.SetValue("123")
	if prefix.Value() != "+1 (123)" || prefix.UnmaskedValue() != "123" {
		t.Fatal("literal prefix consumed raw digits", prefix.Value())
	}
	prefix.SetValue(prefix.Value())
	if prefix.UnmaskedValue() != "123" {
		t.Fatal("mask not idempotent")
	}
	v := Input("").Mask("(###)-###-####")
	v.SetValue("1234567890")
	if v.Value() != "(123)-456-7890" || v.UnmaskedValue() != "1234567890" || !v.MaskComplete() {
		t.Fatal(v.Value(), v.UnmaskedValue())
	}
	v.SetValue("123")
	if v.Value() != "(123" || v.MaskComplete() {
		t.Fatal("partial", v.Value())
	}
	v.Mask(`AA-##9`)
	v.SetValue("中文123")
	if v.Value() != "中文-123" || !v.MaskComplete() {
		t.Fatal("Unicode", v.Value())
	}
	v.SetValue("中文12")
	if !v.MaskComplete() {
		t.Fatal("optional slot required")
	}
	v.Mask(`\`)
	v.SetValue("中文12")
	if v.Value() != "中文-12" {
		t.Fatal("invalid mask replaced config")
	}
	v.NumberMask(',', 2)
	v.SetValue("-12345678901234567890.1234")
	if v.Value() != "-12,345,678,901,234,567,890.12" || v.UnmaskedValue() != "-12345678901234567890.12" {
		t.Fatal("precision lost", v.Value())
	}
	v.SetValue("-.")
	if v.MaskComplete() {
		t.Fatal("draft complete")
	}
	v.NumberMask(0, 0)
	v.SetValue("123.456")
	if v.Value() != "123" {
		t.Fatal("integer truncation")
	}
	v.Mask("##-##").Filter("123").MaxLength(3)
	v.SetValue("123456")
	if v.Value() != "12-3" || v.MaskComplete() {
		t.Fatal("filter/max", v.Value())
	}
	v.Mask("")
	if v.UnmaskedValue() != "12-3" {
		t.Fatal("removing mask changed value")
	}
}

func TestInputMaskEditingUndoAndMenu(t *testing.T) {
	v := Input("masked").Mask("##-##")
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "masked")
	h.Type("1234")
	h.Frame()
	if v.Value() != "12-34" {
		t.Fatal(v.Value())
	}
	cx.SelectInput(v.FocusID(), 3, 3)
	h.Frame()
	h.Key(key.NameDeleteBackward, 0)
	h.Frame()
	if v.Value() != "13-4" {
		t.Fatal("separator backspace", v.Value())
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if v.Value() != "12-34" {
		t.Fatal("undo", v.Value())
	}
	h.Key("Z", key.ModShortcut|key.ModShift)
	h.Frame()
	if v.Value() != "13-4" {
		t.Fatal("redo", v.Value())
	}
	cx.SelectInput(v.FocusID(), 0, 2)
	h.Frame()
	inputRightClick(h, "masked")
	click(t, h, locale.Current().Cut)
	h.Frame()
	if v.Value() != "4" {
		t.Fatal("menu cut bypassed normalization", v.Value())
	}
	if !cx.Focused(v.FocusID()) {
		t.Fatal("menu focus not restored")
	}
	v.ContextMenuEnabled(false)
	h.Frame()
	inputRightClick(h, "masked")
	if v.editMenu.Value() {
		t.Fatal("disabled context menu opened")
	}
}

func clipboardText(h *uitest.Harness, s string) {
	h.Router.Queue(transfer.DataEvent{Type: "application/text", Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader(s)) }})
	h.Frame()
	h.Frame()
}

func TestInputPasteHookFallbackAndAsync(t *testing.T) {
	for _, multi := range []bool{false, true} {
		v := Input("paste")
		if multi {
			v = TextArea("paste")
		}
		calls, failures := 0, 0
		v.OnPaste(func(d core.ClipboardData) bool { calls++; return len(d.Images) > 0 }).OnPasteError(func(error) { failures++ })
		var cx *el.Context
		h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
		clickClass(t, h, "Editor", "paste")
		h.Key("V", key.ModShortcut)
		clipboardText(h, "hello")
		if v.Value() != "hello" || calls != 1 {
			t.Fatal("text hook", multi, v.Value(), calls)
		}
		var done func(core.ClipboardData, error)
		v.PasteReader(func(fn func(core.ClipboardData, error)) { done = fn })
		h.Frame()
		h.Key("V", key.ModShortcut)
		if done == nil {
			t.Fatal("reader not called")
		}
		done(core.ClipboardData{Images: []core.ClipboardImage{{MIME: "image/png", Data: []byte{1}}}}, nil)
		h.Frame()
		if calls != 2 || v.Value() != "hello" {
			t.Fatal("consumed rich paste")
		}
		h.Key("V", key.ModShortcut)
		done(core.ClipboardData{}, errors.New("reader failed"))
		h.Frame()
		clipboardText(h, "!")
		if failures != 1 || v.Value() != "hello!" || calls != 3 {
			t.Fatal("fallback", v.Value(), calls, failures)
		}
		h.Key("V", key.ModShortcut)
		stale := done
		v.SetValue("new")
		h.Frame()
		stale(core.ClipboardData{Text: "old"}, nil)
		h.Frame()
		if v.Value() != "new" || calls != 3 {
			t.Fatal("stale text accepted")
		}
		cx.SelectInput(v.FocusID(), 3, 3)
		h.Frame()
		h.Key("V", key.ModShortcut)
		stale = done
		cx.SelectInput(v.FocusID(), 0, 0)
		h.Frame()
		stale(core.ClipboardData{Text: "old"}, nil)
		h.Frame()
		if v.Value() != "new" || calls != 3 {
			t.Fatal("stale selection accepted")
		}
		h.Key("V", key.ModShortcut)
		stale = done
		h.Key("V", key.ModShortcut)
		fresh := done
		stale(core.ClipboardData{Text: "old"}, nil)
		fresh(core.ClipboardData{Text: "latest"}, nil)
		fresh(core.ClipboardData{Text: "twice"}, nil)
		h.Frame()
		if v.Value() != "latestnew" || calls != 4 {
			t.Fatal("latest-only", v.Value(), calls)
		}
		h.Key("V", key.ModShortcut)
		stale = done
		v.SetReadOnly(true)
		h.Frame()
		stale(core.ClipboardData{Text: "readonly"}, nil)
		h.Frame()
		if v.Value() != "latestnew" || calls != 4 {
			t.Fatal("readonly pending paste")
		}
	}
}

func inputRightClick(h *uitest.Harness, name string) {
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Class.String() == "Editor" && n.Desc.Label == name {
			p := f32.Pt(float32(n.Desc.Bounds.Min.X+2), float32(n.Desc.Bounds.Min.Y+2))
			h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary, Position: p}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, node := range h.Router.AppendSemantics(nil) {
		walk(node)
	}
	h.Frame()
	h.Frame()
}

func TestInputPasteRejectsChangedThenRestoredAndQueuedTyping(t *testing.T) {
	v := Input("paste")
	var done func(core.ClipboardData, error)
	v.PasteReader(func(fn func(core.ClipboardData, error)) { done = fn })
	h := render(func(c *el.Context) el.Element { return v.Render(c) })
	clickClass(t, h, "Editor", "paste")
	h.Key("V", key.ModShortcut)
	h.Type("x")
	h.Key("A", key.ModShortcut)
	h.Key(key.NameDeleteBackward, 0)
	h.Frame()
	if v.Value() != "" {
		t.Fatal("test must restore original input", v.Value())
	}
	done(core.ClipboardData{Text: "stale"}, nil)
	h.Frame()
	if v.Value() != "" {
		t.Fatal("paste survived intervening edit", v.Value())
	}
	h.Key("V", key.ModShortcut)
	done(core.ClipboardData{Text: "stale"}, nil)
	h.Router.Queue(key.EditEvent{Text: "new"})
	h.Frame()
	if v.Value() != "new" {
		t.Fatal("paste raced queued typing", v.Value())
	}
}

func TestInputMenuReadOnlyPasswordAndCustom(t *testing.T) {
	for _, password := range []bool{false, true} {
		v := Input("restricted")
		v.SetValue("secret")
		if password {
			v.Password()
		} else {
			v.SetReadOnly(true)
		}
		var cx *el.Context
		h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
		cx.SelectInput(v.FocusID(), 0, 6)
		h.Frame()
		inputRightClick(h, "restricted")
		if !v.editMenu.Value() {
			t.Fatal("menu missing")
		}
		if !v.editMenu.items[1].disabled || password && !v.editMenu.items[0].disabled || !password && !v.editMenu.items[2].disabled {
			t.Fatal("restricted commands enabled")
		}
		click(t, h, locale.Current().Cut)
		h.Frame()
		if v.Value() != "secret" {
			t.Fatal("restricted cut")
		}
		v.ContextMenuEnabled(false)
		h.Frame()
		count := 0
		v.ContextMenu(Menu().Item("custom action", "", func() { count++ })).ContextMenuEnabled(true)
		h.Frame()
		inputRightClick(h, "restricted")
		click(t, h, "custom action")
		h.Frame()
		if count != 1 {
			t.Fatal("custom menu")
		}
		v.SetDisabled(true)
		h.Frame()
		inputRightClick(h, "restricted")
		if v.customMenu.Value() {
			t.Fatal("disabled custom menu")
		}
	}
}

type pasteReadCloser struct {
	io.Reader
	closed *bool
}

func (r pasteReadCloser) Close() error { *r.closed = true; return nil }
func TestInputPasteOversizeClosesStream(t *testing.T) {
	failed, calls := 0, 0
	v := Input("paste").OnPaste(func(core.ClipboardData) bool { calls++; return false }).OnPasteError(func(error) { failed++ })
	h := render(func(cx *el.Context) el.Element { return v.Render(cx) })
	clickClass(t, h, "Editor", "paste")
	h.Key("V", key.ModShortcut)
	closed := false
	h.Router.Queue(transfer.DataEvent{Type: "application/text", Open: func() io.ReadCloser {
		return pasteReadCloser{strings.NewReader(strings.Repeat("x", (16<<20)+1)), &closed}
	}})
	h.Frame()
	if !closed || failed != 1 || calls != 0 || v.Value() != "" {
		t.Fatal("oversized clipboard handling", closed, failed, calls)
	}
}
