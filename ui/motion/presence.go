package motion

import (
	"time"

	"github.com/dyike/keel/ui/el"
)

// PresenceItem is a currently present or animating-out key.
type PresenceItem[K comparable] struct {
	Key      K
	Progress float32
	Exiting  bool
	Disabled bool
}

type presenceState[K comparable] struct {
	key     K
	value   Value
	seen    bool
	exiting bool
}

// Presence keeps removed keys alive until their exit transition settles.
// The returned slice is reused and is valid only until the next Items call.
type Presence[K comparable] struct {
	states []*presenceState[K]
	byKey  map[K]*presenceState[K]
	items  []PresenceItem[K]
}

func (p *Presence[K]) Items(cx *el.Context, keys []K, enter, exit Transition) []PresenceItem[K] {
	items, animate := p.itemsAt(cx.Now(), keys, enter, exit, el.ReducedMotion())
	if animate {
		cx.Animating()
	}
	return items
}

func (p *Presence[K]) itemsAt(now time.Time, keys []K, enter, exit Transition, reduced bool) ([]PresenceItem[K], bool) {
	if p.byKey == nil {
		p.byKey = make(map[K]*presenceState[K], len(keys))
	}
	for _, state := range p.states {
		state.seen = false
	}
	for _, key := range keys {
		state := p.byKey[key]
		if state == nil {
			state = &presenceState[K]{key: key}
			state.value.Set(0)
			p.byKey[key] = state
			p.states = append(p.states, state)
		}
		state.seen, state.exiting = true, false
	}
	p.items = p.items[:0]
	write := 0
	animate := false
	for _, state := range p.states {
		state.exiting = !state.seen
		target, tr := float32(1), enter
		if state.exiting {
			target, tr = 0, exit
		}
		progress, active := state.value.toAt(now, target, tr, reduced)
		animate = animate || active
		if state.exiting && !state.value.Animating() && progress == 0 {
			delete(p.byKey, state.key)
			continue
		}
		p.states[write] = state
		write++
		p.items = append(p.items, PresenceItem[K]{state.key, progress, state.exiting, state.exiting})
	}
	p.states = p.states[:write]
	return p.items, animate
}
