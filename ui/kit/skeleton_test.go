package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestSkeletonConstraints(t *testing.T) {
	for _, scale := range []int{1, 2} {
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().Name("placeholder").Child(Skeleton().W(el.Full).H(el.Dp(16)).Render(cx))
		}), 100, scale)
		r := bounds(h, "placeholder")
		if r.Dx() != 100*scale || r.Dy() != 16*scale {
			t.Fatal(r)
		}
	}
}
