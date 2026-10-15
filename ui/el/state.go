package el

import (
	"hash/fnv"
	"image"
	"image/color"
	"strconv"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gesture"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/widget"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/editorstyle"
)

// stateKey identifies an element across frames: a hash of its path from the
// root, where each step is the element's ID or else its index among siblings.
type stateKey uint64

func childKey(parent stateKey, id string, index int) stateKey {
	h := fnv.New64a()
	var b [20]byte // also holds any signed 64-bit decimal sibling index
	for i := range b[:8] {
		b[i] = byte(parent >> (8 * i))
	}
	h.Write(b[:8])
	if id != "" {
		h.Write([]byte("#"))
		h.Write([]byte(id))
	} else {
		h.Write(strconv.AppendInt(b[:0], int64(index), 10))
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
	keyFilters []event.Filter
	filterMode uint8
	// Pointer focus keeps keyboard routing without drawing a focus ring.
	pointerFocus bool
	// inputRing: the editor took focus from the keyboard or a program, so its
	// field shows a ring. A click shows only the caret, as native fields do.
	// inputSource says el already knows the source of the pending focus;
	// otherwise (Gio's own Tab, a press on the text) a recent press decides.
	inputRing, inputSource bool

	enabledFrame uint64 // last live, visible, enabled paint
	frame        uint64 // last frame the element was painted in

	click           gesture.Click
	onClick         func()
	onContextMenu   func()
	contextButton   pointer.Buttons
	contextTag      struct{}
	touchHold       touchHold // a finger held down for a context menu
	holdFired       bool      // the hold opened a menu: its release is no click
	onDoubleClick   func()
	drag            gesture.Drag
	conditionalDrag conditionalDrag
	onDrag          func(DragEvent)
	dragAccept      func(float32, float32) bool
	onScroll        *scrollHandler
	scrollTag       struct{}
	size            image.Point // painted size in px, for drag events
	clickable       bool        // registered a click area last frame
	// FocusOnPress, and an input's padding: a press here focuses pressFocus,
	// or this input's own editor.
	pressTag    struct{}
	pressFocus  string
	pressEditor bool
	pressable   bool

	// The background shown last frame and the change it is easing through,
	// for elements with hover or pressed styles.
	bgShown, bgFrom, bgTo color.NRGBA
	bgStart               time.Time
	bgInit                bool
	fresh                 bool // created this frame: dispatch has not seen it yet

	scrollbarX, scrollbarY   scrollbarState
	scrollHover              gesture.Hover
	scrollVisibleUntil       time.Time
	scrollShownAt            time.Time // when hidden bars began to show
	scrollWantedAt           time.Time // the last frame the bars were wanted
	scrollAlpha              float32
	scrollableX, scrollableY bool

	scrollHorizontal                                                 gesture.Scroll
	scrollX, scrollPendingX, scrollMaxX, scrollViewX, scrollContentX int
	scrolledX                                                        bool
	scroll                                                           gesture.Scroll
	scrollPending                                                    int
	scrollY                                                          int
	scrollMax                                                        int             // maxScroll at the last frame, for StickToBottom
	scrollBounds                                                     image.Rectangle // enclosing-root coordinates, before child clipping
	scrollView                                                       int             // viewport and content height at the last frame, for ScrollState
	scrollContent                                                    int
	scrolled                                                         bool // painted before: a first frame starts at the bottom
	version                                                          int  // ScrollToEndOn's value last frame
	keepVersion                                                      int  // KeepBottomOn's value last frame

	inputObjects          inputObjectLayout
	inputTokenHits        []*inputTokenHit
	inputDocument         *InputDocument
	inputDocumentRevision uint64
	inputDocumentFilters  []event.Filter
	inputComposition      key.Range
	editor                *widget.Editor
	caret                 editorstyle.Caret
	edInit                bool
	inputFocused          bool
	inputPaste            *inputPasteRequest
	inputActions          []InputAction
	inputSelection        *[2]int
	inputUndo, inputRedo  []InputEdit
	lastText              string // what Bind last synced, to spot program changes
}

// inputEditor keeps the large native editor out of buttons, scroll containers
// and virtual rows. Its address remains stable for Gio's focus and event tags.
func (st *elemState) inputEditor() *widget.Editor {
	if st.editor == nil {
		st.editor = new(widget.Editor)
	}
	return st.editor
}

// store holds element state for one root. Entries not painted in a frame are
// dropped after it, so state never outlives its element by more than a frame.
type store struct {
	states map[stateKey]*elemState
	frame  uint64
	// pressFrame is the frame of the last pointer press anywhere in the root:
	// an editor that gains focus right after one was clicked into.
	pressFrame uint64
	pressed    bool
}

// recentPress reports whether a pointer press arrived in the last two
// frames: Gio focuses an editor on the frame after the press reaches it.
func (s *store) recentPress() bool { return s.pressed && s.frame-s.pressFrame <= 2 }

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

// touchHold tracks a touch that may become a long press.
type touchHold struct {
	active bool
	id     pointer.ID
	at     f32.Point
	due    time.Time
}

// touchHoldDelay is how long a finger rests before a long press opens the
// context menu; touchHoldSlop is how far it may drift, in dp.
const (
	touchHoldDelay = 500 * time.Millisecond
	touchHoldSlop  = 8
)

// trackTouchHold starts a hold on a single-finger touch press and drops it
// when the finger lifts, drifts or another handler takes the pointer.
func (st *elemState) trackTouchHold(gtx core.C, e pointer.Event) {
	h := &st.touchHold
	switch e.Kind {
	case pointer.Press:
		if e.Source == pointer.Touch && !h.active {
			st.holdFired = false
			*h = touchHold{active: true, id: e.PointerID, at: e.Position, due: gtx.Now.Add(touchHoldDelay)}
		} else if e.Source == pointer.Touch {
			h.active = false // a second finger: a gesture, not a hold
		}
	case pointer.Drag:
		if h.active && e.PointerID == h.id {
			d := e.Position.Sub(h.at)
			if slop := float32(gtx.Dp(touchHoldSlop)); d.X*d.X+d.Y*d.Y > slop*slop {
				h.active = false
			}
		}
	case pointer.Release, pointer.Cancel:
		if e.PointerID == h.id || e.Kind == pointer.Cancel {
			h.active = false
		}
	}
}

// focusFrom records where focus came from: a pointer focuses quietly, the
// keyboard or a program shows the ring, on a control or its input's field.
func (s *elemState) focusFrom(pointer bool) {
	s.pointerFocus, s.inputRing, s.inputSource = pointer, !pointer, true
}
