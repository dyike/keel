package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNavigationCategories(t *testing.T) {
	s := &site{root: "../..", out: t.TempDir(), repo: "https://github.com/dyike/keel", branch: "main"}
	if err := s.build(""); err != nil {
		t.Fatal(err)
	}
	var titles []string
	var catalog navGroup
	for _, group := range s.nav {
		titles = append(titles, group.Title)
		if group.Title == "组件参考" {
			catalog = group
		}
	}
	if want := []string{"开始使用", "编写应用", "组件参考", "测试与调试", "参与开发", "源码导览"}; !slices.Equal(titles, want) {
		t.Fatalf("top-level navigation = %v, want %v", titles, want)
	}
	if len(catalog.Children) != 8 {
		t.Fatalf("component categories = %d, want 8", len(catalog.Children))
	}
	for src, category := range map[string]string{
		"docs/kit/input.md":   "输入与选择",
		"docs/kit/table.md":   "列表与表格",
		"docs/kit/chart.md":   "图表与绘图",
		"docs/kit/dock.md":    "应用布局",
		"docs/kit/message.md": "聊天与内容",
		"docs/kit/dialog.md":  "浮层与反馈",
		"docs/kit/avatar.md":  "基础展示",
		"docs/kit/button.md":  "操作与导航",
	} {
		p := s.pages[src]
		if p == nil {
			t.Fatalf("%s: missing component page", src)
		}
		if p.Group != "组件 / "+category {
			t.Errorf("%s: wrong search category", src)
		}
		if !navigationView(p, catalog).Open {
			t.Errorf("%s: parent category does not open", src)
		}
		for _, child := range catalog.Children {
			if got := navigationView(p, child).Open; got != (child.Title == category) {
				t.Errorf("%s: %s open = %v", src, child.Title, got)
			}
		}
	}
	if navigationView(s.pages["docs/getting-started.md"], catalog).Open {
		t.Error("the component catalog should be collapsed on the getting-started page")
	}
	seen := map[string]int{}
	var visit func([]navGroup)
	visit = func(groups []navGroup) {
		for _, group := range groups {
			for _, p := range group.Pages {
				seen[p.Src]++
			}
			visit(group.Children)
		}
	}
	visit(s.nav)
	for src := range s.pages {
		if src != "README.md" && seen[src] != 1 {
			t.Errorf("%s: %d navigation entries, want exactly one", src, seen[src])
		}
	}
}

func TestComponentIndexValidation(t *testing.T) {
	for _, tc := range []struct {
		name, index, want string
	}{
		{"uncategorized", "# 组件参考\n", "docs/kit/button.md: add this page"},
		{"duplicate", "# 组件参考\n## 操作\n[Button](kit/button.md)\n## 展示\n[Button](kit/button.md)\n", "listed in more than one place"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
				t.Fatal(err)
			}
			for name, data := range map[string]string{
				"README.md": "# 文档\n## 开始使用\n[组件参考](kit.md)\n",
				"kit.md":    tc.index,
			} {
				if err := os.WriteFile(filepath.Join(root, "docs", name), []byte(data), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			s := &site{root: root, pages: map[string]*page{}}
			for _, src := range []string{"README.md", "docs/README.md", "docs/kit.md", "docs/kit/button.md"} {
				p := &page{Src: src, Out: outPath(src), Group: groupOf(src)}
				s.pages[src] = p
				s.order = append(s.order, p)
			}
			if err := s.buildNav(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("buildNav() = %v, want error containing %q", err, tc.want)
			}
		})
	}
}

func TestPageNavigation(t *testing.T) {
	s := &site{root: "../..", out: t.TempDir(), repo: "https://github.com/dyike/keel", branch: "main"}
	if err := s.build(""); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ src, previous, next string }{
		{"README.md", "", ""},
		{"docs/README.md", "", "docs/getting-started.md"},
		{"docs/getting-started.md", "docs/README.md", "docs/web.md"},
		{"docs/kit.md", "docs/base.md", "docs/kit/label.md"},
		{"docs/kit/label.md", "docs/kit.md", "docs/kit/icon.md"},
		{"docs/kit/command.md", "docs/kit/stepper.md", "docs/kit/input.md"},
		{"docs/kit/input.md", "docs/kit/command.md", "docs/kit/input_group.md"},
		{"docs/kit/carousel.md", "docs/kit/settings.md", "docs/automation.md"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			p := s.pages[tc.src]
			data, err := os.ReadFile(filepath.Join(s.out, p.Out))
			if err != nil {
				t.Fatal(err)
			}
			for _, neighbor := range []struct {
				page     *page
				src, rel string
			}{{p.Previous, tc.previous, "prev"}, {p.Next, tc.next, "next"}} {
				if neighbor.src == "" {
					if neighbor.page != nil || strings.Contains(string(data), `rel="`+neighbor.rel+`"`) {
						t.Errorf("unexpected %s link", neighbor.rel)
					}
					continue
				}
				if neighbor.page == nil || neighbor.page.Src != neighbor.src {
					t.Fatalf("%s does not point to %s", neighbor.rel, neighbor.src)
				}
				link := `href="` + relURL(p.Out, neighbor.page.Out) + `" rel="` + neighbor.rel + `"`
				if !strings.Contains(string(data), link) {
					t.Errorf("missing rendered %s", link)
				}
			}
		})
	}
	// Every published document must be reachable exactly once, and links
	// must work in both directions without looping back to the home page.
	seen := map[string]bool{}
	var previous *page
	for p := s.pages["docs/README.md"]; p != nil; p = p.Next {
		if seen[p.Src] {
			t.Fatalf("reading order loops at %s", p.Src)
		}
		seen[p.Src] = true
		if p.Previous != previous {
			t.Fatalf("%s: previous link does not return to the prior page", p.Src)
		}
		previous = p
	}
	if seen["README.md"] || len(seen) != len(s.pages)-1 {
		t.Fatalf("reading order covers %d documents, want %d and no home page", len(seen), len(s.pages)-1)
	}
	data, err := os.ReadFile(filepath.Join(s.out, previous.Out))
	if err != nil || strings.Contains(string(data), `rel="next"`) {
		t.Fatal("the last document should not render a next link")
	}
}
