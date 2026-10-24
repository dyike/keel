package el

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
)

// benchmarkScrollFrame lays out a 500-row list that shows about 20 rows,
// with or without describing the hidden rows for accessibility.
func benchmarkScrollFrame(b *testing.B, accessibility bool) {
	rows := make([]string, 500)
	for i := range rows {
		rows[i] = fmt.Sprintf("Order %d", i)
	}
	root := Root(ViewFunc(func(*Context) Element {
		list := Div().Col().HFull().ScrollY()
		for _, r := range rows {
			list.Child(Div().Row().Gap(8).P(4).OnClick(func() {}).Child(Text(r), Text("Pending")))
		}
		return list
	}))
	loop.SetAccessibility(accessibility)
	defer loop.SetAccessibility(false)
	h := uitest.NewFunc(func(gtx core.C) { root.Layout(gtx) })
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		h.Frame()
	}
}

func BenchmarkScrollFrame(b *testing.B)              { benchmarkScrollFrame(b, false) }
func BenchmarkScrollFrameAccessibility(b *testing.B) { benchmarkScrollFrame(b, true) }
