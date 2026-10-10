package el

import (
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

// Queries made while building a view must not remain stale in callbacks or
// the rerender triggered by the same input frame.
func TestInteractionQueriesRefreshAfterDispatch(t *testing.T) {
	var hovered, within, focus, callbackHover bool
	disabled := false
	root := Root(ViewFunc(func(cx *Context) Element {
		hovered, within, focus = cx.Hovered("button"), cx.FocusWithin("wrap"), cx.FocusVisible("button")
		return Div().Child(Div().ID("wrap").Disabled(disabled).Child(
			Div().ID("button").Name("button").Size(Dp(40)).Focusable(true).OnClick(func() {
				callbackHover = cx.Hovered("button")
			})))
	}))
	h := uitest.New(root)
	if hovered || within || focus {
		t.Fatal("empty prior tree has interaction state")
	}
	h.Click(20, 20)
	h.Frame()
	if !callbackHover || !hovered || !within || focus {
		t.Fatalf("pointer state: callback=%v hover=%v within=%v focus-ring=%v", callbackHover, hovered, within, focus)
	}
	h.Key(key.NameSpace, 0)
	h.Frame()
	if !within || !focus {
		t.Fatal("keyboard focus was not refreshed")
	}
	disabled = true
	h.Frame()
	h.Frame()
	if hovered || within || focus {
		t.Fatal("disabled state retained cached interaction queries")
	}
}
