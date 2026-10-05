package main

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var (
	hrefRE = regexp.MustCompile(`(?:href|src|data-src)="([^"]+)"`)
	idRE   = regexp.MustCompile(`id="([^"]+)"`)
)

// The whole site builds from the repository, and every link inside it
// reaches an existing page and anchor.
func TestSiteLinksResolve(t *testing.T) {
	out := t.TempDir()
	s := &site{root: "../..", out: out, repo: "https://github.com/dyike/keel", branch: "main"}
	if err := s.build(""); err != nil {
		t.Fatal(err)
	}
	for _, w := range s.warnings {
		t.Error(w)
	}
	ids := map[string]map[string]bool{}
	anchors := func(file string) map[string]bool {
		if m, ok := ids[file]; ok {
			return m
		}
		m := map[string]bool{}
		data, _ := os.ReadFile(file)
		for _, x := range idRE.FindAllStringSubmatch(string(data), -1) {
			m[x[1]] = true
		}
		ids[file] = m
		return m
	}
	pages := 0
	filepath.WalkDir(out, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") {
			return err
		}
		pages++
		data, _ := os.ReadFile(p)
		for _, m := range hrefRE.FindAllStringSubmatch(string(data), -1) {
			u, err := url.Parse(m[1])
			if err != nil || u.Scheme != "" || strings.HasPrefix(m[1], "//") {
				continue
			}
			target := p
			if u.Path != "" {
				target = filepath.Join(filepath.Dir(p), filepath.FromSlash(u.Path))
				if strings.HasSuffix(u.Path, "/") {
					target = filepath.Join(target, "index.html")
				}
				if strings.Contains(u.Path, "demo/") {
					continue // published by the deploy workflow
				}
				if _, err := os.Stat(target); err != nil {
					rel, _ := filepath.Rel(out, p)
					t.Errorf("%s: missing %s", rel, m[1])
					continue
				}
			}
			if f, err := url.PathUnescape(u.Fragment); err == nil && f != "" && strings.HasSuffix(target, ".html") && !anchors(target)[f] {
				rel, _ := filepath.Rel(out, p)
				t.Errorf("%s: missing anchor %s", rel, m[1])
			}
		}
		return nil
	})
	if pages < 100 {
		t.Fatalf("only %d pages", pages)
	}
	dock, err := os.ReadFile(filepath.Join(out, "docs/kit/dock.html"))
	if err != nil || !strings.Contains(string(dock), `section=dock`) {
		t.Fatal("the Dock page does not embed its demo")
	}
}

func TestRelURL(t *testing.T) {
	for _, c := range []struct{ from, to, want string }{
		{"index.html", "docs/el.html", "docs/el.html"},
		{"docs/el.html", "docs/app.html", "app.html"},
		{"docs/kit/dock.html", "docs/el.html", "../el.html"},
		{"docs/kit/dock.html", ".", "../../"},
		{"index.html", ".", "./"},
		{"ui/kit/index.html", "docs/kit/dock.html", "../../docs/kit/dock.html"},
	} {
		if got := relURL(c.from, c.to); got != c.want {
			t.Errorf("relURL(%q, %q) = %q, want %q", c.from, c.to, got, c.want)
		}
	}
}

func TestGallerySourceAndMergedGuide(t *testing.T) {
	s := &site{root: "../..", out: t.TempDir(), repo: "https://github.com/dyike/keel", branch: "main"}
	if err := s.build(""); err != nil {
		t.Fatal(err)
	}
	for _, p := range s.order {
		if !strings.HasPrefix(p.Src, "docs/kit/") {
			continue
		}
		t.Run(p.Src, func(t *testing.T) {
			if p.Section == "" || p.DemoFile == "" {
				t.Fatal("component has no gallery source")
			}
			data, err := os.ReadFile(filepath.Join(s.root, p.DemoFile))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), `registerSection("`+p.Section+`"`) {
				t.Fatal("source does not register the displayed example")
			}
			page, err := os.ReadFile(filepath.Join(s.out, p.Out))
			if err != nil {
				t.Fatal(err)
			}
			doc, err := html.Parse(strings.NewReader(string(page)))
			if err != nil {
				t.Fatal(err)
			}
			var code strings.Builder
			var source, copyButton bool
			var visit func(*html.Node, bool, bool)
			visit = func(n *html.Node, inSource, inCode bool) {
				for _, a := range n.Attr {
					if n.Data == "details" && a.Key == "class" && a.Val == "demo-source" {
						inSource, source = true, true
					}
					if inSource && n.Data == "button" && a.Key == "class" && a.Val == "copy" {
						copyButton = true
					}
				}
				inCode = inCode || inSource && n.Data == "pre"
				if inCode && n.Type == html.TextNode {
					code.WriteString(n.Data)
				}
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					visit(child, inSource, inCode)
				}
			}
			visit(doc, false, false)
			if !source || !copyButton || strings.TrimSpace(code.String()) != strings.TrimSpace(string(data)) {
				t.Fatal("rendered example must include a copy button and the complete, current source")
			}
		})
	}
	if s.pages["docs/cli.md"] != nil {
		t.Error("scaffold instructions should share the getting-started page")
	}
	redirect, err := os.ReadFile(filepath.Join(s.out, "docs/cli.html"))
	if err != nil || !strings.Contains(string(redirect), `getting-started.html`) || !strings.Contains(string(redirect), `location.hash`) {
		t.Error("the former scaffold URL should redirect and retain its section anchor")
	}
}
