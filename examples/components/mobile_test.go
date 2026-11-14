package main

import (
	"fmt"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/window"
	"github.com/dyike/keel/ui/window/windowtest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
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
