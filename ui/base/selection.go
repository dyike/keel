package base

// Selection is a set of chosen items, keyed so it survives reordering and
// filtering, with an anchor for Shift ranges. The zero value is empty.
type Selection[K comparable] struct {
	set       map[K]bool
	anchor    K
	hasAnchor bool
}

func (s *Selection[K]) Has(k K) bool { return s.set[k] }
func (s *Selection[K]) Len() int     { return len(s.set) }

// Anchor is where a Shift range starts: the item last clicked without Shift.
func (s *Selection[K]) Anchor() (K, bool) { return s.anchor, s.hasAnchor }

// Set replaces the selection with keys and anchors at the last, as if they
// had been clicked in order.
func (s *Selection[K]) Set(keys ...K) {
	s.set = make(map[K]bool, len(keys))
	for _, k := range keys {
		s.set[k] = true
	}
	var zero K
	s.anchor, s.hasAnchor = zero, len(keys) > 0
	if s.hasAnchor {
		s.anchor = keys[len(keys)-1]
	}
}

// Keep drops selected keys that keep rejects, such as items that were
// removed or disabled.
func (s *Selection[K]) Keep(keep func(K) bool) {
	for k := range s.set {
		if !keep(k) {
			delete(s.set, k)
		}
	}
	if s.hasAnchor && !keep(s.anchor) {
		var zero K
		s.anchor, s.hasAnchor = zero, false
	}
}

// Click applies a click on item i of order, with the platform's rules:
// a plain click selects only i; toggle (Cmd, or Ctrl off macOS) adds or
// removes i; extend (Shift) selects from the anchor to i, added to the
// selection when toggle is also held. Disabled items in a range are
// skipped. A disabled i changes nothing.
func (s *Selection[K]) Click(order []K, i int, extend, toggle bool, disabled func(int) bool) {
	if i < 0 || i >= len(order) || (disabled != nil && disabled(i)) {
		return
	}
	a := -1
	if s.hasAnchor {
		for j, k := range order {
			if k == s.anchor {
				a = j
				break
			}
		}
	}
	if !toggle || s.set == nil {
		s.set = map[K]bool{}
	}
	if extend && a >= 0 {
		for j := min(a, i); j <= max(a, i); j++ {
			if disabled == nil || !disabled(j) {
				s.set[order[j]] = true
			}
		}
		return
	}
	k := order[i]
	if toggle && s.set[k] {
		delete(s.set, k)
	} else {
		s.set[k] = true
	}
	s.anchor, s.hasAnchor = k, true
}

// In returns the selected keys in the order of order.
func (s *Selection[K]) In(order []K) []K {
	var out []K
	for _, k := range order {
		if s.set[k] {
			out = append(out, k)
		}
	}
	return out
}

// Indexes returns the positions in order of the selected keys.
func (s *Selection[K]) Indexes(order []K) []int {
	var out []int
	for i, k := range order {
		if s.set[k] {
			out = append(out, i)
		}
	}
	return out
}
