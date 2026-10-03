package kit

import (
	"image"
	"math"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestAttachmentVerticalPreviewAndActions(t *testing.T) {
	for _, scale := range []int{1, 2} {
		removed, opened := 0, 0
		preview := Image(image.NewNRGBA(image.Rect(0, 0, 80, 40)), "Picture").Size(80, 40)
		a := Attachment("photo.png", 1024).Media(preview).Vertical(true).OnOpen(func() { opened++ }).OnRemove(func() { removed++ }).
			PartStyle(AttachmentPartMedia, func(e *el.DivEl) { e.Name("Media box") })
		h := renderView(a, 180, scale)
		media := bounds(h, "Media box")
		if media.Dx() != media.Dy() || media.Dx() > 180*scale || media.Dx() < 140*scale {
			t.Fatal("square viewport", media)
		}
		picture := bounds(h, "Picture")
		if picture != media {
			t.Fatal("image not filling preview", picture, media)
		}
		remove := bounds(h, locale.Current().Name(locale.Current().Remove, "photo.png"))
		if remove.Max.X != media.Max.X || remove.Min.Y != media.Min.Y {
			t.Fatal("actions not top trailing", remove, media)
		}
		click(t, h, locale.Current().Name(locale.Current().Remove, "photo.png"))
		if removed != 1 || opened != 0 {
			t.Fatal("remove leaked to open")
		}
		a.MediaAspectRatio(2)
		h.Frame()
		media = bounds(h, "Media box")
		if absInt(media.Dx()-media.Dy()*2) > scale {
			t.Fatal("custom ratio", media)
		}
		h.Key(key.NameSpace, 0)
		if removed != 2 {
			t.Fatal("ratio lost action focus")
		}
		a.MediaAspectRatio(float32(math.NaN())).MediaAspectRatio(-1)
		h.Frame()
		if bounds(h, "Media box") != media {
			t.Fatal("invalid ratio changed layout")
		}
		a.MediaAspectRatio(0)
		h.Frame()
		if bounds(h, "Picture").Dx() != 80*scale || bounds(h, "Picture").Dy() != 40*scale {
			t.Fatal("natural sizing not restored")
		}
		if preview.width != 80 || preview.height != 40 || preview.fit != ImageContain {
			t.Fatal("source image mutated")
		}
		a.Vertical(false)
		h.Frame()
		h.Key(key.NameSpace, 0)
		if removed != 3 {
			t.Fatal("orientation lost action focus", removed)
		}
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
