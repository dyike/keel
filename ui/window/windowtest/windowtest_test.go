package windowtest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
	"github.com/dyike/keel/ui/window/windowtest"
)

type greeter struct {
	name   *kit.InputView
	result string
	loaded string
}

func (g *greeter) Render(cx *el.Context) el.Element {
	return el.Div().P(16).Gap(8).Child(
		g.name.Render(cx),
		kit.Button("Say hello", func() { g.result = "Hello, " + g.name.Value() + "!" }).Render(cx),
		el.Text(g.result),
		el.Text(g.loaded),
	)
}

func TestClickTypePressAndUpdate(t *testing.T) {
	g := &greeter{name: kit.Input("Name")}
	g.name.OnSubmit(func(v string) { g.result = "Submitted " + v })
	closed := false
	app := windowtest.New(t, window.Options{Title: "Greeter", Content: el.Root(g), OnClose: func() { closed = true }})

	app.Click("Name")
	app.Type("Ada")
	app.Click("Say hello")
	app.Want("Hello, Ada!")
	if e := app.Find("Name"); e.Role != "textbox" || e.Value != "Ada" {
		t.Fatalf("text box %+v", e)
	}

	app.Click("Name")
	app.Press("enter")
	app.Want("Submitted Ada")

	done := make(chan struct{})
	go func() {
		core.Update(func() { g.loaded = "Loaded 3 items" })
		close(done)
	}()
	<-done
	app.Want("Loaded 3 items") // every query renders, running queued updates first

	app.Window().Close()
	app.Render()
	if !closed || !app.Window().Closed() {
		t.Fatal("Close did not run OnClose")
	}
}

func TestDefaultSizeAndResize(t *testing.T) {
	app := windowtest.New(t, window.Options{Content: el.Root(&greeter{name: kit.Input("Name")})})
	if e := app.Find("Say hello"); e.X+e.Width > 800 || e.Width == 0 {
		t.Fatalf("button %+v outside the default 800 dp", e)
	}
	app.Resize(20, 20)
	app.WantNot("Say hello") // laid out below the visible area
}

// A failed lookup lists what is visible.
func TestFailureListsElements(t *testing.T) {
	ft := &fakeT{TB: t}
	app := windowtest.New(ft, window.Options{Content: el.Root(&greeter{name: kit.Input("Name")})})
	func() {
		defer func() { _ = recover() }()
		app.Click("Missing")
	}()
	if !strings.Contains(ft.msg, `no element matches "Missing"`) || !strings.Contains(ft.msg, `button "Say hello"`) || !strings.Contains(ft.msg, `textbox "Name"`) {
		t.Fatalf("message %q", ft.msg)
	}
}

// fakeT records a fatal message instead of ending the test.
type fakeT struct {
	testing.TB
	msg string
}

func (f *fakeT) Fatalf(format string, args ...any) {
	f.msg = fmt.Sprintf(format, args...)
	panic("fatal")
}
