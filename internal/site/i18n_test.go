package main

import (
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestBilingualSite(t *testing.T) {
	out := t.TempDir()
	s := &site{root: "../..", out: out, repo: "https://github.com/dyike/keel", branch: "main"}
	if err := s.build(""); err != nil {
		t.Fatal(err)
	}
	chinese := &site{root: s.root, out: filepath.Join(out, "zh-CN"), lang: "zh-CN", repo: s.repo, branch: s.branch}
	if err := chinese.build(""); err != nil {
		t.Fatal(err)
	}
	if len(s.pages) != len(chinese.pages) {
		t.Fatal("languages have different coverage")
	}
	for _, local := range []*site{s, chinese} {
		data, err := os.ReadFile(filepath.Join(local.out, "search.json"))
		if err != nil {
			t.Fatal(err)
		}
		var entries []struct {
			Title string `json:"t"`
			URL   string `json:"u"`
		}
		if err := json.Unmarshal(data, &entries); err != nil {
			t.Fatal(err)
		}
		if len(entries) != len(local.pages) {
			t.Fatal("search does not cover the language's pages")
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.URL, "zh-CN/") {
				t.Fatal("search entries must be relative to their own language root")
			}
		}
		var titles []string
		for _, g := range local.nav {
			titles = append(titles, g.Title)
		}
		if local.lang == "zh-CN" && !slices.Equal(titles, []string{"开始使用", "编写应用", "组件参考", "测试与调试", "参与开发", "源码导览"}) {
			t.Fatalf("Chinese navigation: %v", titles)
		}
		seen := map[string]bool{}
		for p := local.pages["docs/README.md"]; p != nil; p = p.Next {
			if p.Lang != local.lang || seen[p.Src] {
				t.Fatal("reading order crosses languages or loops")
			}
			seen[p.Src] = true
		}
		if len(seen) != len(local.pages)-1 {
			t.Fatal("reading order does not cover all documents")
		}
		for canonical, p := range local.pages {
			other := s.pages[canonical]
			otherRoot := s.out
			if local.lang == "en" {
				other = chinese.pages[canonical]
				otherRoot = chinese.out
			}
			data, err := os.ReadFile(filepath.Join(local.out, p.Out))
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if !strings.Contains(text, `<html lang="`+local.lang+`">`) {
				t.Errorf("%s: wrong HTML language", p.Src)
			}
			if strings.Contains(text, "English |") {
				t.Errorf("%s: repository language link leaked into article", p.Src)
			}
			doc, err := html.Parse(strings.NewReader(text))
			if err != nil {
				t.Fatal(err)
			}
			found := false
			var visit func(*html.Node)
			visit = func(n *html.Node) {
				attrs := map[string]string{}
				for _, a := range n.Attr {
					attrs[a.Key] = a.Val
				}
				if n.Data == "a" && attrs["class"] == "language" {
					found = true
					u, err := url.Parse(attrs["href"])
					if err != nil {
						t.Fatal(err)
					}
					target := filepath.Clean(filepath.Join(local.out, filepath.Dir(p.Out), u.Path))
					if target != filepath.Join(otherRoot, other.Out) {
						t.Errorf("%s: switch points to %s", p.Src, target)
					}
					var pairs map[string]string
					if err := json.Unmarshal([]byte(attrs["data-anchors"]), &pairs); err != nil {
						t.Fatal(err)
					}
					for i, id := range local.headingIDs(p.Src) {
						want := local.headingIDs(other.Src)[i]
						if pairs[id] != want {
							t.Errorf("%s: section %s does not map to %s", p.Src, id, want)
						}
					}
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					visit(c)
				}
			}
			visit(doc)
			if !found {
				t.Errorf("%s: missing language switch", p.Src)
			}
			for _, neighbor := range []*page{p.Previous, p.Next} {
				if neighbor != nil && neighbor.Lang != local.lang {
					t.Fatal("pagination crosses languages")
				}
			}
		}
	}
	chineseHome, err := os.ReadFile(filepath.Join(out, "zh-CN/index.html"))
	if err != nil || !strings.Contains(string(chineseHome), "/blob/main/LICENSING.zh-CN.md") || !strings.Contains(string(chineseHome), "/blob/main/CONTRIBUTING.zh-CN.md") {
		t.Fatal("Chinese repository links must point to Chinese documents")
	}
	// Old bookmarks stay valid even though the default page is now English.
	page, err := os.ReadFile(filepath.Join(out, "docs/getting-started.html"))
	if err != nil || !strings.Contains(string(page), `id="写第一个窗口"`) || !strings.Contains(string(page), `id="write-your-first-window"`) {
		t.Fatal("legacy and English anchors must coexist")
	}
}

func TestDocumentationTranslationCoverage(t *testing.T) {
	count := 0
	err := filepath.WalkDir("../..", func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if slices.Contains(excluded, d.Name()) && file != "../.." {
				return filepath.SkipDir
			}
			return nil
		}
		src, err := filepath.Rel("../..", file)
		if err != nil {
			return err
		}
		src = filepath.ToSlash(src)
		if !strings.HasSuffix(src, ".md") || sourceLanguage(src) == "zh-CN" || strings.HasPrefix(src, "examples/chat/samples/") {
			return nil
		}
		counterpart := localizedSource(src, "zh-CN")
		if _, err := os.Stat(filepath.Join("../..", counterpart)); err != nil {
			t.Errorf("%s: missing %s", src, counterpart)
			return nil
		}
		s := &site{root: "../.."}
		en, zh := s.headingIDs(src), s.headingIDs(counterpart)
		if len(en) != len(zh) {
			t.Errorf("%s: headings differ (%d English, %d Chinese)", src, len(en), len(zh))
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count < 142 {
		t.Fatalf("only %d translated documents", count)
	}
}
