package el

import (
	"image/color"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

// A hover background fades in over bgEase instead of switching, and switches
// at once with reduced motion.
func TestHoverBackgroundEases(t *testing.T) {
	from, to := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 255}
	for _, reduced := range []bool{false, true} {
		theme.SetReducedMotion(reduced)
		var shown color.NRGBA
		now := time.Unix(100, 0)
		b := Div().Size(Dp(40)).Bg(from).OnClick(func() {}).Hover(func(s *Style) { s.Bg(to) })
		root := Root(ViewFunc(func(*Context) Element { return Div().Child(b) }))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Now = now
			root.Layout(gtx)
			if st := root.store.states[b.n.key]; st != nil {
				shown = st.bgShown
			}
		})
		h.Frame()
		h.Move(20, 20)
		now = now.Add(bgEase / 2)
		h.Frame()
		mid := shown
		now = now.Add(bgEase)
		h.Frame()
		switch {
		case reduced && mid != to:
			t.Fatalf("reduced motion should switch at once: %v", mid)
		case !reduced && (mid == from || mid == to || mid.R == 0 || mid.B == 0):
			t.Fatalf("halfway through the fade: %v", mid)
		case shown != to:
			t.Fatalf("fade should end at the hover color: %v", shown)
		}
	}
	theme.SetReducedMotion(false)
}
