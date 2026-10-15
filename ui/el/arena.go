package el

// Element construction, like Render itself, is serialized by the UI frame
// lock. Scope the allocator to Render so ordinary elements constructed before
// opening a window and Context.Cache builders retain their own storage.
var buildingArena *nodeArena

const nodeChunkSize = 32

type nodeArena struct {
	chunks      []*[nodeChunkSize]Node
	used, dirty int
	owner       *arenaOwner
	divs        frameHandles[DivEl]
	texts       frameHandles[TextEl]
	inputs      frameHandles[InputEl]
	widgets     frameHandles[WidgetEl]
}

// Validation needs only a generation. This record has no back-pointer to the
// root or its chunks; node storage is owned separately by the arena.
type arenaOwner struct{ generation uint64 }

// Pointer builders need independent handles: reusing a handle would let an old
// pointer acquire the new generation. Batch their small allocations, but never
// reuse a batch in another build pass. Only the large nodes are recycled.
type frameHandles[T any] struct {
	batch      []T
	next, used int
	previous   int
}

func (h *frameHandles[T]) reset() {
	h.previous = h.used
	h.batch, h.next, h.used = nil, 0, 0
}

func (h *frameHandles[T]) alloc() *T {
	if h.next == len(h.batch) {
		// Bound retention by one 32-handle batch, but avoid reserving 32
		// wrappers for a root that only built a few of this kind last pass.
		h.batch = make([]T, min(nodeChunkSize, max(4, h.previous-h.used)))
		h.next = 0
	}
	v := &h.batch[h.next]
	h.next++
	h.used++
	return v
}

func (a *nodeArena) begin() {
	if a.owner == nil {
		a.owner = new(arenaOwner)
	}
	// Stop before wrapping: an old handle must never become valid again.
	if a.owner.generation == ^uint64(0) {
		panic("el: element arena generation exhausted")
	}
	a.owner.generation++
	a.dirty = max(a.dirty, a.used)
	a.used = 0
	a.divs.reset()
	a.texts.reset()
	a.inputs.reset()
	a.widgets.reset()
}

func (a *nodeArena) alloc() *Node {
	if a.owner.generation == 1 {
		// A widget/token renderer may create a root for a single layout. Do
		// not give those short-lived roots a full chunk they will never reuse.
		return &Node{arena: a.owner, forceW: -1, forceH: -1}
	}
	chunk, slot := a.used/nodeChunkSize, a.used%nodeChunkSize
	if chunk == len(a.chunks) {
		a.chunks = append(a.chunks, new([nodeChunkSize]Node))
	}
	n := &a.chunks[chunk][slot]
	a.used++
	children, positions, ranges := n.children[:0], n.flowPositions[:0], n.textRanges[:0]
	clear(n.children)
	*n = Node{arena: a.owner, children: children, flowPositions: positions, textRanges: ranges, forceW: -1, forceH: -1}
	return n
}

// finish drops references from nodes omitted by this pass. One empty chunk
// absorbs small changes in visible row count; larger vanished trees are freed.
func (a *nodeArena) finish() {
	keep := min(len(a.chunks), (a.used+nodeChunkSize-1)/nodeChunkSize+1)
	// Unlike a value handle, a saved pointer builder can retain a dropped
	// chunk. Clear those nodes too, so it cannot keep callbacks or views alive.
	for i, end := a.used, max(a.dirty, a.used); i < end; i++ {
		a.chunks[i/nodeChunkSize][i%nodeChunkSize] = Node{}
	}
	clear(a.chunks[keep:])
	a.chunks = a.chunks[:keep]
	a.dirty = a.used
}

func (a *nodeArena) pooled() bool { return a != nil && a.owner.generation > 1 }

func (s *Styled[T]) init(n *Node, self *T) {
	s.n, s.self = n, self
	if n.arena != nil {
		s.generation = n.arena.generation
	}
}

func (s *Styled[T]) node() *Node {
	n := s.n
	if s.generation != 0 && (n.arena == nil || n.arena.generation != s.generation) {
		panic("el: element used after its build pass; keep view state or use Context.Cache")
	}
	if n.arena != nil && buildingArena != nil && n.arena != buildingArena.owner {
		panic("el: element belongs to another root; build elements separately for each root")
	}
	return n
}

func checkElementOwner(parent, child *Node) {
	if parent.arena == nil && child.arena != nil {
		panic("el: a persistent element cannot retain a frame element; build the whole subtree inside Context.Cache")
	}
	if parent.arena != nil && child.arena != nil && child.arena != parent.arena {
		panic("el: element belongs to another root; build elements separately for each root")
	}
}
