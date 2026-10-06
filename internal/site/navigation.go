package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// The Markdown indexes own both the reading order and the sidebar categories.
// Module READMEs stay accessible under the source guide, apart from app usage.
func (s *site) buildNav() error {
	seen := map[string]bool{"README.md": true, "docs/README.md": true}
	groups, err := s.indexGroups("docs/README.md", "指南", seen)
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("docs/README.md: no documentation sections")
	}
	if p := s.pages["docs/README.md"]; p != nil {
		p.NavTitle = translateUI(p, "文档导览", "Overview")
		p.Group = groups[0].Title
		groups[0].Pages = append([]*page{p}, groups[0].Pages...)
	}
	components, err := s.indexGroups("docs/kit.md", "组件", seen)
	if err != nil {
		return err
	}
	for i := range groups {
		for _, p := range groups[i].Pages {
			if canonicalSource(p.Src) == "docs/kit.md" {
				p.NavTitle = translateUI(p, "全部组件", "Overview")
				groups[i].Children = components
			}
		}
	}

	source := map[string][]*page{}
	for _, p := range s.order {
		if seen[canonicalSource(p.Src)] {
			continue
		}
		if strings.HasPrefix(p.Src, "docs/") {
			return fmt.Errorf("%s: add this page to docs/README.md or a category in docs/kit.md", p.Src)
		}
		group := "工具与内部实现"
		switch {
		case strings.HasPrefix(p.Src, "ui/internal/"):
		case strings.HasPrefix(p.Src, "ui/"):
			group = "界面模块"
		case strings.HasPrefix(p.Src, "native/"):
			group = "系统能力"
		case strings.HasPrefix(p.Src, "cmd/"):
			group = "命令行工具"
		case strings.HasPrefix(p.Src, "examples/"):
			group = "示例"
		}
		p.Group = translateUI(p, "源码导览", "Packages") + " / " + localizedNavTitle(s.lang, group)
		source[group] = append(source[group], p)
	}
	s.nav = groups
	var children []navGroup
	for _, title := range []string{"示例", "界面模块", "系统能力", "命令行工具", "工具与内部实现"} {
		pages := source[title]
		if len(pages) == 0 {
			continue
		}
		sort.Slice(pages, func(i, j int) bool { return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title) })
		children = append(children, navGroup{Title: localizedNavTitle(s.lang, title), Pages: pages})
	}
	if len(children) > 0 {
		s.nav = append(s.nav, navGroup{Title: localizedNavTitle(s.lang, "源码导览"), Children: children})
	}
	return nil
}

// Only second-level headings with page links in their section become groups.
// Parse the Markdown tree so code blocks and inline API examples are ignored.
func (s *site) indexGroups(src, kind string, seen map[string]bool) ([]navGroup, error) {
	data, err := os.ReadFile(filepath.Join(s.root, localizedSource(src, s.lang)))
	if err != nil {
		return nil, err
	}
	doc := md.Parser().Parse(text.NewReader(data))
	var groups []navGroup
	var current *navGroup
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			if n.Level == 2 {
				groups = append(groups, navGroup{Title: nodeText(n, data)})
				current = &groups[len(groups)-1]
			}
		case *ast.Link:
			if current == nil {
				break
			}
			u, err := url.Parse(string(n.Destination))
			if err != nil || u.IsAbs() || u.Host != "" || u.Path == "" {
				break
			}
			target := canonicalSource(path.Clean(path.Join(path.Dir(src), u.Path)))
			p := s.pages[target]
			if p == nil || groupOf(target) != kind {
				break
			}
			if seen[target] {
				if kind == "组件" {
					return ast.WalkStop, fmt.Errorf("%s: component %s is listed in more than one place", src, target)
				}
				break
			}
			seen[target] = true
			p.Group = current.Title
			if kind == "组件" {
				p.Group = translateUI(p, "组件", "Components") + " / " + current.Title
			}
			current.Pages = append(current.Pages, p)
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	var nonempty []navGroup
	for _, group := range groups {
		if len(group.Pages) > 0 {
			nonempty = append(nonempty, group)
		}
	}
	return nonempty, nil
}

type navView struct {
	Page  *page
	Group navGroup
	Open  bool
}

// Follow the sidebar's reading order, including nested component categories.
// The home page is not part of the documentation sequence.
func (s *site) linkReadingOrder() {
	var previous *page
	var visit func([]navGroup)
	visit = func(groups []navGroup) {
		for _, group := range groups {
			for _, p := range group.Pages {
				p.Previous, p.Next = previous, nil
				if previous != nil {
					previous.Next = p
				}
				previous = p
			}
			visit(group.Children)
		}
	}
	visit(s.nav)
}

func navigationView(p *page, group navGroup) navView {
	return navView{Page: p, Group: group, Open: (group.Title == "开始使用" || group.Title == "Getting started") || group.contains(p)}
}

func (g navGroup) contains(p *page) bool {
	for _, child := range g.Pages {
		if child.Src == p.Src {
			return true
		}
	}
	for _, child := range g.Children {
		if child.contains(p) {
			return true
		}
	}
	return false
}
