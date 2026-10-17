// SPDX-License-Identifier: Unlicense OR MIT

package dirty

import (
	"image"
	"testing"
)

func TestSince(t *testing.T) {
	var tr Tracker
	if r, v, ok := tr.Since(0); !ok || v != 0 || !r.Empty() {
		t.Fatalf("fresh tracker: %v %d %v", r, v, ok)
	}
	tr.Invalidate(image.Rectangle{}) // empty changes are not versions
	tr.Invalidate(image.Rect(0, 0, 1, 1))
	tr.Invalidate(image.Rect(5, 5, 6, 6))
	if r, v, ok := tr.Since(0); !ok || v != 2 || r != image.Rect(0, 0, 6, 6) {
		t.Fatalf("Since(0) = %v %d %v", r, v, ok)
	}
	if r, _, ok := tr.Since(1); !ok || r != image.Rect(5, 5, 6, 6) {
		t.Fatalf("Since(1) = %v %v", r, ok)
	}
	// Exactly history versions behind is still answerable; one more is not.
	for i := 0; i < history-2; i++ {
		tr.Invalidate(image.Rect(0, 0, 1, 1))
	}
	if _, v, ok := tr.Since(0); !ok || v != history {
		t.Fatalf("Since(0) at the log's limit: v=%d ok=%v", v, ok)
	}
	tr.Invalidate(image.Rect(0, 0, 1, 1))
	if _, _, ok := tr.Since(0); ok {
		t.Fatal("Since(0) past the log's limit reported ok")
	}
	if _, _, ok := tr.Since(tr.Version() + 1); ok {
		t.Fatal("a version from the future reported ok")
	}
}
