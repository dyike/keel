package el

import (
	"image"
	"reflect"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestFocusableKeyboardAndTabOrder(t *testing.T) {
	calls := 0
	first := Div().ID("first").Focusable().OnClick(func() { calls++ }).P(10).Child(Text("first"))
	hidden := Div().Focusable().Hidden(true).Child(Text("hidden"))
	second := Div().ID("second").Focusable().P(10).Child(Text("second"))
	request := true
	root := Root(viewFunc(func(cx *Context) Element {
		if request {
			cx.Focus("first")
			request = false
		}
		return Div().Child(first, hidden, second)
	}))
	h := uitest.New(root)
	isFocused := func(n Element) bool { return h.Router.Source().Focused(root.store.states[n.node().key]) }
	if !isFocused(first) {
		t.Fatal("program focus failed on first frame")
	}
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if calls != 2 {
		t.Fatalf("keyboard activation count %d", calls)
	}
	before := rect(second)
	// MoveFocus is the platform action for an unhandled Tab, also used by window.press.
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	if !isFocused(second) || rect(second) != before {
		t.Fatal("Tab order or layout changed")
	}
	h.Router.MoveFocus(key.FocusBackward)
	h.Frame()
	if !isFocused(first) {
		t.Fatal("Shift+Tab did not return to first")
	}
	r := nodeBounds(h, "second")
	h.Click(float32(r.Min.X+2), float32(r.Min.Y+2))
	h.Frame()
	if !isFocused(second) {
		t.Fatal("pointer did not focus")
	}
}

func TestKeyBubblingAndHandledActivation(t *testing.T) {
	var got []string
	clicks := 0
	child := Div().ID("child").Focusable().OnClick(func() { clicks++ }).OnKey(func(e KeyEvent) bool {
		if e.State == key.Press {
			got = append(got, "child")
		}
		return e.Name == key.NameSpace
	}).Child(Text("child"))
	parent := Div().OnKey(func(e KeyEvent) bool {
		if e.State == key.Press {
			got = append(got, "parent")
		}
		return true
	}).Child(child)
	request := true
	h := uitest.New(Root(viewFunc(func(cx *Context) Element {
		if request {
			cx.Focus("child")
			request = false
		}
		return parent
	})))
	h.Key("Q", key.ModCtrl)
	if !reflect.DeepEqual(got, []string{"child", "parent"}) {
		t.Fatalf("bubble order: %v", got)
	}
	got = nil
	h.Key(key.NameSpace, 0)
	if !reflect.DeepEqual(got, []string{"child"}) || clicks != 0 {
		t.Fatalf("handled key leaked: %v clicks=%d", got, clicks)
	}
}

func TestFocusableRemovalAndMissingTarget(t *testing.T) {
	show := true
	request := "target"
	calls := 0
	root := Root(viewFunc(func(cx *Context) Element {
		if request != "" {
			cx.Focus(request)
			request = ""
		}
		box := Div()
		if show {
			box.Child(Div().ID("target").Focusable().OnClick(func() { calls++ }).Child(Text("target")))
		}
		return box
	}))
	h := uitest.New(root)
	h.Key(key.NameSpace, 0)
	if calls != 1 {
		t.Fatal("first activation failed")
	}
	request = "missing"
	h.Frame()
	h.Key(key.NameSpace, 0)
	if calls != 2 {
		t.Fatal("missing ID cleared valid focus")
	}
	show = false
	h.Frame()
	h.Key(key.NameSpace, 0)
	if calls != 2 {
		t.Fatal("removed element activated")
	}
}

func TestFocusStyleDoesNotChangeDimensions(t *testing.T) {
	focused := false
	button := Div().ID("button").Focusable().P(8).Focus(func(s *Style) { s.BorderColor(rgb(0x00ff00)) }).Child(Text("Mixed 中文123"))
	root := Root(viewFunc(func(cx *Context) Element {
		if focused {
			cx.Focus("button")
		}
		return Div().Child(button)
	}))
	h := uitest.New(root)
	before := nodeBounds(h, "Mixed 中文123")
	focused = true
	h.Frame()
	h.Frame()
	if got := nodeBounds(h, "Mixed 中文123"); got != before || got == (image.Rectangle{}) {
		t.Fatalf("focus shifted text: before %v after %v", before, got)
	}
}

func TestProgramFocusInputAndNativeTabCoexist(t *testing.T) {
	text := ""
	request := true
	input := Input().ID("input").Bind(&text)
	button := Div().ID("button").Focusable().Child(Text("button"))
	root := Root(viewFunc(func(cx *Context) Element {
		if request {
			cx.Focus("input")
			request = false
		}
		return Div().Child(button, input)
	}))
	h := uitest.New(root)
	h.Type("中文")
	if text != "中文" {
		t.Fatal("program focus did not target editor")
	}
	h.Router.MoveFocus(key.FocusBackward)
	h.Frame()
	if !h.Router.Source().Focused(root.store.states[button.node().key]) {
		t.Fatal("input and button have separate Tab orders")
	}
}

func TestFocusChangeCancelsHeldActivation(t *testing.T) {
	calls := 0
	target := "a"
	root := Root(viewFunc(func(cx *Context) Element {
		if target != "" {
			cx.Focus(target)
			target = ""
		}
		return Div().Child(
			Div().ID("a").Focusable().OnClick(func() { calls++ }).Child(Text("a")),
			Div().ID("b").Focusable().OnClick(func() { calls++ }).Child(Text("b")),
		)
	}))
	h := uitest.New(root)
	h.Router.Queue(key.Event{Name: key.NameSpace, State: key.Press})
	h.Frame()
	target = "b"
	h.Frame()
	h.Router.Queue(key.Event{Name: key.NameSpace, State: key.Release})
	h.Frame()
	if calls != 0 {
		t.Fatal("release after focus change activated a target")
	}
	h.Key(key.NameSpace, 0)
	if calls != 1 {
		t.Fatal("fresh press/release failed")
	}
}
