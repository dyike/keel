package el

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func contentState(root *RootWidget) *textContentState {
	for _, st := range root.store.states {
		if st.content != nil {
			return st.content
		}
	}
	return nil
}

func TestContentPaintOnlyLifecycleAndSemantics(t *testing.T) {
	now := time.Unix(1500, 0)
	value := "旧内容 12"
	renders := 0
	root := Root(ViewFunc(func(*Context) Element {
		renders++
		return Div().W(Dp(300)).Items(Start).Child(Text(value).ID("label").ContentTransition(ContentTransition{Morph: true, Animation: Tween(time.Second, Linear)}))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	if s := contentState(root); s == nil || len(s.tracks) != 0 || root.motionDue {
		t.Fatal("first declaration animated or retained glyphs")
	}
	core.Update(func() { value = "新内容 21" })
	h.Frame()
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before || !root.motionDue || root.repaints == 0 {
		t.Fatal("text transition did not use retained frames")
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		if n.Desc.Label == "旧内容 12" {
			t.Fatal("old text leaked to semantics")
		}
	}
	now = now.Add(time.Second)
	h.Frame()
	if s := contentState(root); s == nil || len(s.tracks) != 0 || s.last.value != value || root.motionDue {
		t.Fatal("settled text retained old content or requested frames")
	}
}

func TestContentPixelsContinuousAndRetainedMatch(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	for _, options := range []ContentTransition{{}, {Numeric: true}, {Morph: true}} {
		t.Run(fmt.Sprintf("numeric=%v/morph=%v", options.Numeric, options.Morph), func(t *testing.T) {
			options.Animation = Tween(time.Second, Linear)
			now := time.Unix(1600, 0)
			value := "Café 中文 123 👩🏽‍💻"
			animated := true
			root := Root(ViewFunc(func(*Context) Element {
				label := Text(value).ID("label").TextSize(24).Ranges(TextRange{Start: 0, End: 5, Color: color.NRGBA{R: 220, A: 255}})
				if animated {
					label.ContentTransition(options)
				}
				return Div().Size(Full).Bg(theme.Bg).P(16).Child(label)
			}))
			root.RetainFrames()
			var shot []byte
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				root.Layout(gtx)
				if err := gpu.Frame(gtx.Ops); err != nil {
					t.Fatal(err)
				}
				img := image.NewRGBA(image.Rect(0, 0, 400, 300))
				if err := gpu.Screenshot(img); err != nil {
					t.Fatal(err)
				}
				shot = img.Pix
			})
			initial := slices.Clone(shot)
			core.Update(func() { value = "中文 Café 132 👩🏽‍💻" })
			h.Frame()
			if !bytes.Equal(initial, shot) {
				t.Fatalf("transition start changed old pixels (%d differing bytes)", pixelByteDifference(initial, shot))
			}
			now = now.Add(400 * time.Millisecond)
			h.Frame()
			mid := slices.Clone(shot)
			saveContentTestFrame(t, "mid", shot)
			loop.Changed()
			h.Frame()
			if !bytes.Equal(mid, shot) {
				t.Fatal("retained/full pixels differ")
			}
			core.Update(func() { value = "再次切换 Café 321 👩🏽‍💻" })
			h.Frame()
			if !bytes.Equal(mid, shot) {
				t.Fatalf("interrupted text jumped (%d differing bytes)", pixelByteDifference(mid, shot))
			}
			now = now.Add(2 * time.Second)
			h.Frame()
			settled := slices.Clone(shot)
			core.Update(func() { animated = false })
			h.Frame()
			if !bytes.Equal(settled, shot) {
				t.Fatal("settled animation differs from ordinary text")
			}
		})
	}
}

func pixelByteDifference(a, b []byte) int {
	n := 0
	for i, v := range a {
		if v != b[i] {
			n++
		}
	}
	return n
}

func TestContentWholeClustersAndRangeColors(t *testing.T) {
	value := "é office 👩🏽‍💻 العربية 中文"
	root := Root(ViewFunc(func(*Context) Element {
		return Text(value).ID("text").Ranges(TextRange{Start: 0, End: 1, Color: color.NRGBA{R: 255, A: 255}}).ContentTransition(ContentTransition{Morph: true})
	}))
	uitest.New(root)
	s := contentState(root)
	if s == nil {
		t.Fatal("no snapshot")
	}
	gs := s.last.shape()
	var combined, emoji bool
	for _, g := range gs {
		if g.key == "é" {
			combined = true
			if g.color.R != 255 || g.color.G != 0 {
				t.Fatal("range split combining cluster")
			}
		}
		if g.key == "👩🏽‍💻" {
			emoji = true
		}
		if len(g.glyphs) == 0 || g.glyphs[len(g.glyphs)-1].Flags&text.FlagClusterBreak == 0 {
			t.Fatal("partial shaping cluster")
		}
	}
	if !combined || !emoji {
		t.Fatalf("split Unicode sequence: accent=%v emoji=%v keys=%v", combined, emoji, clusterKeys(gs))
	}
}

func clusterKeys(gs []contentGlyph) []string {
	out := make([]string, len(gs))
	for i, g := range gs {
		out[i] = g.key
	}
	return out
}

func TestContentNumericPairsAndMorphReorder(t *testing.T) {
	makeGlyphs := func(s string) []contentGlyph {
		var gs []contentGlyph
		for i, r := range s {
			gs = append(gs, contentGlyph{key: string(r), position: f32.Pt(float32(i*10), 0), scale: 1, alpha: 1})
		}
		return gs
	}
	from, to := makeGlyphs("$19.00"), makeGlyphs("$20.00")
	if got := contentPairs(from, to, ContentTransition{Numeric: true}); !slices.Equal(got, []int{0, -1, -1, 3, 4, 5}) {
		t.Fatalf("numeric pairs=%v", got)
	}
	if got := contentPairs(makeGlyphs("abc"), makeGlyphs("cab"), ContentTransition{Morph: true}); !slices.Equal(got, []int{2, 0, 1}) {
		t.Fatalf("morph pairs=%v", got)
	}
	from[1].position.Y = 20
	tracks := pairContent(from, to, ContentTransition{Numeric: true, Animation: Smooth}, time.Unix(1, 0), textContentSnapshot{})
	for _, track := range tracks {
		if track.from.position != track.to.position {
			t.Fatal("multiline numeric fallback moved a cluster")
		}
	}
}

func TestContentReleasesOffscreenOmittedAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	now := time.Unix(1700, 0)
	value := "old"
	y := float32(0)
	enabled := true
	root := Root(ViewFunc(func(*Context) Element {
		label := Text(value).ID("text")
		if enabled {
			label.ContentTransition(ContentTransition{Morph: true, Animation: Tween(time.Second, Linear)})
		}
		return Div().Items(Start).Child(Div().ID("container").Translate(0, y).Child(label))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	value = "new"
	h.Frame()
	if len(contentState(root).tracks) == 0 {
		t.Fatal("no transition")
	}
	y = 1000
	h.Frame()
	if contentState(root) != nil || len(root.e.contentStates) != 0 {
		t.Fatal("offscreen ancestor retained glyphs")
	}
	y = 0
	h.Frame()
	if len(contentState(root).tracks) != 0 {
		t.Fatal("return to viewport animated stale text")
	}
	value = "third"
	h.Frame()
	theme.SetReducedMotion(true)
	h.Frame()
	if len(contentState(root).tracks) != 0 || root.motionDue {
		t.Fatal("reduced motion retained animation")
	}
	enabled = false
	h.Frame()
	if contentState(root) != nil || len(root.e.contentStates) != 0 {
		t.Fatal("omission retained content")
	}
}

func TestContentEmptyTextFinishesExitThroughCollapsedParents(t *testing.T) {
	now := time.Unix(1800, 0)
	value := "visible"
	root := Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Div().Items(Start).Child(Text(value).ID("label").ContentTransition(ContentTransition{Animation: Tween(time.Second, Linear)})))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	core.Update(func() { value = "" })
	h.Frame()
	s := contentState(root)
	if s == nil || len(s.tracks) == 0 || !root.motionDue {
		t.Fatal("empty text skipped its exit")
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if root.motionDue || contentState(root) != nil && len(contentState(root).tracks) > 0 {
		t.Fatal("empty text retained departing glyphs")
	}
}

func TestContentRepeatedRetargetingExpiresOldGlyphs(t *testing.T) {
	now := time.Unix(1900, 0)
	value := "before"
	root := Root(ViewFunc(func(*Context) Element {
		return Text(value).ID("text").ContentTransition(ContentTransition{Morph: true, Animation: Tween(time.Second, Linear)})
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	for i := 0; i < 180; i++ {
		value = fmt.Sprintf("%c", rune(0x4e00+i))
		now = now.Add(16 * time.Millisecond)
		h.Frame()
		if s := contentState(root); s != nil && len(s.tracks) > 65 {
			t.Fatalf("old transitions restarted their lifetime: %d tracks", len(s.tracks))
		}
	}
	now = now.Add(2 * time.Second)
	h.Frame()
	if s := contentState(root); s == nil || len(s.tracks) != 0 || s.last.value != value {
		t.Fatal("old glyphs survived settling")
	}
}

func saveContentTestFrame(t *testing.T, stage string, pixels []byte) {
	t.Helper()
	dir := os.Getenv("KEEL_CONTENT_ARTIFACTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	name := strings.NewReplacer("/", "-", "=", "-").Replace(t.Name()) + "-" + stage + ".png"
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	img := &image.RGBA{Pix: pixels, Stride: 400 * 4, Rect: image.Rect(0, 0, 400, 300)}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestContentGoFallbackKeepsGraphemeClusters(t *testing.T) {
	root := Root(ViewFunc(func(*Context) Element {
		return Text("é 👩🏽‍💻").ID("text").ContentTransition(ContentTransition{Morph: true})
	}))
	uitest.New(root)
	snapshot := contentState(root).last
	snapshot.shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	gs := snapshot.shape()
	keys := clusterKeys(gs)
	if !slices.Contains(keys, "é") || !slices.Contains(keys, "👩🏽‍💻") {
		t.Fatalf("fallback font split grapheme clusters: %q", keys)
	}
}

func TestContentSlowSpringKeepsOriginalExitClock(t *testing.T) {
	now := time.Unix(2000, 0)
	source := contentGlyph{alpha: 1, scale: 1}
	dest := source
	dest.alpha = 0
	tr := Spring(1, 2)
	track := departingContent(source, dest, now, tr, make(map[float32]*contentExit))
	visible := track.at(.1, now.Add(3*time.Second))
	if visible.alpha < .1 {
		t.Fatalf("slow spring was cut short: alpha=%v", visible.alpha)
	}
	dest = visible
	dest.alpha = 0
	next := departingContent(visible, dest, now.Add(3*time.Second), tr, make(map[float32]*contentExit))
	got := next.at(.2, now.Add(4*time.Second)).alpha
	want := float32(5 * math.Exp(-4))
	if math.Abs(float64(got-want)) > .001 {
		t.Fatalf("exit restarted: alpha=%v want=%v", got, want)
	}
}
