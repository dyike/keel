package markdown

import (
	"context"
	"fmt"
	"image"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestMarkdownImagesLoadReflowAndCache(t *testing.T) {
	loaded := make(chan struct{})
	var requests atomic.Int32
	d := New("![first](fixture)\n\n![second](fixture)\n\n![failure](missing)\n\nafter").ImageLoader(func(ctx context.Context, src string) (image.Image, error) {
		requests.Add(1)
		<-loaded
		if src == "missing" {
			return nil, fmt.Errorf("missing fixture")
		}
		return image.NewNRGBA(image.Rect(0, 0, 600, 60)), nil
	})
	h := uitest.New(el.Root(docView{d}))
	first := d.chunks[0].blocks[0].view.(*richBlock)
	before := first.rt.size.Y
	close(loaded)
	for until := time.Now().Add(2 * time.Second); time.Now().Before(until); {
		h.Frame()
		if d.images["fixture"].Ready() && d.images["missing"].Ready() {
			break
		}
		time.Sleep(time.Millisecond)
	}
	h.Frame()
	if requests.Load() != 2 {
		t.Fatalf("identical sources were loaded %d times", requests.Load())
	}
	if got := first.rt.size.Y; got >= before || got < 35 || got > 45 {
		t.Fatalf("image completion did not reflow cached block: %d -> %d", before, got)
	}
	failure := d.chunks[2].blocks[0].view.(*richBlock)
	if failure.runs[0].image.Asset.Error() == nil {
		t.Fatal("missing image has no failure state")
	}
	found := false
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if strings.Contains(n.Desc.Description, "image:loaded") && n.Desc.Label == "first" {
			found = true
		}
	})
	if !found {
		t.Fatal("loaded image has no accessible alt text")
	}
	textPoint(t, h, d, "after", 0)
}
func TestLinkedImageAndEmptyAlt(t *testing.T) {
	clicked := ""
	d := New("[![click me](fixture)](https://example.com)\n\n![](fixture)").ImageLoader(func(context.Context, string) (image.Image, error) {
		return image.NewNRGBA(image.Rect(0, 0, 120, 40)), nil
	}).OnLink(func(s string) { clicked = s })
	h := uitest.New(el.Root(docView{d}))
	for until := time.Now().Add(time.Second); !d.images["fixture"].Ready() && time.Now().Before(until); {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	h.Frame()
	var bounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Class.String() == "Button" && n.Desc.Label == "[图片 click me]" {
			bounds = n.Desc.Bounds
		}
	})
	if bounds.Empty() {
		t.Fatal("linked image is not clickable")
	}
	h.Click(float32(bounds.Min.X+5), float32(bounds.Min.Y+5))
	if clicked != "https://example.com" {
		t.Fatalf("image opened %q", clicked)
	}
	r := d.chunks[1].blocks[0].view.(*richBlock)
	if r.runs[0].image == nil || r.runs[0].image.Asset.Size().X != 120 {
		t.Fatal("empty alt suppressed image")
	}
}

func TestImageFormattedAlt(t *testing.T) {
	d := New("![a **bold** and `code`](fixture)")
	s := documentBlocks(d)[0].spans[0]
	if s.imageAlt != "a bold and code" {
		t.Fatalf("formatted alt lost text: %q", s.imageAlt)
	}
}
