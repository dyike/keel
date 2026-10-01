package widget

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestIconConstrainedCustomAndNil(t *testing.T) {
	for _, name := range []IconName{IconCheck, IconClose, IconPlus, IconSearch, IconCopy, IconChevronDown, IconChevronRight} {
		var d core.D
		uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(10, 10); d = Icon(name).Size(24).Layout(gtx) })
		if d.Size != image.Pt(10, 10) {
			t.Fatalf("icon exceeded constraints: %v", d.Size)
		}
	}
	var d core.D
	uitest.NewFunc(func(gtx core.C) { d = VectorIcon(nil).Layout(gtx) })
	if d.Size != (image.Point{}) {
		t.Fatal("nil icon reserves space")
	}
}
