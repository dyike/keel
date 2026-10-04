package core

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/dyike/keel/ui/internal/loop"
)

// The keymap binds named actions, such as "editor.save", to key chords. Views
// handle actions by name (el's Context.Action) and show their keys by name
// (kit.KbdFor), so users can rebind a key in one place and every handler,
// menu and hint follows on the next frame.
var keymap = struct {
	sync.Mutex
	bindings map[string][]string
	contexts map[string]map[string][]string
	// Compiled context predicates, and the order bindings were made in, so
	// the latest of several matching at one level wins.
	predicates map[string]predicate
	order      map[[2]string]uint64
	seq        uint64
}{bindings: map[string][]string{}}

// Bind sets the chords that trigger action, replacing its earlier ones; no
// chords unbinds it. Chords use ParseShortcut's syntax, e.g. "mod+s". It
// returns an error, and changes nothing, if a chord is invalid. Every window
// redraws.
func Bind(action string, chords ...string) error {
	for _, c := range chords {
		if _, _, err := ParseShortcut(c); err != nil {
			return fmt.Errorf("bind %s: %w", action, err)
		}
	}
	keymap.Lock()
	if len(chords) == 0 {
		delete(keymap.bindings, action)
	} else {
		keymap.bindings[action] = slices.Clone(chords)
	}
	keymap.Unlock()
	loop.InvalidateAll()
	return nil
}

// Bindings returns the chords bound to action, the first being the one to
// show in hints; nil if it has none.
func Bindings(action string) []string {
	keymap.Lock()
	defer keymap.Unlock()
	return slices.Clone(keymap.bindings[action])
}

// Keymap returns a copy of every binding, e.g. for a settings page.
func Keymap() map[string][]string {
	keymap.Lock()
	defer keymap.Unlock()
	out := make(map[string][]string, len(keymap.bindings))
	for a, c := range keymap.bindings {
		out[a] = slices.Clone(c)
	}
	return out
}

// LoadKeymap binds the actions in a JSON object such as
// {"editor.save": ["mod+s"], "app.quit": []} on top of the current keymap;
// an empty list unbinds. Nothing changes if any entry is invalid.
func LoadKeymap(data []byte) error {
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("keymap: %w", err)
	}
	for _, a := range slices.Sorted(maps.Keys(m)) {
		for _, c := range m[a] {
			if _, _, err := ParseShortcut(c); err != nil {
				return fmt.Errorf("keymap %s: %w", a, err)
			}
		}
	}
	keymap.Lock()
	for a, c := range m {
		if len(c) == 0 {
			delete(keymap.bindings, a)
		} else {
			keymap.bindings[a] = slices.Clone(c)
		}
	}
	keymap.Unlock()
	loop.InvalidateAll()
	return nil
}
