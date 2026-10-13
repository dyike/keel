package markdown

import (
	"bytes"
	"image"
	"testing"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestDisabledDocumentStillPaintsCachedText(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	doc := New("## 最终建议\n\n这是包含 **加粗** 的正文。\n\n> 引用中的消息也应保持可见。\n\n| 名称 | 内容 |\n|---|---|\n| 项目 | 完整正文 |\n\n```go\nfmt.Println(\"hello\")\n```")
	disabled := false
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Bg(theme.Bg).Disabled(disabled).Child(doc.Render(cx))
	}))
	var pixels *image.RGBA
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints = layout.Exact(image.Pt(400, 300))
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		pixels = image.NewRGBA(image.Rect(0, 0, 400, 300))
		if err := gpu.Screenshot(pixels); err != nil {
			t.Fatal(err)
		}
	})
	h.Frame()
	before := append([]byte(nil), pixels.Pix...)
	disabled = true
	for range 3 {
		h.Frame()
	}
	if !bytes.Equal(before, pixels.Pix) {
		t.Fatal("disabling document changed painted text")
	}
	disabled = false
	h.Frame()
	if !bytes.Equal(before, pixels.Pix) {
		t.Fatal("re-enabling document changed painted text")
	}
}
