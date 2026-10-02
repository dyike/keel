package el

import (
	"hash/fnv"
	"image"
	"strconv"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/widget"

	"github.com/dyike/keel/ui/internal/editorstyle"
)

// stateKey identifies an element across frames: a hash of its path from the
// root, where each step is the element's ID or else its index among siblings.
type stateKey uint64

func childKey(parent stateKey, id string, index int) stateKey {
	h := fnv.New64a()
	var b [8]byte
	for i := range b {
		b[i] = byte(parent >> (8 * i))
	}
	h.Write(b[:])
	if id != "" {
		h.Write([]byte("#" + id))
	} else {
		h.Write([]byte(strconv.Itoa(index)))
	}
	return stateKey(h.Sum64())
}

// elemState is what an element keeps between frames.
type elemState struct {
	hoverTag   int
	hovered    bool
	blocked    bool
	id         string
	focusable  bool
	disabled   bool
	onKey      func(KeyEvent) bool
	keyParent  *elemState
	keyFrame   uint64
	pressedKey key.Name

	frame uint64 // last frame the element was painted in

	click         gesture.Click
	onClick       func()
	onDoubleClick func()
	drag          gesture.Drag
	onDrag        func(DragEvent)
	size          image.Point // painted size in px, for drag events
	clickable     bool        // registered a click area last frame
	fresh         bool        // created this frame: dispatch has not seen it yet

	scrollbarX, scrollbarY   scrollbarState
	scrollableX, scrollableY bool

	scrollHorizontal                                                 gesture.Scroll
	scrollX, scrollPendingX, scrollMaxX, scrollViewX, scrollContentX int
	scrolledX                                                        bool
	scroll                                                           gesture.Scroll
	scrollPending                                                    int
	scrollY                                                          int
	scrollMax                                                        int // maxScroll at the last frame, for StickToBottom
	scrollView                                                       int // viewport and content height at the last frame, for ScrollState
	scrollContent                                                    int
	scrolled                                                         bool // painted before: a first frame starts at the bottom
	version                                                          int  // ScrollToEndOn's value last frame
	keepVersion                                                      int  // KeepBottomOn's value last frame

	editor   widget.Editor
	caret    editorstyle.Caret
	edInit   bool
	lastText string // what Bind last synced, to spot program changes
}

// store holds element state for one root. Entries not painted in a frame are
// dropped after it, so state never outlives its element by more than a frame.
type store struct {
	states map[stateKey]*elemState
	frame  uint64
}

func newStore() *store { return &store{states: map[stateKey]*elemState{}} }

func (s *store) get(k stateKey) *elemState {
	st := s.states[k]
	if st == nil {
		st = &elemState{fresh: true}
		s.states[k] = st
	}
	st.frame = s.frame
	return st
}

func (s *store) sweep() {
	for k, st := range s.states {
		if st.frame != s.frame {
			delete(s.states, k)
		}
	}
}

// assignKeys gives every node its state key, and keeps the state of every
// element in the tree alive, painted or not (it may be scrolled out of view).
func (s *store) assignKeys(n *Node, key stateKey) {
	n.key = key
	if st := s.states[key]; st != nil {
		st.frame = s.frame
	}
	for i, c := range n.children {
		cn := c.node()
		s.assignKeys(cn, childKey(key, cn.id, i))
	}
}
