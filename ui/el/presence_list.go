package el

// PresenceListOptions coordinates entries of a keyed collection. ExitBeforeEnter
// holds new keys until all removed keys have finished exiting. Surviving keys
// remain interactive. Reintroducing an exiting key reverses it immediately.
type PresenceListOptions struct {
	PresenceOptions
	ExitBeforeEnter bool
}

// PresentItem includes the last value of a removed item until its exit finishes.
// Apply Frame through element.Presence to suppress input during that exit.
type PresentItem[T any] struct {
	Key   any
	Value T
	Frame PresenceFrame
}

type presenceListKey struct{ group, item any }
type presenceEntry[T any] struct {
	key           any
	value         T
	seen, entered bool
}
type presenceCollection interface{ epoch() uint64 }
type presenceList[T any] struct {
	frame   uint64
	entries []*presenceEntry[T]
	byKey   map[any]*presenceEntry[T]
	output  []PresentItem[T]
}

func (s *presenceList[T]) epoch() uint64 { return s.frame }

// PresentItems retains removed values while they animate out. Call it once per
// Render with a root-unique comparable group key and unique comparable item IDs.
// The returned slice is borrowed until the next call. State belongs to the root;
// omitting the group cancels its exits and drops its retained values and callbacks.
// Empty input must still be declared to finish exits. No element trees are saved.
// Reordering keeps existing order until exits finish; give elements stable IDs.
func PresentItems[T any](cx *Context, key any, items []T, id func(T) any, options PresenceListOptions) []PresentItem[T] {
	r := cx.root
	var state *presenceList[T]
	if previous := r.presenceLists[key]; previous != nil {
		var ok bool
		state, ok = previous.(*presenceList[T])
		if !ok {
			panic("el: presence collection key reused with a different value type")
		}
	}
	if state == nil && len(items) == 0 {
		return nil
	}
	live := r.e.gtx.Enabled()
	if state == nil {
		state = &presenceList[T]{byKey: make(map[any]*presenceEntry[T])}
	} else if !live {
		copy := &presenceList[T]{byKey: make(map[any]*presenceEntry[T], len(state.byKey))}
		for _, entry := range state.entries {
			e := *entry
			copy.entries = append(copy.entries, &e)
			copy.byKey[e.key] = &e
		}
		state = copy
	}
	if live {
		if r.presenceLists == nil {
			r.presenceLists = make(map[any]presenceCollection)
		}
		if state.frame == r.presenceEpoch {
			panic("el: presence collection declared twice in one build")
		}
		r.presenceLists[key] = state
	}
	state.frame = r.presenceEpoch
	for _, entry := range state.entries {
		entry.seen = false
	}
	for _, item := range items {
		itemKey := id(item)
		entry := state.byKey[itemKey]
		if entry == nil {
			entry = &presenceEntry[T]{key: itemKey}
			state.byKey[itemKey] = entry
			state.entries = append(state.entries, entry)
		}
		if entry.seen {
			panic("el: duplicate presence item key")
		}
		entry.value, entry.seen = item, true
	}
	oldOutput := state.output
	state.output = state.output[:0]
	exiting := false
	// Sample departures first, so an ordered arrival can start in the same
	// frame in which the last departure finishes, without an extra wakeup.
	for _, entry := range state.entries {
		if entry.seen || !entry.entered {
			continue
		}
		p := cx.Presence(presenceListKey{key, entry.key}, false, options.PresenceOptions)
		if p.Visible {
			exiting = true
			state.output = append(state.output, PresentItem[T]{entry.key, entry.value, p})
		} else {
			delete(state.byKey, entry.key)
		}
	}
	for _, entry := range state.entries {
		if !entry.seen {
			if !entry.entered {
				delete(state.byKey, entry.key)
			}
			continue
		}
		if !entry.entered && options.ExitBeforeEnter && exiting {
			continue
		}
		entry.entered = true
		p := cx.Presence(presenceListKey{key, entry.key}, true, options.PresenceOptions)
		state.output = append(state.output, PresentItem[T]{entry.key, entry.value, p})
	}
	// Keep source ordering when no departures remain. During exit, old entries
	// retain their slots instead of jumping past newcomers.
	if !exiting {
		for i, item := range items {
			entry := state.byKey[id(item)]
			state.entries[i] = entry
		}
		clear(state.entries[len(items):])
		state.entries = state.entries[:len(items)]
	} else {
		write := 0
		for _, entry := range state.entries {
			if state.byKey[entry.key] != nil {
				state.entries[write] = entry
				write++
			}
		}
		clear(state.entries[write:])
		state.entries = state.entries[:write]
	}
	// Output follows the stable entry order, including ghosts in their slots.
	write := 0
	for _, entry := range state.entries {
		for j := write; j < len(state.output); j++ {
			if state.output[j].Key == entry.key {
				state.output[write], state.output[j] = state.output[j], state.output[write]
				write++
				break
			}
		}
	}
	// Release any stale references in the reused output backing array.
	if len(oldOutput) > len(state.output) {
		clear(oldOutput[len(state.output):])
	}
	if len(state.entries) == 0 && live {
		delete(r.presenceLists, key)
	}
	return state.output
}
