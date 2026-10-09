package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestNumberInputManagedStepDraftAndModes(t *testing.T) {
	var events []NumberStepEvent
	changes, strategies := 0, 0
	n := NumberInput("value").Range(0, 100).StepBy(func(float64, NumberStepAction) float64 { strategies++; return 2 }).OnChange(func(float64) { changes++ })
	n.SetValue(4)
	n.OnStep(func(e NumberStepEvent) { events = append(events, e) })
	n.text = "9.5"
	n.move(-10)
	if len(events) != 1 || events[0] != (NumberStepEvent{9.5, NumberStepActionDecrement, 10}) || n.Value() != 4 || n.text != "9.5" || changes != 0 || strategies != 0 {
		t.Fatal("event committed draft", events, n.Value(), n.text, changes, strategies)
	}
	n.text = "-"
	n.move(1)
	if events[1].Value != 4 || n.text != "-" {
		t.Fatal("invalid draft fallback")
	}
	n.OnStep(func(e NumberStepEvent) { n.SetValue(e.Value + 3) })
	n.move(1)
	if n.Value() != 7 || changes != 0 {
		t.Fatal("callback update overwritten")
	}
	n.OnStep(nil)
	n.move(1)
	if n.Value() != 9 || strategies != 1 || changes != 1 {
		t.Fatal("strategy not restored")
	}
	n.OnStep(func(NumberStepEvent) { t.Fatal("old handler") }).Step(1)
	n.move(1)
	if n.Value() != 10 {
		t.Fatal("Step did not leave event mode")
	}
	n.OnStep(func(NumberStepEvent) { t.Fatal("old handler") }).StepBy(nil)
	n.move(1)
	if n.Value() != 11 {
		t.Fatal("StepBy did not leave event mode")
	}
}

func TestNumberInputManagedStepControlsAndDisabled(t *testing.T) {
	var events []NumberStepEvent
	ancestor := false
	n := NumberInput("value").Range(0, 20).OnStep(func(e NumberStepEvent) { events = append(events, e) })
	n.SetValue(5)
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(ancestor).Child(n.Render(cx)) })
	loc := locale.Current()
	click(t, h, loc.Name(loc.Increase, "value"))
	click(t, h, loc.Name(loc.Decrease, "value"))
	clickClass(t, h, "Editor", "value")
	h.Key(key.NamePageUp, 0)
	h.Key(key.NamePageDown, 0)
	if len(events) != 4 || events[0].Action != NumberStepActionIncrement || events[1].Action != NumberStepActionDecrement || events[2].Count != 10 || events[3].Count != 10 || n.Value() != 5 {
		t.Fatal("controls", events, n.Value())
	}
	n.SetValue(20)
	h.Frame()
	h.Key(key.NameUpArrow, 0)
	if len(events) != 4 {
		t.Fatal("keyboard stepped past boundary")
	}
	ancestor = true
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	if len(events) != 4 {
		t.Fatal("ancestor disabled step")
	}
	n.SetDisabled(true)
	n.move(-1)
	if len(events) != 4 {
		t.Fatal("disabled step")
	}
}
