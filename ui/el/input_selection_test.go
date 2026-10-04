package el

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestInputSelectOnFocusAndProgramSelection(t *testing.T) {
	value := "中文09"
	var cx *Context
	in := Input().ID("edit").Bind(&value).SelectOnFocus(true)
	root := Root(viewFunc(func(c *Context) Element { cx = c; return in }))
	h := uitest.New(root)
	cx.Focus("edit")
	h.Frame()
	h.Frame()
	st := root.store.states[in.node().key]
	a, b := st.editor.Selection()
	if a != 0 || b != 4 {
		t.Fatal("focus did not select rune range", a, b)
	}
	cx.SelectInput("edit", 2, 99)
	h.Frame()
	a, b = st.editor.Selection()
	if a != 2 || b != 4 {
		t.Fatal("program selection not clamped", a, b)
	}
}

func TestInputCapturedKeysPreserveModifiedEditing(t *testing.T) {
	value := "12"
	calls := 0
	in := Input().ID("edit").Bind(&value).CaptureKeys(string(key.NameLeftArrow)).OnKey(func(e KeyEvent) bool {
		if e.Name == string(key.NameLeftArrow) {
			calls++
		}
		return true
	})
	root := Root(viewFunc(func(*Context) Element { return in }))
	h := uitest.New(root)
	cx := &Context{root: root}
	cx.Focus("edit")
	h.Frame()
	h.Frame()
	st := root.store.states[in.node().key]
	st.editor.SetCaret(2, 2)
	h.Key(key.NameLeftArrow, 0)
	a, b := st.editor.Selection()
	if a != 2 || b != 2 || calls != 2 {
		t.Fatal("captured arrow moved caret", a, b, calls)
	}
	h.Key(key.NameLeftArrow, key.ModShift)
	a, b = st.editor.Selection()
	if a == b || calls != 2 {
		t.Fatal("modified editing shortcut captured", a, b, calls)
	}
}
