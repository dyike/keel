package el

import (
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

// A content-sized box widened by MinW stretches its children to the final width.
func TestMinWidthStretchesChildren(t *testing.T) {
	h := uitest.New(Embed(ViewFunc(func(cx *Context) Element {
		return Div().MinW(Dp(200)).Items(Stretch).Child(Div().Name("row").H(Dp(20)).Child(Text("短")))
	})))
	if b := nodeBounds(h, "row"); b.Dx() != 200 {
		t.Fatalf("row width %d, want 200", b.Dx())
	}
}
