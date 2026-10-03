package el

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"strings"
	"testing"
)

func TestInputAutoGrowUnboundAndReflow(t *testing.T) {
	width := float32(320)
	var field *InputEl
	root := Root(viewFunc(func(cx *Context) Element {
		field = TextArea().ID("draft").AutoGrow(1, 8)
		return Div().Child(Div().W(Dp(width)).Child(field))
	}))
	h := uitest.New(root)
	empty := field.n.size.Y
	h.Click(30, 25)
	h.Type("some words to wrap across several short lines")
	h.Frame()
	wide := field.n.size.Y
	width = 110
	h.Frame()
	narrow := field.n.size.Y
	if narrow <= wide || narrow <= empty {
		t.Fatalf("no wrapped reflow: empty=%d wide=%d narrow=%d", empty, wide, narrow)
	}
	state := root.store.states[field.n.key]
	h.Type(strings.Repeat("\nline", 30))
	h.Frame()
	capped := field.n.size.Y
	if !strings.Contains(state.editor.Text(), strings.Repeat("\nline", 30)) {
		t.Fatalf("unbound text lost: %q", state.editor.Text())
	}
	h.Type(" end")
	h.Frame()
	if field.n.size.Y != capped || !strings.Contains(state.editor.Text(), " end") {
		t.Fatal("overflow stopped editing or exceeded cap")
	}
}
