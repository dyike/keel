package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestButtonLoadingSizesAndKeyboard(t *testing.T) {
	n := 0
	b := Button("保存", func() { n++ }).Icon(Icon(IconCheck))
	h := uitest.New(b)
	clickNamed(t, h, "保存")
	h.Key(key.NameSpace, 0)
	if n != 2 {
		t.Fatalf("mouse/keyboard callbacks: %d", n)
	}
	b.SetLoading(true)
	h.Frame()
	clickNamed(t, h, "保存")
	h.Key(key.NameReturn, 0)
	if n != 2 {
		t.Fatal("loading button accepted activation")
	}
	b.SetLoading(false)
	h.Frame()
	clickNamed(t, h, "保存")
	if n != 3 {
		t.Fatal("button did not recover from loading")
	}
	var sizes []image.Point
	for _, size := range []ComponentSize{Small, Medium, Large} {
		b.Size(size)
		uitest.NewFunc(func(gtx core.C) { sizes = append(sizes, b.Layout(gtx).Size) })
	}
	if sizes[0].Y >= sizes[1].Y || sizes[1].Y >= sizes[2].Y {
		t.Fatalf("sizes not distinct: %v", sizes)
	}
}
