package kit

import (
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"strings"
	"testing"
)

func TestInputSizesKeepEditingAndGrowFrames(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Input("").Placeholder("Sized")
		v.SetValue("gyp 中文")
		var cx *el.Context
		h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }), 260, scale)
		clickClass(t, h, "Editor", "Sized")
		previous := float32(0)
		for _, size := range []InputSize{InputSizeXSmall, InputSizeSmall, InputSizeMedium, InputSizeLarge} {
			v.Size(size)
			h.Frame()
			_, height := cx.LastSize(autoID("input", v))
			if height <= previous {
				t.Fatal("size did not increase frame height", size, height, previous)
			}
			if !cx.Focused(v.FocusID()) || v.Value() != "gyp 中文" {
				t.Fatal("size changed value or focus")
			}
			previous = height
		}
		v.Size(InputSize(255))
		h.Frame()
		_, height := cx.LastSize(autoID("input", v))
		if height != previous {
			t.Fatal("invalid size changed frame")
		}
	}
}

func TestTextAreaSizeAutoGrowAndRows(t *testing.T) {
	for _, size := range []InputSize{InputSizeXSmall, InputSizeSmall, InputSizeMedium, InputSizeLarge} {
		v := TextArea("").Placeholder("Draft").Size(size).AutoGrow(1, 3)
		h := renderView(v, 200, 1)
		initial := bounds(h, "Draft").Dy()
		v.SetValue(strings.Repeat("gyp 中文\n", 10))
		h.Frame()
		grown := bounds(h, "Draft").Dy()
		if grown <= initial || grown > initial*3+4 {
			t.Fatal("autogrow size limit", size, initial, grown)
		}
		v.SetValue("")
		h.Frame()
		if bounds(h, "Draft").Dy() != initial {
			t.Fatal("autogrow failed to shrink")
		}
		v.Rows(5)
		h.Frame()
		if bounds(h, "Draft").Dy() < grown {
			t.Fatal("fixed rows lost minimum height")
		}
	}
}

func TestTextAreaFixedRowsFollowFontScale(t *testing.T) {
	for _, size := range []InputSize{InputSizeXSmall, InputSizeSmall, InputSizeMedium, InputSizeLarge} {
		v := TextArea("").Placeholder("Draft").Rows(5).Size(size)
		root := el.Embed(v)
		fontScale := float32(1)
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: fontScale}
			gtx.Constraints.Max = image.Pt(500, 600)
			root.Layout(gtx)
		})
		before := bounds(h, "Draft").Dy()
		fontScale = 2
		h.Frame()
		after := bounds(h, "Draft").Dy()
		if after < before*2-1 {
			t.Fatal("fixed rows ignored font scaling", before, after)
		}
	}
}
