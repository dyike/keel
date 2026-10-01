package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMarkerShapesRespectSize(t *testing.T) {
	for _, shape := range []MarkerShape{MarkerDot, MarkerSquare, MarkerDiamond} {
		for _, scale := range []int{1, 2} {
			h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Name("host").Child(Marker(shape).Size(12).Render(cx)) }), 40, scale)
			if r := bounds(h, "host"); r.Dx() != 12*scale || r.Dy() != 12*scale {
				t.Fatal(r)
			}
		}
	}
}
