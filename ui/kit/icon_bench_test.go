package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

var iconRenderSink el.Element

func BenchmarkIconRender(b *testing.B) {
	for _, fresh := range []bool{false, true} {
		name := "retained"
		if fresh {
			name = "fresh"
		}
		b.Run(name, func(b *testing.B) {
			icon := Icon(IconHome)
			icon.Render(nil)
			b.ReportAllocs()
			for b.Loop() {
				if fresh {
					icon = Icon(IconHome)
				}
				iconRenderSink = icon.Render(nil)
			}
		})
	}
}

func BenchmarkRetainedIconFrame(b *testing.B) {
	icon := Icon(IconHome)
	h := render(viewFunc(func(cx *el.Context) el.Element { return icon.Render(cx) }))
	for range 3 {
		h.Frame()
	}
	b.ReportAllocs()
	for b.Loop() {
		h.Frame()
	}
}
