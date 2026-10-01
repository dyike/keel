package el

// DragKind is the phase of a drag.
type DragKind uint8

const (
	DragStart DragKind = iota // the pointer went down on the element
	DragMove                  // it moved while down, even outside the element
	DragEnd                   // it went up, or the drag was cancelled
)

// DragEvent reports a pointer drag in dp, relative to the element's top left
// corner; W and H are the element's size, so X/W is a fraction of its width.
// X and Y may fall outside 0..W and 0..H while the pointer is outside.
type DragEvent struct {
	Kind       DragKind
	X, Y, W, H float32
}

// OnDrag reports presses, moves and releases on the element, e.g. for a
// slider thumb or a splitter. A press also starts an OnClick if both are set.
func (s *Styled[T]) OnDrag(fn func(DragEvent)) *T { s.n.onDrag = fn; return s.self }
