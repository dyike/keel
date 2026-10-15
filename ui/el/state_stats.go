package el

import (
	"log"
	"os"
	"sync/atomic"
)

// Opt-in diagnostics, not a redraw source. Root IDs distinguish recreated
// roots from unstable keys inside a retained root. New rows are legitimate
// state creation too; a miss alone is not proof of a bad identity.
var stateStatsEnabled = os.Getenv("KEEL_STATE_STATS") == "1"
var stateStatsSequence atomic.Uint64
var stateStatsLog = log.Printf

type stateLabel struct{ id, role, name, kind string }
type stateCreation struct {
	key   stateKey
	label stateLabel
}
type stateStats struct {
	id           uint64
	labels       map[stateKey]stateLabel
	new, removed int
	first        []stateCreation
}

func labelState(n *Node) stateLabel {
	kind := "div"
	switch {
	case n.input != nil:
		kind = "input"
	case n.widget != nil:
		kind = "widget"
	case n.isText:
		kind = "text"
	}
	name := n.name
	if name == "" && n.isText {
		name = n.text
	}
	return stateLabel{n.id, n.role, name, kind}
}

func (d *stateStats) note(key stateKey) {
	d.new++
	if len(d.first) < 8 {
		d.first = append(d.first, stateCreation{key, d.labels[key]})
	}
}

func (s *store) beginStateStats() {
	if d := s.stats; d != nil {
		d.new, d.removed = 0, 0
		clear(d.labels)
		clear(d.first)
		d.first = d.first[:0]
	}
}

func (s *store) endStateStats() {
	if d := s.stats; d != nil {
		if d.new != 0 || d.removed != 0 {
			stateStatsLog("el: state root=%d frame=%d new=%d removed=%d retained=%d total-created=%d", d.id, s.frame, d.new, d.removed, len(s.states), s.created)
			for _, c := range d.first {
				stateStatsLog("el: state root=%d key=%016x kind=%s id=%q role=%q name=%q", d.id, c.key, c.label.kind, c.label.id, c.label.role, c.label.name)
			}
		}
		// Diagnostic strings must not extend a frame element's lifetime.
		clear(d.labels)
		clear(d.first)
		d.first = d.first[:0]
	}
}
