package widget

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"math"
	"testing"
)

func TestProgressAnimationStopsAndPreservesValue(t *testing.T) {
	p := Progress("下载")
	p.SetValue(.4)
	p.SetIndeterminate(true)
	h := uitest.New(p)
	if _, ok := h.Router.WakeupTime(); !ok {
		t.Fatal("indeterminate progress has no frame wakeup")
	}
	p.SetIndeterminate(false)
	h.Frame()
	if _, ok := h.Router.WakeupTime(); ok {
		t.Fatal("determinate progress still animating")
	}
	if p.Value() != .4 {
		t.Fatal("indeterminate mode discarded progress")
	}
	p.SetValue(float32(math.NaN()))
	if p.Value() != 0 {
		t.Fatal("NaN progress not sanitized")
	}
}
