package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dyike/keel/ui/internal/loop"
)

// BindIn sets action's chords where a key context predicate holds: a name
// such as "Editor", or an expression such as "Editor && !ReadOnly" or
// "Pane > Editor" (see keymap_predicate.go). No chords explicitly disables
// the action there, hiding outer and global bindings. An invalid chord or
// predicate leaves the map unchanged.
func BindIn(context, action string, chords ...string) error {
	if strings.TrimSpace(context) == "" {
		return fmt.Errorf("keymap: empty context")
	}
	pred, err := parsePredicate(context)
	if err != nil {
		return err
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
	if keymap.predicates == nil {
		keymap.predicates = map[string]predicate{}
		keymap.order = map[[2]string]uint64{}
	}
	keymap.predicates[context] = pred
	keymap.seq++
	keymap.order[[2]string{context, action}] = keymap.seq
	keymap.Unlock()
	loop.InvalidateAll()
	return nil
}

// ClearBindingIn removes a contextual override, restoring inheritance.
func ClearBindingIn(context, action string) {
	keymap.Lock()
	delete(keymap.contexts[context], action)
	delete(keymap.order, [2]string{context, action})
	if len(keymap.contexts[context]) == 0 {
		delete(keymap.contexts, context)
		delete(keymap.predicates, context)
	}
	keymap.Unlock()
	loop.InvalidateAll()
}

// BindingsIn resolves action along a focus path of KeyContext names, inner
// to outer: the innermost level where some binding's predicate holds wins,
// and among bindings matching at that level, the one bound last. With none,
// the global keymap applies. The result is owned by the caller. An explicit
// empty binding prevents fallback.
func BindingsIn(action string, contexts ...string) []string {
	keymap.Lock()
	defer keymap.Unlock()
	levels := contextLevels(contexts)
	for at := range levels {
		var best []string
		var bestSeq uint64
		for context, actions := range keymap.contexts {
			chords, found := actions[action]
			if !found {
				continue
			}
			if p := keymap.predicates[context]; p != nil && p.match(levels, at) {
				if seq := keymap.order[[2]string{context, action}]; best == nil || seq > bestSeq {
					best, bestSeq = chords, seq
					if best == nil {
						best = []string{} // an explicit disable still wins
					}
				}
			}
		}
		if best != nil {
			return slices.Clone(best)
		}
	}
	return slices.Clone(keymap.bindings[action])
}
