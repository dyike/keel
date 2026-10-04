package el

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/dyike/keel/ui/internal/inputcontent"
)

// InputAction is an editing command directed to one identified input.
type InputAction uint8

const (
	InputCopy InputAction = iota
	InputCut
	InputPaste
	InputSelectAll
)

// InputSelection returns the input's last editor text and rune selection.
// Missing or disabled inputs return false. It does not change focus.
func (cx *Context) InputSelection(id string) (InputEdit, bool) {
	for _, st := range cx.root.store.states {
		if st.id == id && st.edInit && !st.disabled {
			if d := st.inputDocument; d != nil {
				r, _ := inputcontent.RuneRange(d.Content().Text(), d.Selection())
				return InputEdit{Text: d.Content().Text(), Start: r.Start, End: r.End}, true
			}
			start, end := st.editor.Selection()
			return InputEdit{Text: st.editor.Text(), Start: start, End: end}, true
		}
	}
	return InputEdit{}, false
}

// InputAction queues a command for the next paint of an enabled input. Read-only
// inputs reject Cut/Paste; password inputs reject Copy/Cut. Paste uses the system
// clipboard's asynchronous text path and the input's normal filter/transform.
func (cx *Context) InputAction(id string, action InputAction) {
	if action > InputSelectAll || !cx.root.e.gtx.Enabled() {
		return
	}
	for _, st := range cx.root.store.states {
		if st.id == id && st.edInit && !st.disabled {
			st.inputActions = append(st.inputActions, action)
			cx.root.e.gtx.Execute(op.InvalidateCmd{})
			return
		}
	}
}

func (e *engine) inputAction(n *Node, st *elemState, action InputAction) (widget.EditorEvent, bool) {
	if !e.gtx.Enabled() {
		return nil, true
	}
	ed := &st.editor
	switch action {
	case InputCopy, InputCut:
		if n.input.password || action == InputCut && ed.ReadOnly {
			return nil, true
		}
		if text := ed.SelectedText(); text != "" {
			e.gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
			if action == InputCut {
				ed.Insert("")
				return widget.ChangeEvent{}, true
			}
		}
	case InputPaste:
		if !ed.ReadOnly {
			if n.input.onPaste != nil || n.input.pasteReader != nil {
				e.requestInputPaste(n, st)
			} else {
				e.gtx.Execute(clipboard.ReadCmd{Tag: ed})
			}
		}
	case InputSelectAll:
		ed.SetCaret(0, ed.Len())
	}
	return nil, true
}
