package el

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/dyike/keel/ui/internal/uitest"
)

// Changing virtual rows creates new keyed state while old rows leave the tree.
func BenchmarkKeyedRowState(b *testing.B) {
	offset := 0
	root := Root(viewFunc(func(*Context) Element {
		box := Div()
		for i := range 60 {
			box.Child(Div().ID(strconv.Itoa(offset + i)).H(Dp(18)))
		}
		return box
	}))
	h := uitest.New(root)

	b.ReportAllocs()
	for b.Loop() {
		offset += 3
		h.Frame()
	}
	b.ReportMetric(float64(unsafe.Sizeof(elemState{})), "state-B")
}

func TestKeyedStateAllocatesEditingOnlyForInputs(t *testing.T) {
	showInput := false
	value := "preserved"
	root := Root(viewFunc(func(*Context) Element {
		box := Div().ID("rows")
		for i := range 20 {
			box.Child(Div().ID(strconv.Itoa(i)).H(Dp(18)).Disabled(i%2 == 0))
		}
		if showInput {
			box.Child(Input().ID("editor").Bind(&value))
		}
		return box
	}))
	h := uitest.New(root)
	for _, st := range root.store.states {
		if st.inputState != nil || st.editor != nil {
			t.Fatal("non-input allocated editing state")
		}
	}
	showInput = true
	h.Frame()
	var input *elemState
	for _, st := range root.store.states {
		if st.id == "editor" {
			input = st
		}
	}
	if input == nil || input.inputState == nil || input.editor != &input.inputState.editorValue {
		t.Fatal("missing stable embedded editor")
	}
	editor, editing := input.editor, input.inputState
	for range 3 {
		h.Frame()
	}
	if input.editor != editor || input.inputState != editing || value != "preserved" {
		t.Fatal("editing state changed across frames")
	}
	showInput = false
	h.Frame()
	for _, st := range root.store.states {
		if st == input {
			t.Fatal("removed input retained by state store")
		}
	}
}

var measuredState *elemState

func BenchmarkElementStateCreation(b *testing.B) {
	for _, input := range []bool{false, true} {
		name := "ordinary"
		if input {
			name = "input"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				st := new(elemState)
				if input {
					st.inputEditor()
				}
				measuredState = st
			}
			measuredState = nil
		})
	}
}
