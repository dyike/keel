package main

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

// Use the exact -section category composition, not independent scene roots.
// Shared icons/shapers alternate colors and sizes across all its sections.
// This includes CPU paint-op generation, but not GPU execution or OS footprint.
func BenchmarkCategoryFrame(b *testing.B) {
	for _, category := range []string{"controls", "inputs", "overlays", "data", "shell"} {
		b.Run(category, func(b *testing.B) {
			content, ok := sectionContent(category)
			if !ok {
				b.Fatal("missing category", category)
			}
			var ops op.Ops
			var router input.Router
			frame := 0
			draw := func() {
				ops.Reset()
				content.Layout(layout.Context{
					Ops: &ops, Source: router.Source(),
					Now:         time.Unix(0, 0).Add(time.Duration(frame) * time.Second / 60),
					Constraints: layout.Exact(image.Pt(1360, 1720)),
					Metric:      unit.Metric{PxPerDp: 2, PxPerSp: 2},
				})
				router.Frame(&ops)
				frame++
			}
			for range 5 {
				draw()
			}
			b.ReportAllocs()
			for b.Loop() {
				draw()
			}
		})
	}
}

// Stateful views in Render must keep their pointer identity. Otherwise a theme
// preview loses focus (and input) on the very next frame.
func TestThemePreviewKeepsEditorFocusAcrossFrames(t *testing.T) {
	v := &themeGallery{inputs: [2]*kit.InputView{kit.Input("first"), kit.Input("second")}}
	content := el.Embed(v)
	var ops op.Ops
	var router input.Router
	draw := func() {
		ops.Reset()
		content.Layout(layout.Context{Ops: &ops, Source: router.Source(), Now: time.Unix(0, 0),
			Constraints: layout.Exact(image.Pt(680, 1040)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}})
		router.Frame(&ops)
	}
	draw()
	var at image.Rectangle
	var find func(input.SemanticNode)
	find = func(n input.SemanticNode) {
		if at.Empty() && n.Desc.Class.String() == "Editor" {
			at = n.Desc.Bounds
		}
		for _, child := range n.Children {
			find(child)
		}
	}
	for _, n := range router.AppendSemantics(nil) {
		find(n)
	}
	if at.Empty() {
		t.Fatal("missing preview editor")
	}
	pos := f32.Pt(float32(at.Min.X+8), float32(at.Min.Y+8))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: pos},
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: pos},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: pos})
	draw()
	draw()
	router.Queue(key.EditEvent{Text: "kept"})
	draw()
	draw()
	if got := v.inputs[0].Value(); got != "kept" {
		t.Fatalf("preview editor lost focus or text across frames: %q", got)
	}
	router.Queue(key.EditEvent{Range: key.Range{Start: 4, End: 4}, Text: " again"})
	draw()
	draw()
	if got := v.inputs[0].Value(); got != "kept again" {
		t.Fatalf("typing replaced the preview editor: %q", got)
	}
}
