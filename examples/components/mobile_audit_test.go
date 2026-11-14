package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/window"
)

// This opt-in census reports candidates, not a blanket accessibility pass:
// clipped horizontal scrollers and partially visible rows need visual review.
func TestGalleryMobileAudit(t *testing.T) {
	dir := os.Getenv("KEEL_MOBILE_AUDIT")
	if dir == "" {
		t.Skip("set KEEL_MOBILE_AUDIT to a report directory")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	type sample struct {
		Section      string
		Width        int
		TextScale    float32
		Elements     []window.Element
		Outside      []window.Element
		SmallTargets []window.Element
	}
	var report []sample
	for _, section := range demoSections {
		for _, width := range []int{320, 390, 600} {
			for _, scale := range []float32{1, 1.6} {
				t.Run(fmt.Sprintf("%s/%d/%.1f", section.name, width, scale), func(t *testing.T) {
					g := newGallery()
					g.nav.SetValue(section.name)
					g.mobileDetail = true
					h := window.OpenHeadless(window.Options{Width: width, Height: 844, TextScale: scale, Content: el.Root(g)})
					defer h.Close()
					s := sample{Section: section.name, Width: width, TextScale: scale, Elements: h.Elements()}
					for _, e := range s.Elements {
						if e.X < 0 || e.X+e.Width > width {
							s.Outside = append(s.Outside, e)
						}
						switch e.Role {
						case "button", "checkbox", "radio", "switch", "combobox", "textbox":
							if e.Width < 44 || e.Height < 44 {
								s.SmallTargets = append(s.SmallTargets, e)
							}
						}
					}
					if os.Getenv("KEEL_MOBILE_AUDIT_IMAGES") == "1" && width == 320 {
						png, err := h.Screenshot()
						if err != nil {
							t.Fatal(err)
						}
						if err = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s-%d-%.1f.png", section.name, width, scale)), png, 0600); err != nil {
							t.Fatal(err)
						}
					}
					report = append(report, s)
				})
			}
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "audit.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
