package el

import "image"

// Oversized roots remain correct, but their map capacity is not retained for
// subsequent small frames. LastSize reads the previous frame during Render;
// clear only when painting starts, before collecting this frame's anchors.
const maxRetainedAnchors = 4096

func (e *engine) beginAnchors() {
	if len(e.anchors) > maxRetainedAnchors {
		e.anchors = nil
	}
	if e.anchors == nil {
		e.anchors = make(map[string]image.Rectangle)
	} else {
		clear(e.anchors)
	}
}
