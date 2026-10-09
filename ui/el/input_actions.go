package el

import (
	"io"
	"strings"

	"github.com/dyike/keel/third_party/gio/io/clipboard"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/inputcontent"
)

// InputAction is an editing command directed to one identified input.
type InputAction uint8

const (
	InputCopy InputAction = iota
	InputCut
	InputPaste
	InputSelectAll
	InputUndo
	InputRedo
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
	if action > InputRedo || !cx.root.e.gtx.Enabled() {
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
	case InputUndo, InputRedo:
		e.inputUndoAction(n, st, action == InputRedo)
	}
	return nil, true
}

// Menu operations preserve the editor's focus, including rich document editors.
func (e *engine) inputMenuActions(n *Node, st *elemState) {
	for {
		action, ok := core.NextEditAction(e.gtx, &st.editor)
		if !ok {
			return
		}
		switch action {
		case core.EditCopy:
			st.inputActions = append(st.inputActions, InputCopy)
		case core.EditCut:
			st.inputActions = append(st.inputActions, InputCut)
		case core.EditPaste:
			st.inputActions = append(st.inputActions, InputPaste)
		case core.EditSelectAll:
			st.inputActions = append(st.inputActions, InputSelectAll)
		case core.EditUndo:
			st.inputActions = append(st.inputActions, InputUndo)
		case core.EditRedo:
			st.inputActions = append(st.inputActions, InputRedo)
		}
	}
}
