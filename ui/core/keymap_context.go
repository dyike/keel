package core

import (
	"fmt"
	"slices"

	"github.com/dyike/keel/ui/internal/loop"
)

// BindIn sets action's chords in a named key context. No chords explicitly
// disables the action there, hiding outer/global bindings. Invalid chords
// leave the map unchanged. Contexts are exact names, not predicate expressions.
func BindIn(context, action string, chords ...string) error {
	if context == "" {
		return fmt.Errorf("keymap: empty context")
	}
	for _, chord := range chords {
		if _, _, err := ParseShortcut(chord); err != nil {
			return fmt.Errorf("bind %s in %s: %w", action, context, err)
		}
	}
	keymap.Lock()
	if keymap.contexts == nil {
		keymap.contexts = map[string]map[string][]string{}
	}
	if keymap.contexts[context] == nil {
		keymap.contexts[context] = map[string][]string{}
	}
	keymap.contexts[context][action] = slices.Clone(chords)
	keymap.Unlock()
	loop.InvalidateAll()
	return nil
}

// ClearBindingIn removes a contextual override, restoring inheritance.
func ClearBindingIn(context, action string) {
	keymap.Lock()
	delete(keymap.contexts[context], action)
	if len(keymap.contexts[context]) == 0 {
		delete(keymap.contexts, context)
	}
	keymap.Unlock()
	loop.InvalidateAll()
}

// BindingsIn resolves the nearest explicit binding, checking contexts in the
// supplied order (inner to outer), then the global keymap. The result is owned
// by the caller. An explicit empty binding prevents fallback.
func BindingsIn(action string, contexts ...string) []string {
	keymap.Lock()
	defer keymap.Unlock()
	for _, context := range contexts {
		if chords, found := keymap.contexts[context][action]; found {
			return slices.Clone(chords)
		}
	}
	return slices.Clone(keymap.bindings[action])
}
