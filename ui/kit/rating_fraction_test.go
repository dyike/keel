package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"math"
	"testing"
)

func TestRatingFractionalReadOnlyAndNormalization(t *testing.T) {
	calls := 0
	v := Rating("Average", 5).ReadOnly().OnChange(func(int) { calls++ })
	v.SetScore(3.7)
	h := page(v)
	n, ok := semanticNode(h, "slider:3.7/5")
	if !ok {
		t.Fatal("fractional semantics missing")
	}
	x, y := center(n.Desc.Bounds)
	h.Click(x, y)
	h.Key(key.NameRightArrow, 0)
	if v.Score() != 3.7 || calls != 0 {
		t.Fatal("read-only score changed")
	}
	v.SetScore(4.5)
	h.Frame()
	if _, ok := semanticNode(h, "slider:4.5/5"); !ok {
		t.Fatal("half-star value missing")
	}
	for _, tc := range []struct{ in, want float64 }{{math.NaN(), 0}, {math.Inf(1), 5}, {math.Inf(-1), 0}, {-1, 0}, {8, 5}} {
		v.SetScore(tc.in)
		h.Frame()
		if v.Score() != tc.want || calls != 0 {
			t.Fatalf("normalize %v", tc.in)
		}
	}
	v.SetValue(2)
	if v.Score() != 2 || v.Value() != 2 {
		t.Fatal("legacy setter")
	}
}

func TestRatingFractionalValueStillEditsWholeStars(t *testing.T) {
	var selected int
	v := Rating("Score", 5).OnChange(func(n int) { selected = n })
	v.SetScore(2.5)
	h := page(v)
	n, _ := semanticNode(h, "slider:2.5/5")
	h.Click(float32(n.Desc.Bounds.Min.X+24*3+11), float32(n.Desc.Bounds.Min.Y+11))
	if v.Score() != 4 || selected != 4 {
		t.Fatalf("integer edit %v %d", v.Score(), selected)
	}
}
