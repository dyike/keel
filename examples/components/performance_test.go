package main

import (
	"encoding/json"
	"image"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
)

// The runner reads the actual registry rather than parsing Go source or keeping
// a second component list. Ordinary tests do not write a manifest.
func TestComponentManifest(t *testing.T) {
	path := os.Getenv("KEEL_COMPONENT_MANIFEST")
	if path == "" {
		t.Skip("requested by scripts/check-components.py")
	}
	type entry struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	var entries []entry
	for _, section := range demoSections {
		entries = append(entries, entry{section.name, displayName(section.name)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

// BenchmarkComponentFrame measures warmed layout and input registration for
// every gallery scene, without GPU submission or background update delivery.
// Run selected scenes in separate processes when comparing allocations: optional
// highlighting and image loading can otherwise spill into the next scene.
func BenchmarkComponentFrame(b *testing.B) {
	for _, section := range demoSections {
		b.Run(section.name, func(b *testing.B) {
			frame := componentFrame(section)
			b.ReportAllocs()
			for b.Loop() {
				frame()
			}
		})
	}
}

func componentFrame(section demoSection) func() {
	widget := section.build()
	var ops op.Ops
	var router input.Router
	frame := func() {
		ops.Reset()
		widget.Layout(layout.Context{
			Ops: &ops, Source: router.Source(), Now: time.Unix(0, 0),
			Constraints: layout.Constraints{Max: image.Pt(680, 1040)},
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		})
		router.Frame(&ops)
	}
	for range 3 {
		frame()
	}
	return frame
}

func TestSteadyFrameAllocationCeilings(t *testing.T) {
	ceilings := map[string]float64{
		"label":    35,
		"list":     250,
		"button":   260,
		"calendar": 1050,
		"table":    550,
		"input":    500,
	}
	for _, section := range demoSections {
		limit, ok := ceilings[section.name]
		if !ok {
			continue
		}
		t.Run(section.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(20, componentFrame(section)); got > limit {
				t.Fatalf("steady frame allocations %.0f exceed ceiling %.0f", got, limit)
			}
		})
		delete(ceilings, section.name)
	}
	for name := range ceilings {
		t.Errorf("allocation ceiling names missing scene %q", name)
	}
}

func TestGalleryCategoriesCoverEverySection(t *testing.T) {
	categories := map[string]bool{}
	for _, category := range categoryTitles() {
		categories[category.id] = true
	}
	names := map[string]bool{}
	for _, section := range demoSections {
		if names[section.name] {
			t.Errorf("duplicate gallery section %q", section.name)
		}
		names[section.name] = true
		category := section.category
		if foundations[section.name] {
			category = "foundations"
		}
		if !categories[category] {
			t.Errorf("%s is unreachable in the gallery: category %q has no sidebar section", section.name, category)
		}
	}
}
