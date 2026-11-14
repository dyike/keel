package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/window"
	"github.com/dyike/keel/ui/window/windowtest"
)

func TestGalleryMobileNavigationAndResize(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, width := range []int{320, 390, 600} {
		g := newGallery()
		app := windowtest.New(t, window.Options{Width: width, Height: 780, Content: el.Root(g)})
		app.Want("Keel components")
		app.Click("Search components")
		app.Type("button")
		app.Click("Button")
		app.Want("Back to components")
		first := g.built["button"]
		if first == nil {
			t.Fatal("section not built")
		}
		back := app.Find("Back to components")
		if back.Height < 44 || back.Width < 44 {
			t.Fatalf("small back target: %+v", back)
		}
		app.Click("Back to components")
		if g.search.Value() != "button" {
			t.Fatal("query lost")
		}
		app.Click("Button")
		if g.built["button"] != first {
			t.Fatal("section state discarded")
		}
		app.Resize(1200, 800)
		app.WantNot("Back to components")
		app.Resize(width, 780)
		app.Want("Back to components")
		if g.built["button"] != first {
			t.Fatal("resize discarded state")
		}
	}
}
func TestGalleryMobileEmptySearchAndLanguage(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	g := newGallery()
	app := windowtest.New(t, window.Options{Width: 320, Height: 780, Content: el.Root(g)})
	app.Click("Search components")
	app.Type("no-such-component")
	app.Want("No components found")
	app.Click("中文")
	app.Want("没有找到组件")
	if g.mobileDetail {
		t.Fatal("language change left index")
	}
}

// The optional captures are for visual review, not golden pixel assertions
// across platform fonts. Navigation/input still uses the real event router.
func TestGalleryMobileCorePages(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, name := range []string{"button", "input", "select", "form"} {
		t.Run(name, func(t *testing.T) {
			g := newGallery()
			g.nav.SetValue(name)
			g.mobileDetail = true
			app := windowtest.New(t, window.Options{Width: 390, Height: 844, Content: el.Root(g)})
			app.Want("Back to components")
			if dir := os.Getenv("KEEL_GALLERY_CAPTURE"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("gallery-mobile-%s.png", name)), app.Screenshot(), 0600); err != nil {
					t.Fatal(err)
				}
			}
			app.Click("Back to components")
			app.Want("Search components")
		})
	}
}

func TestGalleryBundleRoutePreservesPreferences(t *testing.T) {
	target := galleryBundleURL("/preview/demo/", "?lang=en&theme=light&section=button&custom=1", "input", "inputs", "zh-CN", "dark")
	u, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Path != "/preview/demo/" || u.Fragment != "component=input" || q.Get("bundle") != "inputs" || q.Get("lang") != "zh-CN" || q.Get("theme") != "dark" || q.Get("custom") != "1" || q.Has("section") {
		t.Fatal(target)
	}
}

func TestGalleryMobileOverlays(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, width := range []int{320, 390} {
		for _, tc := range []struct{ section, trigger, content, close string }{
			{"dialog", "Custom", "Edit customer", "Cancel"},
			{"sheet", "Details on the right", "Order details", "Done"},
			{"sheet", "Bottom panel", "Bulk actions", "Close panel"},
			{"popover", "Filters", "Filter conditions", ""},
			{"select", "Status", "Awaiting payment", "Awaiting payment"},
		} {
			t.Run(fmt.Sprintf("%s/%s/%d", tc.section, tc.trigger, width), func(t *testing.T) {
				g := newGallery()
				g.nav.SetValue(tc.section)
				g.mobileDetail = true
				app := windowtest.New(t, window.Options{Width: width, Height: 640, Content: el.Root(g)})
				e := app.Find(tc.trigger)
				if e.X < 0 || e.X+e.Width > width {
					t.Fatalf("trigger outside viewport: %+v", e)
				}
				if (tc.section == "dialog" || tc.section == "sheet") && e.Height < 44 {
					t.Fatalf("small touch target: %+v", e)
				}
				app.Click(tc.trigger)
				app.Want(tc.content)
				if tc.close != "" {
					app.Click(tc.close)
				} else {
					app.Press("escape")
				}
				app.Click("Back to components")
				app.Want("Search components")
			})
		}
	}
}

func TestGalleryResponsiveGrid(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, tc := range []struct{ width, columns int }{{320, 1}, {600, 2}, {1000, 3}} {
		t.Run(fmt.Sprint(tc.width), func(t *testing.T) {
			content, ok := sectionContent("layout")
			if !ok {
				t.Fatal("missing layout section")
			}
			app := windowtest.New(t, window.Options{Width: tc.width, Height: 1600, Content: content})
			first := app.Find("Card 1")
			for i := 2; i <= tc.columns; i++ {
				next := app.Find(fmt.Sprintf("Card %d", i))
				if next.Y != first.Y || next.X <= first.X {
					t.Fatalf("expected %d columns: first=%+v next=%+v", tc.columns, first, next)
				}
			}
			next := app.Find(fmt.Sprintf("Card %d", tc.columns+1))
			if next.Y <= first.Y || next.X != first.X {
				t.Fatalf("next row misplaced: first=%+v next=%+v", first, next)
			}
			if dir := os.Getenv("KEEL_GALLERY_CAPTURE"); dir != "" {
				if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("gallery-layout-%d.png", tc.width)), app.Screenshot(), 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestGalleryMobileDialogInputAndShortViewport(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	g := newGallery()
	g.nav.SetValue("dialog")
	g.mobileDetail = true
	app := windowtest.New(t, window.Options{Width: 320, Height: 640, Content: el.Root(g)})
	app.Click("Custom")
	app.Click("East China Logistics")
	app.Press("mod+a")
	app.Type("Mobile customer")
	// A shorter viewport models the space available above a software keyboard;
	// this does not emulate a browser IME or establish physical-phone support.
	app.Resize(320, 360)
	app.Want("Mobile customer")
	app.Click("Save")
	app.Want("Saved: Mobile customer")
	app.Resize(320, 640)
	app.Click("Back to components")
	app.Click("Search components")
	app.Type("dialog")
	app.Click("Dialog")
	app.Want("Saved: Mobile customer")
	app.Click("Custom")
	app.Want("Mobile customer")
	app.Press("escape")
	app.Click("Back to components")
	app.Want("Search components")
}

func TestGalleryMobileAvatarGeometry(t *testing.T) {
	content, ok := sectionContent("avatar")
	if !ok {
		t.Fatal("missing avatar")
	}
	app := windowtest.New(t, window.Options{Width: 320, Height: 640, TextScale: 1, Content: content})
	count := 0
	for _, e := range app.Elements() {
		if e.Role != "avatar" {
			continue
		}
		count++
		if e.Width != e.Height || e.Width < 32 {
			t.Fatalf("squeezed avatar: %+v", e)
		}
	}
	if count != 8 {
		t.Fatalf("visible avatars: %d", count)
	}
}

func TestFramelessDemoResponsiveSidebarPreservesInput(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	demo := newFramelessDemo()
	app := windowtest.New(t, window.Options{Width: 820, Height: 600, Content: el.Root(demo)})
	if demo.nav.Collapsed() {
		t.Fatal("desktop sidebar starts collapsed")
	}
	app.Click("Search")
	app.Type("kept query")
	app.Resize(320, 600)
	if !demo.nav.Collapsed() {
		t.Fatal("narrow sidebar did not collapse")
	}
	if demo.search.Value() != "kept query" {
		t.Fatal("resize discarded input")
	}
	app.Resize(820, 600)
	if demo.nav.Collapsed() {
		t.Fatal("desktop sidebar preference lost")
	}
	demo.nav.SetCollapsed(true)
	app.Render()
	app.Resize(320, 600)
	app.Resize(820, 600)
	if !demo.nav.Collapsed() {
		t.Fatal("explicit desktop collapse preference lost")
	}
}

func TestResizableGroupDemoResponsiveState(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	content, ok := sectionContent("resizable_group")
	if !ok {
		t.Fatal("missing resizable group")
	}
	app := windowtest.New(t, window.Options{Width: 820, Height: 844, Content: content})
	app.Click("Edit content")
	app.Type("keep across orientations")
	app.Click("Show / hide preview")
	app.Resize(320, 844)
	app.Want("keep across orientations")
	app.WantNot("Preview")
	app.Click("Show / hide preview")
	app.Want("Preview")
	app.Resize(820, 844)
	app.Want("keep across orientations")
	app.Want("Preview")
}
