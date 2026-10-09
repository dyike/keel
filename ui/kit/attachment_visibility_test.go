package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"image"
	"testing"
)

func TestAttachmentOptionalPartsAndTile(t *testing.T) {
	for _, scale := range []int{1, 2} {
		opened, removed, used := 0, 0, 0
		preview := Image(image.NewNRGBA(image.Rect(0, 0, 100, 60)), "Preview")
		action := Button("Custom action", func() { used++ }).Size(24)
		a := Attachment("tile", 1024).Vertical(true).Media(preview).Actions(action).OnOpen(func() { opened++ }).OnRemove(func() { removed++ })
		h := renderView(a, 180, scale)
		normal := bounds(h, "tile")
		a.ShowContent(false)
		h.Frame()
		tile := bounds(h, "tile")
		if tile.Dx() != tile.Dy() || tile.Dx() > 180*scale || tile.Dy() >= normal.Dy() {
			t.Fatal("tile size", normal, tile)
		}
		media := bounds(h, "Preview")
		if media.Min.X-tile.Min.X != scale || media.Min.Y-tile.Min.Y != scale || tile.Max.X-media.Max.X != scale || tile.Max.Y-media.Max.Y != scale {
			t.Fatal("tile not edge-to-edge inside border", tile, media)
		}
		if shown(h, "1.0 KB") {
			t.Fatal("hidden description visible")
		}
		a.ShowActions(false)
		h.Frame()
		if shown(h, "Custom action") || shown(h, locale.Current().Name(locale.Current().Remove, "tile")) {
			t.Fatal("hidden action visible")
		}
		click(t, h, "Preview")
		h.Key(key.NameSpace, 0)
		if opened != 2 {
			t.Fatal("tile open and keyboard", opened)
		}
		a.ShowActions(true)
		h.Frame()
		click(t, h, "Custom action")
		if used != 1 || opened != 2 {
			t.Fatal("restored action leaked")
		}
		a.ShowMedia(false).ShowContent(true)
		h.Frame()
		if shown(h, "Preview") || !shown(h, "1.0 KB") {
			t.Fatal("metadata-only visibility")
		}
		a.ShowContent(false)
		h.Frame()
		click(t, h, "Custom action")
		if used != 2 || opened != 2 {
			t.Fatal("action-only behavior")
		}
		a.ShowMedia(true).ShowContent(true)
		h.Frame()
		if !shown(h, "Preview") || !shown(h, "1.0 KB") {
			t.Fatal("parts not restored")
		}
		a.SetProgress(.5)
		a.ShowContent(false).ShowMedia(false)
		h.Frame()
		if a.titleShimmer != nil && !a.titleShimmer.disabled {
			t.Fatal("hidden title animates")
		}
		if a.Status() != AttachmentStatusUploading {
			t.Fatal("visibility changed lifecycle")
		}
		if removed != 0 {
			t.Fatal("unexpected remove")
		}
	}
}
