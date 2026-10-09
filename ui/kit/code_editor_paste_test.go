package kit

import (
	"errors"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"testing"
)

func TestCodeEditorRichPasteAndStaleResults(t *testing.T) {
	v, h := focusedEditor("one\ntwo")
	var done func(core.ClipboardData, error)
	calls, failures := 0, 0
	v.PasteReader(func(fn func(core.ClipboardData, error)) { done = fn }).OnPaste(func(d core.ClipboardData) bool { calls++; return len(d.Files) > 0 }).OnPasteError(func(error) { failures++ })
	v.SetCursor(0, 0)
	v.sels = []codeSel{{codePos{0, 0}, codePos{0, 0}}, {codePos{1, 0}, codePos{1, 0}}}
	h.Frame()
	h.Key("V", key.ModShortcut)
	done(core.ClipboardData{Text: "A\nB"}, nil)
	h.Frame()
	if v.Value() != "Aone\nBtwo" || calls != 1 {
		t.Fatal("rich multi-caret paste", v.Value(), calls)
	}
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if v.Value() != "one\ntwo" {
		t.Fatal("paste did not undo atomically", v.Value())
	}
	h.Key("V", key.ModShortcut)
	done(core.ClipboardData{Files: []string{"/example.txt"}}, nil)
	h.Frame()
	if calls != 2 || v.Value() != "one\ntwo" {
		t.Fatal("consumed file paste")
	}
	h.Key("V", key.ModShortcut)
	stale := done
	typeAt(h, "x")
	h.Key("Z", key.ModShortcut)
	h.Frame()
	stale(core.ClipboardData{Text: "old"}, nil)
	h.Frame()
	if v.Value() != "one\ntwo" || calls != 2 {
		t.Fatal("revision guard", v.Value(), calls)
	}
	h.Key("V", key.ModShortcut)
	stale = done
	v.SetCursor(1, 1)
	h.Frame()
	v.SetCursor(0, 0)
	h.Frame()
	stale(core.ClipboardData{Text: "old"}, nil)
	h.Frame()
	if v.Value() != "one\ntwo" || calls != 2 {
		t.Fatal("selection guard")
	}
	h.Key("V", key.ModShortcut)
	done(core.ClipboardData{}, errors.New("failed"))
	h.Frame()
	clipboardText(h, "fallback")
	if v.Value() != "fallbackone\ntwo" || failures != 1 || calls != 3 {
		t.Fatal("native error fallback", v.Value(), failures, calls)
	}
	h.Key("V", key.ModShortcut)
	stale = done
	v.SetReadOnly(true)
	h.Frame()
	stale(core.ClipboardData{Text: "old"}, nil)
	h.Frame()
	if calls != 3 {
		t.Fatal("readonly handled paste")
	}
}

func TestCodeEditorCanceledTextPasteCannotArriveLate(t *testing.T) {
	v, h := focusedEditor("original")
	v.SetCursor(0, 0)
	h.Frame()
	h.Key("V", key.ModShortcut)
	typeAt(h, "new")
	clipboardText(h, "stale")
	if v.Value() != "neworiginal" {
		t.Fatal("canceled platform reply inserted", v.Value())
	}
}
