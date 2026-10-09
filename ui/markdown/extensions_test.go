package markdown

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
)

func TestDocumentCodeExtensions(t *testing.T) {
	d := New("> ```go\n> hello\n> ```")
	clicks := 0
	var snapshot CodeBlockContext
	d.CodeBlockActions(func(cx *el.Context, c CodeBlockContext) el.Element {
		snapshot = c
		return el.Div().Role("button").Name("inspect").OnClick(func() { clicks++ }).Child(el.Text("inspect"))
	})
	h := uitest.New(el.Root(docView{d}))
	id := snapshot.ID
	clickCodeButton(t, h, "inspect")
	if clicks != 1 || snapshot.Text != "hello" || snapshot.Language != "go" || id == "" {
		t.Fatalf("extension: clicks=%d context=%+v", clicks, snapshot)
	}
	clickCodeButton(t, h, "自动换行")
	if !snapshot.Wrapped || snapshot.ID != id {
		t.Fatal("context lost wrap or identity", snapshot)
	}
	d.CodeBlockRenderer(" GO ", func(cx *el.Context, c CodeBlockContext) el.Element { return el.Text("replacement") })
	h.Frame()
	if strings.Contains(d.RenderedText(), "hello") {
		t.Fatal("custom renderer retained original selection")
	}
	d.CodeBlockRenderer("go", nil)
	d.CodeBlockActions(nil)
	h.Frame()
	if d.RenderedText() != "hello" {
		t.Fatal("unregister retained stale cache", d.RenderedText())
	}
	clickCodeButton(t, h, "复制")
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != "hello" {
		t.Fatal("default copy lost", string(data))
	}
}

func TestDocumentRangesUnicodeAndRebase(t *testing.T) {
	d := New("前**中**尾\n\n后段")
	if d.SetRangeHighlights([]RangeHighlight{{Range: TextRange{0, 1}}}) {
		t.Fatal("accepted before paint")
	}
	h := uitest.New(el.Root(docView{d}))
	if d.RenderedText() != "前中尾\n\n后段" {
		t.Fatal(d.RenderedText())
	}
	red := color.NRGBA{R: 255, A: 255}
	items := []RangeHighlight{{TextRange{0, 3}, red}, {TextRange{11, 17}, red}}
	if !d.SetRangeHighlights(items) {
		t.Fatal("valid UTF8 ranges rejected")
	}
	if d.SetRangeHighlights([]RangeHighlight{{TextRange{1, 3}, red}}) || len(d.selection.highlights) != 2 {
		t.Fatal("invalid boundary wasn't atomic")
	}
	d.SetSource("前**变化**尾\n\n后段")
	if d.RevealRange(TextRange{0, 3}) {
		t.Fatal("accepted stale snapshot")
	}
	h.Frame()
	hs := d.selection.highlights
	if len(hs) != 2 || hs[0].start != 0 || hs[1].start != 6 || hs[1].end != 8 {
		t.Fatal("suffix highlights didn't migrate", hs)
	}
	d.SetRangeHighlights(nil)
	if len(d.selection.highlights) != 0 {
		t.Fatal("clear failed")
	}
}

func TestDocumentRevealRangeScrollsBothWays(t *testing.T) {
	d := New("first\n\n" + strings.Repeat("middle\n\n", 30) + "last")
	h := uitest.New(el.Root(scrollingDocView{d}))
	end := strings.Index(d.RenderedText(), "last")
	if !d.RevealRange(TextRange{end, end + 4}) {
		t.Fatal("reveal rejected")
	}
	h.Frame()
	h.Frame()
	_, y := textPoint(t, h, d, "last", 0)
	if y < 10 || y > 130 {
		t.Fatal("last outside viewport", y)
	}
	if !d.RevealRange(TextRange{0, 1}) {
		t.Fatal("reverse reveal rejected")
	}
	h.Frame()
	h.Frame()
	_, y = textPoint(t, h, d, "first", 0)
	if y < 10 || y > 130 {
		t.Fatal("first outside viewport", y)
	}
	d.RevealRange(TextRange{end, end})
	d.Append("!")
	if d.ranges.reveal != nil {
		t.Fatal("source change retained reveal")
	}
}

func TestDocumentPreviewScaleAndExpansion(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		d := New("# Heading\n\n" + strings.Repeat("body words 中文\n\n", 12)).MaxLines(3)
		root := el.Embed(d)
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			gtx.Constraints.Max = image.Pt(int(400*scale), int(1200*scale))
			root.Layout(gtx)
		})
		if !d.IsClamped() {
			t.Fatal("long document not clamped", scale)
		}
		d.MaxLines(0)
		h.Frame()
		if d.IsClamped() {
			t.Fatal("expanded document clamped", scale)
		}
		d.SetSource("short")
		d.MaxLines(3)
		h.Frame()
		if d.IsClamped() {
			t.Fatal("short document clamped", scale)
		}
	}
}

func TestDocumentStreamFadeAppendStyleAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	now := time.Now()
	d := New("old ").StreamFade(true)
	root := el.Embed(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	d.Append("new")
	h.Frame()
	if !d.fade.active || d.fade.alpha(now, d.fade.times[len(d.fade.times)-1]) != 0 {
		t.Fatal("new text didn't start transparent")
	}
	start := d.fade.times[len(d.fade.times)-1]
	now = now.Add(100 * time.Millisecond)
	d.Append("!")
	h.Frame()
	if d.fade.times[len(d.fade.times)-2] != start {
		t.Fatal("append restarted old fade")
	}
	now = now.Add(time.Second)
	h.Frame()
	if d.fade.active {
		t.Fatal("fade didn't settle")
	}
	d.SetSource("**bold")
	h.Frame()
	d.Append("**")
	h.Frame()
	if !d.fade.active {
		t.Fatal("completed syntax didn't fade changed styles")
	}
	theme.SetReducedMotion(true)
	h.Frame()
	if d.fade.active {
		t.Fatal("reduced motion kept transition")
	}
	d.SetSource("replacement")
	h.Frame()
	for _, at := range d.fade.times {
		if !at.IsZero() {
			t.Fatal("replacement faded")
		}
	}
}

func TestDocumentPluginsIsolationFallbackAndInlineCopy(t *testing.T) {
	clicked := 0
	object := InlineObject{Text: "@中文", Widget: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Role("button").Name("inline action").W(el.Dp(60)).H(el.Dp(20)).OnClick(func() { clicked++ }).Child(el.Text("@中文"))
	}))}
	p := Plugin{Inlines: map[ast.NodeKind]func(ast.Node, []byte) InlineObject{ast.KindCodeSpan: func(n ast.Node, b []byte) InlineObject {
		if string(n.Text(b)) == "tag" {
			return object
		}
		return InlineObject{}
	}}}
	d := New("before `tag` after `code`").Plugins(p)
	h := uitest.New(el.Root(docView{d}))
	if d.RenderedText() != "before @中文 after code" {
		t.Fatal("plugin or fallback text", d.RenderedText())
	}
	r := d.selection.parts[0].r
	found := false
	for _, piece := range r.rt.pieces {
		if r.rt.runs[piece.run].object != nil {
			found = true
			if piece.runes != 3 || len(piece.glyphs) != 1 {
				t.Fatal("inline object not atomic", piece)
			}
		}
	}
	if !found {
		t.Fatal("inline widget missing")
	}
	clickCodeButton(t, h, "inline action")
	if clicked != 1 {
		t.Fatal("inline action blocked", clicked)
	}
	x0, y0 := textPoint(t, h, d, "before @中文 after code", 7)
	x1, y1 := textPoint(t, h, d, "before @中文 after code", 10)
	h.Drag(x0-1, y0, x1, y1)
	assertSelectionCopy(t, h, d, "@中文")
	plain := New("`tag`")
	uitest.New(el.Root(docView{plain}))
	if plain.RenderedText() != "tag" {
		t.Fatal("plugin leaked")
	}
	d.Append(" tail")
	h.Frame()
	if !strings.HasSuffix(d.RenderedText(), "code tail") {
		t.Fatal("plugin append stale")
	}
	d.Plugins()
	h.Frame()
	if strings.Contains(d.RenderedText(), "@中文") {
		t.Fatal("plugin unregister stale")
	}
}

func TestDocumentExtensionPixels(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	for _, scale := range []int{1, 2} {
		gpu, err := headless.NewWindow(400*scale, 500*scale)
		if err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		d := New("old words").StreamFade(true)
		root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().WFull().HFull().Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).Child(d.Render(cx))
		}))
		var img *image.RGBA
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Now = now
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 500*scale)
			root.Layout(gtx)
			if err := gpu.Frame(gtx.Ops); err != nil {
				t.Fatal(err)
			}
			img = image.NewRGBA(image.Rect(0, 0, 400*scale, 500*scale))
			if err := gpu.Screenshot(img); err != nil {
				t.Fatal(err)
			}
		})
		before := img
		d.Append(" NEW")
		h.Frame()
		initial := img
		// Stable prefix keeps its pixels; appended text is initially invisible.
		// A one-level antialiasing difference at the split glyph edge is allowed.
		for i, a := range before.Pix {
			delta := int(a) - int(initial.Pix[i])
			if delta < -1 || delta > 1 {
				t.Fatal("settled text changed or new text visible at fade start", scale, i, a, initial.Pix[i])
			}
		}

		now = now.Add(time.Second)
		h.Frame()
		if string(initial.Pix) == string(img.Pix) {
			t.Fatal("fade never painted appended text", scale)
		}
		red := color.NRGBA{R: 255, A: 255}
		blue := color.NRGBA{B: 255, A: 255}
		if !d.SetRangeHighlights([]RangeHighlight{{TextRange{0, 9}, red}, {TextRange{4, 9}, blue}}) {
			t.Fatal("highlight rejected")
		}
		h.Frame()
		reds, blues := 0, 0
		for y := 0; y < 40*scale; y++ {
			for x := 0; x < 200*scale; x++ {
				c := img.RGBAAt(x, y)
				if c.R > 240 && c.G < 10 && c.B < 10 {
					reds++
				}
				if c.B > 240 && c.G < 10 && c.R < 10 {
					blues++
				}
			}
		}
		if reds == 0 || blues == 0 {
			t.Fatal("overlapping highlights missing", scale, reds, blues)
		}
		d.SetRangeHighlights(nil)
		d.SetSource("a first line\n\n| Head |\n| --- |\n| row |\n| next |\n\nlast")
		d.MaxLines(2)
		h.Frame()
		if !d.IsClamped() {
			t.Fatal("mixed table not clamped")
		}
		bottom := d.preview.height
		for y := bottom; y < 500*scale; y++ {
			for x := 0; x < 400*scale; x++ {
				c := img.RGBAAt(x, y)
				if c != (color.RGBA{255, 255, 255, 255}) {
					t.Fatalf("content beyond preview: scale=%d xy=%d,%d bottom=%d color=%v", scale, x, y, bottom, c)
				}
			}
		}
		gpu.Release()
	}
}

func TestDocumentPluginSyntaxAndBlockState(t *testing.T) {
	clicks := 0
	view := el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Role("button").Name("custom block").OnClick(func() { clicks++ }).Child(el.Text("custom block"))
	})
	calls := 0
	plugin := Plugin{Extensions: []goldmark.Extender{extension.DefinitionList}, Blocks: map[ast.NodeKind]func(ast.Node, []byte) el.View{east.KindDefinitionList: func(n ast.Node, source []byte) el.View { calls++; return view }}}
	d := New("> Term\n> : description\n\nplain").Plugins(plugin)
	h := uitest.New(el.Root(docView{d}))
	if calls == 0 || strings.Contains(d.RenderedText(), "description") {
		t.Fatal("custom syntax wasn't converted")
	}
	clickCodeButton(t, h, "custom block")
	if clicks != 1 {
		t.Fatal("custom block action failed")
	}
	d.Append(" tail")
	h.Frame()
	clickCodeButton(t, h, "custom block")
	if clicks != 2 || calls < 2 {
		t.Fatal("reparse lost app state", clicks, calls)
	}
	d.Plugins()
	h.Frame()
	if !strings.Contains(d.RenderedText(), "description") {
		t.Fatal("default parser wasn't restored", d.RenderedText())
	}
}
