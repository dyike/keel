package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestProgressMotionSettlesExactly(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	p := Progress("Progress")
	p.SetValue(.5)
	h := c.harness(func(cx *el.Context) el.Element { return p.Render(cx) })
	p.SetValue(.1)
	h.Frame()
	c.advance(h, ProgressDuration)
	if p.motion.value != p.motion.to || p.motion.value != p.Value() {
		t.Fatal("fractional transition never settled", p.motion)
	}
}
