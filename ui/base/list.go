// Package base holds the behavior of Keel's components without their look:
// keyboard navigation over a list, type-to-find, multi-selection with
// ranges, and open/closed state. It draws nothing and depends on no other
// Keel package. kit builds its components from it; build your own with base
// for the behavior and el for the look.
package base

import "github.com/dyike/keel/third_party/gio/io/key"

// List navigates Count items, some of which may be disabled. The zero
// Disabled enables every item. Indexes count from 0; -1 means none.
type List struct {
	Count    int
	Disabled func(i int) bool
	// Wrap moves from the last item to the first and back, as menus do.
	Wrap bool
	// Page is how far PageUp and PageDown move; 10 by default.
	Page int
}

func (l List) enabled(i int) bool {
	return i >= 0 && i < l.Count && (l.Disabled == nil || !l.Disabled(i))
}

// First is the first enabled item, or -1.
func (l List) First() int { return l.Next(-1, 1) }

// Last is the last enabled item, or -1.
func (l List) Last() int { return l.Next(l.Count, -1) }

// Next is the nearest enabled item after i in direction dir (1 or -1). It
// wraps when Wrap is set; otherwise, and when no item is enabled, it
// returns i if i is enabled and -1 if not.
func (l List) Next(i, dir int) int {
	n := l.Count
	if l.Wrap && (i < 0 || i >= n) { // from outside: start at the near end
		i = -1
		if dir < 0 {
			i = n
		}
	}
	for k := 1; k <= n; k++ {
		j := i + dir*k
		if l.Wrap {
			j = ((j % n) + n) % n
		} else if j < 0 || j >= n {
			break
		}
		if l.enabled(j) {
			return j
		}
	}
	if l.enabled(i) {
		return i
	}
	return -1
}

// Key moves from i for ↑ ↓ Home End PageUp PageDown and reports whether
// name was one of them. From no item (-1), ↓ goes to the first and ↑ to the
// last. Disabled items are skipped; a page move that lands on one goes on
// in the same direction, then back toward i. With nowhere to go it stays
// at i.
func (l List) Key(name string, i int) (int, bool) {
	j, ok := l.key(name, i)
	if j < 0 {
		j = i
	}
	return j, ok
}

func (l List) key(name string, i int) (int, bool) {
	if l.Count == 0 {
		switch key.Name(name) {
		case key.NameDownArrow, key.NameUpArrow, key.NameHome, key.NameEnd, key.NamePageDown, key.NamePageUp:
			return -1, true
		}
		return i, false
	}
	page := l.Page
	if page <= 0 {
		page = 10
	}
	switch key.Name(name) {
	case key.NameDownArrow:
		if i < 0 {
			return l.First(), true
		}
		return l.Next(i, 1), true
	case key.NameUpArrow:
		if i < 0 {
			return l.Last(), true
		}
		return l.Next(i, -1), true
	case key.NameHome:
		return l.First(), true
	case key.NameEnd:
		return l.Last(), true
	case key.NamePageDown:
		return l.page(i, min(max(i+page, 0), l.Count-1), 1), true
	case key.NamePageUp:
		return l.page(i, max(i-page, 0), -1), true
	}
	return i, false
}

func (l List) page(from, target, dir int) int {
	for j := target; j >= 0 && j < l.Count; j += dir {
		if l.enabled(j) {
			return j
		}
	}
	for j := target - dir; j != from && j >= 0 && j < l.Count; j -= dir {
		if l.enabled(j) {
			return j
		}
	}
	if l.enabled(from) {
		return from
	}
	return -1
}
