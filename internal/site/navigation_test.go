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
