package el

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/ui/theme"
)

func TestInlineStyleCopiesAreIndependent(t *testing.T) {
	base := Style{}
	base.Bg(theme.Primary).TextColor(theme.Text)
	copy := base
	copy.Bg(color.NRGBA{}).TextColor(color.NRGBA{})
	if base.bg.value != theme.Primary || base.text.color.value != theme.Text {
		t.Fatal("a visual copy modified its source")
	}
	if !copy.bg.set || !copy.text.color.set {
		t.Fatal("transparent overrides became absent")
	}
	parent := Text("parent").Bold().Mono().Italic().TextAlign(End).TextColor(theme.Text).node().style.text
	child := textStyle{weight: some(font.Normal), mono: some(false), italic: some(false), align: some(Start), color: some(color.NRGBA{})}
	got := child.inherit(parent)
	if got != child {
		t.Fatal("explicit zero-valued overrides inherited parent values")
	}
	var retained *Style
	result, current := applyVisualStyle(func(s *Style) { retained = s; s.Bg(theme.Primary) }, Style{}, nil)
	_, same := applyVisualStyle(func(s *Style) { s.Bg(theme.Danger) }, result, current)
	if same != retained || retained.bg.value != theme.Danger {
		t.Fatal("callbacks in one paint no longer share their visual copy")
	}
	applyVisualStyle(func(s *Style) { s.Bg(theme.Danger) }, Style{}, nil)
	retained.Bg(theme.Success)
	if result.bg.value != theme.Primary {
		t.Fatal("callback argument or returned style was reused/aliased")
	}
}

func TestSmallFrameHandleBatchesRemainIndependent(t *testing.T) {
	var handles frameHandles[DivEl]
	first := handles.alloc()
	if len(handles.batch) != 4 {
		t.Fatal("small root reserved a full handle chunk")
	}
	first.generation = 1
	handles.reset()
	second := handles.alloc()
	second.generation = 2
	if first == second || first.generation != 1 {
		t.Fatal("a prior pointer handle was overwritten")
	}
	for range 100 {
		handles.alloc()
	}
	handles.reset()
	handles.alloc()
	if len(handles.batch) != nodeChunkSize {
		t.Fatal("large root lost bounded batching")
	}
}

func TestTextMeasurementHashCollisionsAndReferenceRelease(t *testing.T) {
	var e engine
	a := textMeasureCacheKey{textMeasureKey: textMeasureKey{value: "first"}}
	b := textMeasureCacheKey{textMeasureKey: textMeasureKey{value: "second"}}
	da, db := layout.Dimensions{Size: image.Pt(1, 2)}, layout.Dimensions{Size: image.Pt(3, 4)}
	e.cacheTextMeasurement(7, a, da)
	e.cacheTextMeasurement(7, b, db)
	for _, tc := range []struct {
		key  textMeasureCacheKey
		want layout.Dimensions
	}{{a, da}, {b, db}} {
		if got, ok := e.cachedTextMeasurement(7, tc.key); !ok || got != tc.want {
			t.Fatal("hash collision reused another label's dimensions")
		}
	}
	entries := e.textMeasurementEntries
	e.beginTextMeasurements()
	for _, entry := range entries {
		if entry != (textMeasurement{}) {
			t.Fatal("frame cache retained prior text/shaper references")
		}
	}
	if got := testing.AllocsPerRun(20, func() {
		e.beginTextMeasurements()
		e.cacheTextMeasurement(7, a, da)
		e.cacheTextMeasurement(7, b, db)
	}); got != 0 {
		t.Fatalf("warmed measurement insertions allocated %.0f times", got)
	}
}

func TestDocumentPresentationCacheTracksContentNotRevision(t *testing.T) {
	token := InputToken{ID: "cat", Text: "@cat", Label: "Cat"}
	content, err := NewInputContent("@cat", InputTokenSpan{Range: InputRange{Start: 0, End: 4}, Token: token})
	if err != nil {
		t.Fatal(err)
	}
	var d InputDocument
	d.SetContent(content)
	plain := d.presentation()
	if plain.Text != "Cat" {
		t.Fatal(plain.Text)
	}
	d.objects = true
	objects := d.presentation()
	if objects.Text == plain.Text || len(objects.Spans) != 1 {
		t.Fatal("object mode did not invalidate presentation")
	}
	if got := testing.AllocsPerRun(20, func() { d.presentation() }); got != 0 {
		t.Fatalf("unchanged presentation allocated %.0f times", got)
	}
	// The editor changes the session without going through document setters.
	if err := d.session.ReplaceSource(InputRange{Start: 0, End: 4}, "plain"); err != nil {
		t.Fatal(err)
	}
	if p := d.presentation(); p.Text != "plain" || len(p.Spans) != 0 {
		t.Fatal("direct session edit kept stale tokens")
	}
	d.session.Undo()
	if p := d.presentation(); p.Text != objects.Text || len(p.Spans) != 1 {
		t.Fatal("undo did not restore token layout")
	}
	// Equal text with reference metadata removed must also invalidate.
	d.session.SetText("@cat")
	if p := d.presentation(); p.Text != "@cat" || len(p.Spans) != 0 {
		t.Fatal("token removal retained stale geometry")
	}
	if plain.Text != "Cat" || objects.Spans[0].Token != token {
		t.Fatal("previous presentation was mutated")
	}
	var st elemState
	filters := documentFilters(&st)
	if len(filters) != 13 {
		t.Fatal("document filter set changed")
	}
	if got := testing.AllocsPerRun(20, func() { documentFilters(&st) }); got != 0 {
		t.Fatal("stable editor rebuilt filters")
	}
}
