package main

import (
	"encoding/json"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/window"
)

// This opt-in diagnostic samples retained Go memory after collection. It does
// not measure a real window, native presentation, or total client memory.
func TestComponentMemoryLifecycle(t *testing.T) {
	path := os.Getenv("KEEL_MEMORY_REPORT")
	if path == "" {
		t.Skip("set KEEL_MEMORY_REPORT to run the headless lifecycle diagnostic")
	}
	previous := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(previous)
	gpu := os.Getenv("KEEL_MEMORY_GPU") == "1"
	type sample struct {
		Round                              int
		HeapAlloc, HeapObjects, TotalAlloc uint64
		NumGC                              uint32
		Goroutines                         int
	}
	rounds := 7
	if n, err := strconv.Atoi(os.Getenv("KEEL_MEMORY_ROUNDS")); err == nil && n > 0 {
		rounds = n
	}
	var samples []sample
	for round := 0; round < rounds; round++ {
		for cycle := 0; cycle < 5; cycle++ {
			for _, name := range []string{"virtual_list", "table", "input", "text_view"} {
				var section demoSection
				for _, candidate := range demoSections {
					if candidate.name == name {
						section = candidate
						break
					}
				}
				if section.build == nil {
					t.Fatalf("missing scene %s", name)
				}
				func() {
					h := window.OpenHeadless(window.Options{Content: section.build(), Width: 680, Height: 860})
					defer h.Close()
					h.Render()
					if len(h.Elements()) == 0 {
						t.Fatalf("%s has no visible elements", name)
					}
					if name == "virtual_list" {
						if err := h.Click("Scroll horizontally to the last item"); err != nil {
							t.Fatal(err)
						}
						if _, err := h.Find("Card 100000"); err != nil {
							t.Fatal("horizontal scroll did not execute: ", err)
						}
						firstRow := ""
						for step := 0; step < 20; step++ {
							found := false
							for _, e := range h.Elements() {
								if strings.HasPrefix(e.Name, "[INFO] Entry ") {
									if step == 0 {
										firstRow = e.Name
									}
									if step == 10 && e.Name == firstRow {
										t.Fatal("vertical scroll did not change visible rows")
									}
									dy := float32(72)
									if step >= 10 {
										dy = -dy
									}
									if err := h.Scroll(e.Name, dy); err != nil {
										t.Fatal(err)
									}
									found = true
									break
								}
							}
							if !found {
								t.Fatal("no visible virtual log row")
							}
						}
					}
					for _, size := range [][2]int{{820, 720}, {680, 860}} {
						h.Resize(size[0], size[1])
						if w, hh := h.Window().Size(); w != size[0] || hh != size[1] {
							t.Fatalf("resize did not execute: %dx%d", w, hh)
						}
						if gpu && cycle == 0 {
							if png, err := h.Screenshot(); err != nil || len(png) == 0 {
								t.Fatalf("GPU screenshot %s: %v", name, err)
							}
						}
					}
				}()
			}
		}
		// Collection is confined to diagnostic boundaries. This reports retained
		// objects after closing windows, not idle physical memory or production GC.
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		samples = append(samples, sample{round, m.HeapAlloc, m.HeapObjects, m.TotalAlloc, m.NumGC, runtime.NumGoroutine()})
	}
	time.Sleep(6 * time.Second)
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	samples = append(samples, sample{rounds, m.HeapAlloc, m.HeapObjects, m.TotalAlloc, m.NumGC, runtime.NumGoroutine()})
	if os.Getenv("KEEL_MEMORY_PROFILE") == "1" {
		f, err := os.Create(path + ".heap")
		if err != nil {
			t.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}
	data, err := json.MarshalIndent(samples, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
