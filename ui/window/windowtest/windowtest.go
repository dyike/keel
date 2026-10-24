// Package windowtest tests a Keel app's views the way a person uses them: it
// draws a window in memory, finds elements by their visible text and role,
// clicks, types and presses keys. No native window opens, so tests run in
// CI and in parallel with nothing on screen.
//
//	func TestGreeting(t *testing.T) {
//		app := windowtest.New(t, window.Options{Content: el.Root(newApp())})
//		app.Click("Name")
//		app.Type("Ada")
//		app.Click("Say hello")
//		app.Want("Hello, Ada!")
//	}
//
// Elements are what keel-mcp reports to agents; see docs/automation.md for
// roles and values. Reduced motion is on, so nothing is caught mid-animation.
package windowtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/window"
)

// App is a window drawn in memory, failing its test when an action cannot
// be performed. Every query and action renders first, so it sees functions
// already passed to core.Update, such as the result of a finished goroutine.
type App struct {
	t testing.TB
	h *window.Headless
}

// New opens o in memory for the duration of the test. Width and Height
// default to 800×600 dp.
func New(t testing.TB, o window.Options) *App {
	t.Helper()
	a := &App{t: t, h: window.OpenHeadless(o)}
	t.Cleanup(a.h.Close)
	a.h.Render()
	return a
}

// Headless is the underlying window, for actions that should not fail the
// test, such as probing whether an element exists.
func (a *App) Headless() *window.Headless { return a.h }

// Window is the window being drawn.
func (a *App) Window() *window.Window { return a.h.Window() }

// Render runs queued core.Update functions and draws until the window is
// settled. Queries and actions render on their own; call it to run OnClose
// after Window().Close, say.
func (a *App) Render() { a.h.Render() }

// Elements lists what the window shows.
func (a *App) Elements() []window.Element { return a.h.Elements() }

// Find returns the element best matching text by name or value: an exact
// match over a partial one, a control over plain text.
func (a *App) Find(text string) window.Element {
	a.t.Helper()
	e, err := a.h.Find(text)
	if err != nil {
		a.t.Fatalf("%v\n%s", err, a.describe())
	}
	return e
}

// Has reports whether an element matches text.
func (a *App) Has(text string) bool {
	_, err := a.h.Find(text)
	return err == nil
}

// Want fails the test unless an element matches text.
func (a *App) Want(text string) {
	a.t.Helper()
	if !a.Has(text) {
		a.t.Fatalf("no element matches %q\n%s", text, a.describe())
	}
}

// WantNot fails the test if an element matches text.
func (a *App) WantNot(text string) {
	a.t.Helper()
	if e, err := a.h.Find(text); err == nil {
		a.t.Fatalf("unexpected %s %q\n%s", e.Role, e.Name+e.Value, a.describe())
	}
}

// Click clicks the element best matching text.
func (a *App) Click(text string) {
	a.t.Helper()
	a.h.ClickElement(a.Find(text))
}

// Type inserts s into the focused text box; click the text box first.
func (a *App) Type(s string) {
	a.t.Helper()
	if err := a.h.Type(s); err != nil {
		a.t.Fatalf("type %q: %v\n%s", s, err, a.describe())
	}
}

// Press sends a key chord such as "enter", "tab" or "mod+a".
func (a *App) Press(chord string) {
	a.t.Helper()
	if err := a.h.Press(chord); err != nil {
		a.t.Fatalf("press %q: %v", chord, err)
	}
}

// Scroll turns the mouse wheel by dy dp (positive: down) over the element
// best matching text.
func (a *App) Scroll(text string, dy float32) {
	a.t.Helper()
	if err := a.h.Scroll(text, dy); err != nil {
		a.t.Fatalf("scroll %q: %v\n%s", text, err, a.describe())
	}
}

// Resize changes the window's size in dp.
func (a *App) Resize(width, height int) { a.h.Resize(width, height) }

// Screenshot renders the window with the GPU as PNG. It skips the test on
// machines without a usable GPU.
func (a *App) Screenshot() []byte {
	a.t.Helper()
	png, err := a.h.Screenshot()
	if err != nil {
		a.t.Skipf("screenshot: %v", err)
	}
	return png
}

// describe lists the visible elements, to explain a failed lookup.
func (a *App) describe() string {
	var b strings.Builder
	b.WriteString("visible elements:")
	for _, e := range a.h.Elements() {
		fmt.Fprintf(&b, "\n  %s %q", e.Role, e.Name)
		if e.Value != "" {
			fmt.Fprintf(&b, " value=%q", e.Value)
		}
		if e.Disabled {
			b.WriteString(" disabled")
		}
	}
	return b.String()
}
