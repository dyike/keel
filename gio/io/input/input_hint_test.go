// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"testing"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
)

func TestInputHintWithoutFocusInLargeHandlerTable(t *testing.T) {
	handlers := make(map[event.Tag]*handler)
	for i := range 4096 {
		tag := new(int)
		*tag = i
		handlers[tag] = &handler{}
	}
	q := keyQueue{hint: key.HintEmail}
	if hint, changed := q.InputHint(handlers, keyState{}); hint != key.HintEmail || changed {
		t.Fatalf("unfocused hint changed: %v, %v", hint, changed)
	}
	tag := new(int)
	handlers[tag] = &handler{key: keyHandler{hint: key.HintNumeric}}
	state := keyState{focus: tag}
	if hint, changed := q.InputHint(handlers, state); hint != key.HintNumeric || !changed {
		t.Fatalf("focus did not change hint: %v, %v", hint, changed)
	}
	if _, changed := q.InputHint(handlers, state); changed {
		t.Fatal("unchanged focus repeats notification")
	}
	if hint, changed := q.InputHint(handlers, keyState{}); hint != key.HintNumeric || changed {
		t.Fatalf("losing focus changed remembered hint: %v, %v", hint, changed)
	}
}
