package el

import (
	"fmt"
	"io"
	"sync"

	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/transfer"
	"gioui.org/widget"
	"github.com/dyike/keel/ui/core"
)

type inputPasteRequest struct {
	before          InputEdit
	data            core.ClipboardData
	err             error
	ready, fallback bool
}

// OnPaste handles clipboard contents before default text insertion. Return true
// to consume them. It applies to keyboard and InputPaste commands; nil restores
// ordinary text paste. Read-only and disabled inputs do not invoke the handler.
func (v *InputEl) OnPaste(fn func(core.ClipboardData) bool) *InputEl {
	v.n.input.onPaste = fn
	return v
}

// PasteReader supplies rich clipboard contents. Without one, OnPaste receives
// Gio's text contents. Failed reads fall back to the Gio text path.
func (v *InputEl) PasteReader(reader core.ClipboardReader) *InputEl {
	v.n.input.pasteReader = reader
	return v
}

// OnPasteError reports read failures before fallback or rejection of oversized
// text. It runs on the UI thread; nil suppresses error reporting.
func (v *InputEl) OnPasteError(fn func(error)) *InputEl { v.n.input.onPasteError = fn; return v }

func (e *engine) requestInputPaste(n *Node, st *elemState) {
	if !e.gtx.Enabled() || n.input.readOnly {
		return
	}
	a, b := st.editor.Selection()
	r := &inputPasteRequest{before: InputEdit{Text: st.editor.Text(), Start: a, End: b}}
	st.inputPaste = r
	if reader := n.input.pasteReader; reader != nil {
		var once sync.Once
		reader(func(data core.ClipboardData, err error) {
			once.Do(func() {
				core.Update(func() {
					if st.inputPaste == r {
						r.data, r.err, r.ready = data, err, true
					}
				})
			})
		})
	} else {
		r.fallback = true
		e.gtx.Execute(clipboard.ReadCmd{Tag: r})
	}
}
func (e *engine) inputPasteKeys(n *Node, st *elemState) {
	if n.input.onPaste == nil && n.input.pasteReader == nil {
		return
	}
	for {
		ev, ok := e.gtx.Event(key.Filter{Focus: &st.editor, Name: "V", Required: key.ModShortcut})
		if !ok {
			return
		}
		if k, ok := ev.(key.Event); ok && k.State == key.Press {
			e.requestInputPaste(n, st)
		}
	}
}
func (e *engine) inputPasteEvent(n *Node, st *elemState) (widget.EditorEvent, bool) {
	r := st.inputPaste
	if r == nil {
		return nil, false
	}
	start, end := st.editor.Selection()
	valid := e.gtx.Enabled() && !n.input.readOnly && !st.disabled && st.editor.Text() == r.before.Text && start == r.before.Start && end == r.before.End
	if n.input.bind != nil && *n.input.bind != r.before.Text {
		valid = false
	}
	if selection := st.inputSelection; selection != nil && (selection[0] != r.before.Start || selection[1] != r.before.End) {
		valid = false
	}
	if !valid {
		st.inputPaste = nil
		return nil, true
	}
	if r.fallback && !r.ready {
		ev, ok := e.gtx.Event(transfer.TargetFilter{Target: r, Type: "application/text"})
		if !ok {
			return nil, false
		}
		data, ok := ev.(transfer.DataEvent)
		if !ok {
			return nil, true
		}
		reader := data.Open()
		bytes, err := io.ReadAll(io.LimitReader(reader, 16<<20+1))
		reader.Close()
		if len(bytes) > 16<<20 {
			err = fmt.Errorf("clipboard text exceeds 16MiB")
		}
		r.data, r.err, r.ready = core.ClipboardData{Text: string(bytes)}, err, true
	}
	if !r.ready {
		return nil, false
	}
	if r.err != nil {
		if fn := n.input.onPasteError; fn != nil {
			core.Call(e.gtx, func() { fn(r.err) })
		}
		if !r.fallback {
			r.err, r.ready, r.fallback = nil, false, true
			e.gtx.Execute(clipboard.ReadCmd{Tag: r})
			return nil, true
		}
		st.inputPaste = nil
		return nil, true
	}
	st.inputPaste = nil
	consumed := false
	if fn := n.input.onPaste; fn != nil {
		core.Call(e.gtx, func() { consumed = fn(r.data) })
	}
	if consumed || r.data.Text == "" || n.input.bind != nil && *n.input.bind != r.before.Text {
		return nil, true
	}
	st.editor.Insert(r.data.Text)
	return widget.ChangeEvent{}, true
}
