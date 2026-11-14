package main

import (
	"os"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"image"
	"time"
)

func TestGalleryCategoriesAndMotionCoverage(t *testing.T) {
	categories := map[string]bool{}
	for _, c := range categoryTitles() {
		if categories[c.id] {
			t.Fatal("duplicate category", c.id)
		}
		categories[c.id] = true
	}
	sections := map[string]bool{}
	for _, s := range demoSections {
		if !categories[s.category] {
			t.Error("unlisted category", s.name, s.category)
		}
		if sections[s.name] {
			t.Error("duplicate section", s.name)
		}
		sections[s.name] = true
		if _, err := os.Stat("../../" + sectionDoc(s)); err != nil {
			t.Errorf("%s: invalid documentation: %v", s.name, err)
		}
		if strings.HasPrefix(s.name, "motion_") {
			if s.category != "motion" {
				t.Error("motion topic in wrong category", s.name)
			}
			componentFrame(s)() // Includes each migrated topic's initial layout and paint.
		}
	}
	for _, name := range []string{"motion_values", "motion_gestures", "motion_layout", "motion_scroll", "motion_presets", "motion_pages", "motion_pose", "motion_timeline", "motion_effects", "rising_glow", "shimmer_text"} {
		if !sections[name] {
			t.Error("missing merged motion example", name)
		}
	}
}

func TestGalleryRetainsPagesAndResizedWidth(t *testing.T) {
	previous := locale.Current()
	defer locale.Apply(previous)
	locale.Apply(locale.English())
	g := newGallery()
	root := el.Root(g)
	var ops op.Ops
	var router input.Router
	frame := func() {
		ops.Reset()
		root.Layout(layout.Context{Ops: &ops, Source: router.Source(), Now: time.Unix(100, 0), Constraints: layout.Exact(image.Pt(1100, 800)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}})
		router.Frame(&ops)
	}
	frame()
	g.nav.SetValue("motion_values")
	frame()
	first := g.built["motion_values"]
	if first == nil {
		t.Fatal("motion section not built")
	}
	g.nav.SetValue("button")
	frame()
	g.nav.SetValue("motion_values")
	frame()
	if g.built["motion_values"] != first {
		t.Fatal("switching recreated section")
	}
	// Same state written by the sidebar's OnResize callback.
	g.navWidth = 310
	g.nav.Width(310)
	g.search.SetValue("Motion")
	g.nav.Filter("Motion")
	locale.Apply(locale.Chinese())
	frame()
	if g.nav.CurrentWidth() != 310 || g.nav.Value() != "motion_values" || g.search.Value() != "Motion" {
		t.Fatal("language change lost navigation state")
	}
}
