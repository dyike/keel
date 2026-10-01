package widget

import (
	"math"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestSliderPointerKeyboardAndDisabled(t *testing.T) {
	var changes []float32
	s := Slider("volume", 0, 100).Step(5).OnChange(func(v float32) { changes = append(changes, v) })
	h := uitest.New(s)
	h.Click(200, 35)
	if s.Value() != 50 || len(changes) != 1 {
		t.Fatalf("track click: value=%v changes=%v", s.Value(), changes)
	}
	h.Key(key.NameRightArrow, 0)
	if s.Value() != 55 {
		t.Fatalf("right key value %v", s.Value())
	}
	h.Key(key.NamePageDown, 0)
	if s.Value() != 5 {
		t.Fatalf("page down value %v", s.Value())
	}
	h.Key(key.NameEnd, 0)
	if s.Value() != 100 {
		t.Fatal("end key did not reach max")
	}
	h.Key(key.NameHome, 0)
	if s.Value() != 0 {
		t.Fatal("home key did not reach min")
	}
	h.Drag(200, 35, 550, 35)
	if s.Value() != 100 {
		t.Fatalf("drag outside track stopped at %v", s.Value())
	}
	before := len(changes)
	s.SetValue(45)
	h.Frame()
	if len(changes) != before {
		t.Fatal("setter triggered callback")
	}
	s.SetDisabled(true)
	h.Frame()
	h.Click(40, 35)
	h.Key(key.NameLeftArrow, 0)
	if s.Value() != 45 || len(changes) != before {
		t.Fatal("disabled slider changed")
	}
}
func TestSliderFractionalStepsAndBounds(t *testing.T) {
	s := Slider("fraction", -1, 1).Step(.1)
	h := uitest.New(s)
	h.Click(200, 35)
	for i := 0; i < 10; i++ {
		h.Key(key.NameRightArrow, 0)
	}
	if s.Value() != 1 {
		t.Fatalf("fractional steps stopped at %v", s.Value())
	}
	s.SetRange(0, 10)
	s.Step(3)
	s.SetValue(10)
	h.Frame()
	h.Key(key.NameLeftArrow, 0)
	if s.Value() != 9 {
		t.Fatalf("non-aligned endpoint skipped step: %v", s.Value())
	}
	s.SetValue(float32(math.NaN()))
	if s.Value() != 0 {
		t.Fatal("NaN escaped into layout")
	}
	s.SetRange(3, 3)
	h.Frame()
	h.Click(80, 35)
	h.Key(key.NameEnd, 0)
	if s.Value() != 3 {
		t.Fatal("constant range changed")
	}
}
