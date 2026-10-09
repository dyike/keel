package el

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
)

func (e *engine) inputUndoKeys(n *Node, st *elemState) {
	ed := &st.editor
	for {
		ev, ok := e.gtx.Event(key.Filter{Focus: ed, Name: "Z", Required: key.ModShortcut, Optional: key.ModShift}, key.Filter{Focus: ed, Name: "Y", Required: key.ModShortcut})
		if !ok {
			return
		}
		ke, ok := ev.(key.Event)
		if !ok || ke.State != key.Press || !e.gtx.Enabled() || ed.ReadOnly {
			continue
		}
		e.inputUndoAction(n, st, ke.Name == "Y" || ke.Modifiers.Contain(key.ModShift))
	}
}

func (e *engine) inputUndoAction(n *Node, st *elemState, redo bool) {
	ed := &st.editor
	if !e.gtx.Enabled() || ed.ReadOnly {
		return
	}
	from, to := &st.inputUndo, &st.inputRedo
	if redo {
		from, to = to, from
	}
	if len(*from) == 0 {
		return
	}
	start, end := ed.Selection()
	*to = append(*to, InputEdit{Text: ed.Text(), Start: start, End: end})
	snapshot := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]
	ed.SetText(snapshot.Text)
	ed.SetCaret(snapshot.Start, snapshot.End)
	st.lastText = snapshot.Text
	if n.input.bind != nil {
		*n.input.bind = snapshot.Text
	}
	if fn := n.input.onChange; fn != nil {
		core.Call(e.gtx, func() { fn(snapshot.Text) })
	}
}
