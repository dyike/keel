package el

import (
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// signalSource is a reactive value (Signal or Computed) whose reads are tracked
// inside Computed evaluations and Context.Memo builders.
type signalSource interface {
	signalVersion() uint64
}

type signalDep struct {
	src     signalSource
	version uint64
}

type depTracker struct {
	deps []signalDep
}

func (t *depTracker) record(src signalSource, version uint64) {
	for i := range t.deps {
		if t.deps[i].src == src {
			t.deps[i].version = version
			return
		}
	}
	t.deps = append(t.deps, signalDep{src: src, version: version})
}

// activeTracker is the dependency collector currently recording reads under the
// UI frame lock. Nested Computed or Memo evaluations push and restore it.
var activeTracker *depTracker

// Signal holds a reactive UI value. Reading it with Get inside a Computed or
// Context.Memo automatically subscribes that computation or subtree to future
// changes. Writing an unchanged value with Set is a no-op and triggers no redraw.
//
// Like other UI state in Keel, read and write Signals from UI callbacks or
// inside core.Update.
type Signal[T any] struct {
	value   T
	version uint64
	equal   func(a, b T) bool
}

// NewSignal creates a Signal for a comparable type. Set compares with == and
// skips invalidation when the value is unchanged.
func NewSignal[T comparable](initial T) *Signal[T] {
	return &Signal[T]{
		value: initial,
		equal: func(a, b T) bool { return a == b },
	}
}

// NewSignalFunc creates a Signal with a custom equality check (useful for
// slices or structs with non-comparable fields). If equal is nil, every Set
// marks the signal changed.
func NewSignalFunc[T any](initial T, equal func(a, b T) bool) *Signal[T] {
	return &Signal[T]{value: initial, equal: equal}
}

// Get returns the current value and records this Signal as a dependency of any
// enclosing Computed or Context.Memo.
func (s *Signal[T]) Get() T {
	if activeTracker != nil {
		activeTracker.record(s, s.version)
	}
	return s.value
}

// Peek returns the current value without recording a dependency.
func (s *Signal[T]) Peek() T { return s.value }

// Set updates the value. If it equals the current value, Set returns false and
// does not bump the version or request a redraw. Otherwise it bumps the version,
// requests a redraw, and returns true.
func (s *Signal[T]) Set(v T) bool {
	if s.equal != nil && s.equal(s.value, v) {
		return false
	}
	s.value = v
	s.version++
	loop.InvalidateAll()
	return true
}

// Update replaces the value with fn(current) and reports whether it changed.
func (s *Signal[T]) Update(fn func(T) T) bool { return s.Set(fn(s.value)) }

// Version returns the number of times the signal has changed.
func (s *Signal[T]) Version() uint64 { return s.version }

func (s *Signal[T]) signalVersion() uint64 { return s.version }

// Computed is a lazily cached value derived from one or more Signals (or other
// Computed values). It re-evaluates only when an upstream dependency changes.
type Computed[T any] struct {
	fn        func() T
	value     T
	deps      []signalDep
	version   uint64
	evaluated bool
}

// NewComputed creates a lazy derived value computed by fn.
func NewComputed[T any](fn func() T) *Computed[T] {
	return &Computed[T]{fn: fn}
}

func (c *Computed[T]) eval() {
	if c.evaluated && !depsChanged(c.deps) {
		return
	}
	var tracker depTracker
	prior := activeTracker
	activeTracker = &tracker
	var v T
	func() {
		defer func() { activeTracker = prior }()
		v = c.fn()
	}()
	c.value = v
	c.deps = tracker.deps
	c.evaluated = true
	c.version++
}

// Get returns the derived value, recomputing only if a dependency changed, and
// records this Computed as a dependency of any enclosing Computed or Context.Memo.
func (c *Computed[T]) Get() T {
	c.eval()
	if activeTracker != nil {
		activeTracker.record(c, c.version)
	}
	return c.value
}

// Peek returns the derived value without recording a dependency.
func (c *Computed[T]) Peek() T {
	c.eval()
	return c.value
}

// Version returns the number of times this Computed has evaluated.
func (c *Computed[T]) Version() uint64 {
	c.eval()
	return c.version
}

func (c *Computed[T]) signalVersion() uint64 {
	c.eval()
	return c.version
}

func depsChanged(deps []signalDep) bool {
	for i := range deps {
		if deps[i].src.signalVersion() != deps[i].version {
			return true
		}
	}
	return false
}

// Memo returns a cached element subtree identified by a stable comparable id,
// automatically tracking every Signal.Get and Computed.Get read inside build
// (as well as theme and locale revisions).
//
// When none of the tracked dependencies have changed since the previous frame,
// build is skipped and the retained subtree (along with its cached layout when
// constraints match) is reused. Nested Memo calls update in place when only an
// inner Signal changes, without rebuilding the enclosing Memo.
//
// Like Context.Cache, build must construct its own persistent subtree and must
// not return frame-arena elements from the surrounding Render.
func (cx *Context) Memo(id any, build func() Element) Element {
	r := cx.root
	mc := &r.memos
	if mc.entries == nil {
		mc.entries = map[any]*memoEntry{}
	}
	if rev := [2]uint64{theme.Revision(), locale.Revision()}; mc.revision != rev {
		clear(mc.entries)
		mc.revision = rev
	}
	frame := r.store.frame
	var parent *memoEntry
	if n := len(r.memoStack); n > 0 {
		parent = r.memoStack[n-1]
	}

	if m, ok := mc.entries[id]; ok {
		m.build = build
		if !depsChanged(m.deps) {
			changed, mustRebuild := m.refreshChildren(cx, frame)
			if !mustRebuild {
				if changed && m.el != nil {
					clearLayoutMemo(m.el.node())
				}
				m.markFrame(frame)
				if parent != nil {
					parent.children = append(parent.children, m)
				}
				return m.el
			}
		}
		el := m.rebuild(cx, build, frame)
		if parent != nil {
			parent.children = append(parent.children, m)
		}
		return el
	}

	m := &memoEntry{id: id, build: build}
	mc.entries[id] = m
	el := m.rebuild(cx, build, frame)
	if parent != nil {
		parent.children = append(parent.children, m)
	}
	return el
}

type memoCache struct {
	entries  map[any]*memoEntry
	revision [2]uint64
}

type memoEntry struct {
	id       any
	el       Element
	build    func() Element
	deps     []signalDep
	children []*memoEntry
	host     *Node
	hostIdx  int
	frame    uint64
}

func (m *memoEntry) markFrame(frame uint64) {
	m.frame = frame
	for _, child := range m.children {
		child.markFrame(frame)
	}
}

func (m *memoEntry) rebuild(cx *Context, build func() Element, frame uint64) Element {
	r := cx.root
	m.children = m.children[:0]
	m.frame = frame

	priorArena := buildingArena
	buildingArena = nil
	var tracker depTracker
	priorTracker := activeTracker
	activeTracker = &tracker
	r.memoStack = append(r.memoStack, m)

	var el Element
	func() {
		defer func() {
			r.memoStack = r.memoStack[:len(r.memoStack)-1]
			activeTracker = priorTracker
			buildingArena = priorArena
		}()
		el = build()
	}()

	m.deps = tracker.deps
	m.el = el
	if el == nil {
		return nil
	}
	checkPersistentSubtree(el.node(), m.children)
	el.node().cached = true
	return el
}

// refreshChildren checks nested Memo entries in place without re-running the
// enclosing Memo's build. It reports whether any descendant subtree changed and
// whether the enclosing Memo must rebuild (e.g. a nil child became non-nil).
func (m *memoEntry) refreshChildren(cx *Context, frame uint64) (changed, mustRebuild bool) {
	for _, child := range m.children {
		if depsChanged(child.deps) {
			oldNode := nodeOf(child.el)
			if oldNode == nil {
				return false, true
			}
			newEl := child.rebuild(cx, child.build, frame)
			if child.host != nil && child.hostIdx >= 0 && child.hostIdx < len(child.host.children) && nodeOf(child.host.children[child.hostIdx]) == oldNode {
				if newEl != nil {
					child.host.children[child.hostIdx] = newEl
				} else {
					return false, true
				}
			} else if nodeOf(m.el) == oldNode {
				m.el = newEl
			}
			clearLayoutMemo(child.host)
			changed = true
		} else {
			cChanged, cRebuild := child.refreshChildren(cx, frame)
			if cRebuild {
				return false, true
			}
			if cChanged {
				clearLayoutMemo(nodeOf(child.el))
				clearLayoutMemo(child.host)
				changed = true
			}
		}
	}
	return changed, false
}

func nodeOf(el Element) *Node {
	if el == nil {
		return nil
	}
	return el.node()
}

func clearLayoutMemo(n *Node) {
	if n != nil && n.memo != nil {
		n.memo.ok = false
	}
}

func checkPersistentSubtree(root *Node, childMemos []*memoEntry) {
	var byNode map[*Node]*memoEntry
	if len(childMemos) > 0 {
		byNode = make(map[*Node]*memoEntry, len(childMemos))
		for _, cm := range childMemos {
			cm.host = nil
			cm.hostIdx = -1
			if cm.el != nil {
				byNode[cm.el.node()] = cm
			}
		}
	}
	var walk func(*Node)
	walk = func(n *Node) {
		if n.arena != nil {
			panic("el: Context.Memo must build its own persistent subtree, not return a frame element")
		}
		for i, child := range n.children {
			cn := child.node()
			if cm, ok := byNode[cn]; ok {
				cm.host = n
				cm.hostIdx = i
				continue // already validated when the child Memo built
			}
			walk(cn)
		}
	}
	walk(root)
}

func (c *memoCache) sweep(frame uint64) {
	for k, m := range c.entries {
		if m.frame != frame {
			delete(c.entries, k)
		}
	}
}
