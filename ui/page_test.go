package ui

import (
	"errors"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend/soft"
	"github.com/dyike/keel/capability"
	"testing"
)

func render(t *testing.T, w *gui.Window) {
	t.Helper()
	if _, err := soft.RenderToImage(w, 1); err != nil {
		t.Fatal(err)
	}
}
func TestNativeEventsAndState(t *testing.T) {
	input := Input(InputOptions{Label: "Name", MaxLength: 8})
	check := Checkbox(CheckboxOptions{Label: "Welcome"})
	output := Text("Ready")
	button := Button("Greet", func() { output.SetText("Hello " + input.Value()) })
	reset := Button("Reset", func() { input.SetValue(""); check.SetValue(false); output.SetText("Ready") })
	page := NewPage("UI", Column(input, check, Row(button, reset), output))
	w := gui.NewWindow(gui.WindowCfg{Width: 640, Height: 420})
	defer w.WindowCleanup()
	if err := page.Attach(w); err != nil {
		t.Fatal(err)
	}
	render(t, w)
	if err := w.TestType(input.ID(), "Keel"); err != nil {
		t.Fatal(err)
	}
	render(t, w)
	if err := w.TestClick(check.ID()); err != nil {
		t.Fatal(err)
	}
	if err := w.TestClick(button.ID()); err != nil {
		t.Fatal(err)
	}
	if input.Value() != "Keel" || !check.Value() || output.Text() != "Hello Keel" {
		t.Fatal(input.Value(), check.Value(), output.Text())
	}
	render(t, w)
	if err := w.TestClick(reset.ID()); err != nil {
		t.Fatal(err)
	}
	render(t, w)
	if input.Value() != "" || check.Value() || output.Text() != "Ready" {
		t.Fatal("reset did not reach controls")
	}
	if err := w.TestType(input.ID(), "123456789"); err != nil {
		t.Fatal(err)
	}
	if input.Value() != "12345678" {
		t.Fatal(input.Value())
	}
}
func TestRequiredDisabledReadOnlyAndPanic(t *testing.T) {
	field := Input(InputOptions{Label: "Required", Required: true})
	read := Input(InputOptions{Label: "Locked", Value: "original", ReadOnly: true})
	calls := 0
	button := Button("Save", func() { calls++ })
	panics := Button("Panic", func() { panic("secret") })
	page := NewPage("validation", field, read, button, panics)
	w := gui.NewWindow(gui.WindowCfg{Width: 640, Height: 480})
	defer w.WindowCleanup()
	_ = page.Attach(w)
	var received error
	page.SetOnError(func(err error) { received = err })
	render(t, w)
	_ = w.TestClick(button.ID())
	if calls != 0 || received == nil {
		t.Fatal("required check did not block callback")
	}
	_ = w.TestType(read.ID(), "change")
	if read.Value() != "original" {
		t.Fatal(read.Value())
	}
	field.SetValue("value")
	page.Refresh()
	render(t, w)
	button.SetDisabled(true)
	page.Refresh()
	render(t, w)
	_ = w.TestClick(button.ID())
	if calls != 0 {
		t.Fatal("disabled button fired")
	}
	_ = w.TestClick(panics.ID())
	if received == nil || received.Error() != "ui: event callback panicked" {
		t.Fatal(received)
	}
}
func TestPageOwnershipAndValidation(t *testing.T) {
	same := Text("a")
	var missing *InputComponent
	for _, page := range []*Page{NewPage("x", nil), NewPage("x", missing), NewPage("x", same, same), NewPage("x", Input(InputOptions{MaxLength: -1})), NewPage("x", Input(InputOptions{Type: "unknown"}))} {
		if page.Validate() == nil {
			t.Fatal("invalid component tree accepted")
		}
	}
	page := NewPage("x", Text("one"))
	a := gui.NewWindow(gui.WindowCfg{})
	defer a.WindowCleanup()
	b := gui.NewWindow(gui.WindowCfg{})
	defer b.WindowCleanup()
	if err := page.Attach(a); err != nil {
		t.Fatal(err)
	}
	if err := page.Attach(b); !errors.Is(err, capability.ErrConflict) {
		t.Fatal(err)
	}
	page.Close()
	if err := page.Validate(); !errors.Is(err, capability.ErrClosed) {
		t.Fatal(err)
	}
}
func TestIndependentPagesAndBackgroundRefresh(t *testing.T) {
	one, two := Input(InputOptions{Value: "one"}), Input(InputOptions{Value: "two"})
	p1, p2 := NewPage("one", one), NewPage("two", two)
	w1 := gui.NewWindow(gui.WindowCfg{Width: 400, Height: 250})
	defer w1.WindowCleanup()
	w2 := gui.NewWindow(gui.WindowCfg{Width: 400, Height: 250})
	defer w2.WindowCleanup()
	_ = p1.Attach(w1)
	_ = p2.Attach(w2)
	render(t, w1)
	render(t, w2)
	done := make(chan struct{})
	go func() { one.SetValue("worker"); p1.Refresh(); close(done) }()
	<-done
	render(t, w1)
	render(t, w2)
	if one.Value() != "worker" || two.Value() != "two" {
		t.Fatal("window state leaked")
	}
}
