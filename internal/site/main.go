// Command site builds Keel's documentation website: every Markdown file in
// docs/ and every module and example README becomes a static HTML page, the
// component pages embed the live component gallery (the WebAssembly build of
// examples/components), and a search index covers all of it.
//
//	go run ./internal/site -out _site -demo /path/to/gogio/output
//
// The output is plain static files with relative links, so it works from a
// GitHub Pages project URL, a custom domain or a local file server alike.
package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/dyike/keel/internal/svgicon"
)

//go:embed assets
var assets embed.FS

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "_site", "output directory")
	repo := flag.String("repo", "https://github.com/dyike/keel", "GitHub URL for links to source files")
	branch := flag.String("branch", "main", "branch for links to source files")
	demo := flag.String("demo", "", "gogio -target js output of examples/components to publish as demo/")
	version := flag.String("version", "", "release shown in the header, such as v0.0.1; empty shows none")
	flag.Parse()
	s := &site{root: *root, out: *out, repo: strings.TrimSuffix(*repo, "/"), branch: *branch, version: *version}
	if err := s.build(*demo); err != nil {
		log.Fatal(err)
	}
	for _, w := range s.warnings {
		log.Print(w)
	}
}

type site struct {
	root, out, repo, branch string
	version                 string           // release tag, or ""
	pages                   map[string]*page // by repository path of the .md file
	order                   []*page
	sections                map[string]bool // gallery sections, by name
	nav                     []navGroup
	warnings                []string
	copies                  map[string]bool // local images to copy, by repository path
}

type page struct {
	Src, Out string // repository paths: the .md file and the .html file
	Title    string
	Group    string
	Section  string // gallery section to embed, or ""
	HTML     template.HTML
	TOC      []heading
	Text     string // plain text, for search
}

type heading struct {
	Level int
	ID    string
	Text  string
}

type navGroup struct {
	Title string
	Pages []*page
}

// excluded directories hold working notes and dependencies, not documentation.
var excluded = []string{".git", "work", "node_modules", "_site", "testdata"}

// unpublished paths stay in the repository but not on the site: progress
// reports are working records, not documentation.
var unpublished = []string{"docs/reports/"}

func isUnpublished(rel string) bool {
	for _, p := range unpublished {
		if strings.HasPrefix(rel, p) {
			return true
		}
	}
	return false
}

func (s *site) build(demo string) error {
	s.pages, s.sections, s.copies = map[string]*page{}, map[string]bool{}, map[string]bool{}
	if err := s.collect(); err != nil {
		return err
	}
	if err := s.findSections(); err != nil {
		return err
	}
	for _, p := range s.order {
		if err := s.render(p); err != nil {
			return fmt.Errorf("%s: %w", p.Src, err)
		}
	}
	s.buildNav()
	if err := os.RemoveAll(s.out); err != nil {
		return err
	}
	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"rel": func(from *page, to string) string { return relURL(from.Out, to) },
	}).ParseFS(assets, "assets/layout.html")
	if err != nil {
		return err
	}
	// Stylesheets and scripts are linked with a hash of their contents, so
	// a deploy never pairs new pages with cached old ones (Pages caches
	// them for ten minutes).
	files := map[string][]byte{"code.css": codeCSS()}
	for _, name := range []string{"site.css", "site.js"} {
		files[name], _ = assets.ReadFile("assets/" + name)
	}
	asset := map[string]string{}
	for name, data := range files {
		sum := sha256.Sum256(data)
		asset[name] = "assets/" + name + "?v=" + hex.EncodeToString(sum[:5])
	}
	for _, p := range s.order {
		var b bytes.Buffer
		if err := tmpl.Execute(&b, map[string]any{"Page": p, "Nav": s.nav, "Repo": s.repo, "Home": p.Out == "index.html", "Version": s.version, "Asset": asset,
			"Root": relURL(p.Out, "."), "Demo": relURL(p.Out, "demo/index.html"), "Source": s.repo + "/blob/" + s.branch + "/" + p.Src}); err != nil {
			return fmt.Errorf("%s: %w", p.Src, err)
		}
		if err := s.write(p.Out, b.Bytes()); err != nil {
			return err
		}
	}
	for name, data := range files {
		if err := s.write("assets/"+name, data); err != nil {
			return err
		}
	}
	if err := s.writeSearch(); err != nil {
		return err
	}
	for src := range s.copies {
		data, err := os.ReadFile(filepath.Join(s.root, src))
		if err != nil {
			return err
		}
		if err := s.write(src, data); err != nil {
			return err
		}
	}
	if err := s.writeIcons(); err != nil {
		return err
	}
	// GitHub Pages would otherwise run Jekyll and drop files it dislikes.
	if err := s.write(".nojekyll", nil); err != nil {
		return err
	}
	if demo != "" {
		return s.copyDemo(demo)
	}
	return nil
}

// collect finds the Markdown files to publish.
func (s *site) collect() error {
	return filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(s.root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if slices.Contains(excluded, d.Name()) && rel != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(rel, ".md") {
			return nil
		}
		if !strings.HasPrefix(rel, "docs/") && path.Base(rel) != "README.md" || isUnpublished(rel) {
			return nil
		}
		pg := &page{Src: rel, Out: outPath(rel), Group: groupOf(rel)}
		s.pages[rel] = pg
		s.order = append(s.order, pg)
		return nil
	})
}

// outPath maps a Markdown path to its page: README.md is the directory's index.
func outPath(src string) string {
	if path.Base(src) == "README.md" {
		return path.Join(path.Dir(src), "index.html")
	}
	return strings.TrimSuffix(src, ".md") + ".html"
}

func groupOf(src string) string {
	switch {
	case src == "README.md":
		return ""
	case strings.HasPrefix(src, "docs/kit/"):
		return "组件"
	case strings.HasPrefix(src, "docs/"):
		return "指南"
	case strings.HasPrefix(src, "examples/"):
		return "示例"
	}
	return "模块"
}

var sectionRE = regexp.MustCompile(`registerSection\("([a-z_]+)"`)

// findSections reads which components the gallery can show on its own, so
// their pages embed it.
func (s *site) findSections() error {
	files, err := filepath.Glob(filepath.Join(s.root, "examples/components/*.go"))
	if err != nil {
		return err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		for _, m := range sectionRE.FindAllSubmatch(data, -1) {
			s.sections[string(m[1])] = true
		}
	}
	for _, p := range s.order {
		if name, ok := strings.CutPrefix(p.Src, "docs/kit/"); ok {
			if name = strings.TrimSuffix(name, ".md"); s.sections[name] {
				p.Section = name
			}
		}
	}
	return nil
}

var linkRE = regexp.MustCompile(`\]\(([^)#\s]+\.md)(#[^)]*)?\)`)

// buildNav orders the guide as docs/README.md lists it, and everything else
// by title.
func (s *site) buildNav() {
	var guide []*page
	if readme, err := os.ReadFile(filepath.Join(s.root, "docs/README.md")); err == nil {
		for _, m := range linkRE.FindAllStringSubmatch(string(readme), -1) {
			if p := s.pages[path.Join("docs", m[1])]; p != nil && p.Group == "指南" && !slices.Contains(guide, p) {
				guide = append(guide, p)
			}
		}
	}
	if p := s.pages["docs/README.md"]; p != nil {
		guide = append([]*page{p}, guide...)
	}
	groups := map[string][]*page{}
	for _, p := range s.order {
		if p.Group == "指南" && slices.Contains(guide, p) {
			continue
		}
		groups[p.Group] = append(groups[p.Group], p)
	}
	for _, g := range groups {
		sort.Slice(g, func(i, j int) bool { return strings.ToLower(g[i].Title) < strings.ToLower(g[j].Title) })
	}
	guide = append(guide, groups["指南"]...)
	s.nav = []navGroup{{"指南", guide}, {"组件", groups["组件"]}, {"模块", groups["模块"]}, {"示例", groups["示例"]}}
}

func (s *site) write(rel string, data []byte) error {
	p := filepath.Join(s.out, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (s *site) writeSearch() error {
	type entry struct {
		Title    string      `json:"t"`
		URL      string      `json:"u"`
		Group    string      `json:"g"`
		Headings [][2]string `json:"h"` // text and anchor
		Text     string      `json:"x"`
	}
	var index []entry
	for _, p := range s.order {
		e := entry{Title: p.Title, URL: p.Out, Group: p.Group}
		for _, h := range p.TOC {
			e.Headings = append(e.Headings, [2]string{h.Text, h.ID})
		}
		e.Text = p.Text
		if r := []rune(e.Text); len(r) > 4000 {
			e.Text = string(r[:4000])
		}
		index = append(index, e)
	}
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}
	return s.write("search.json", data)
}

// relURL is the relative link from page from to target, both repository
// paths of output files; "." is the site root.
func relURL(from, target string) string {
	dir := strings.Split(path.Dir(from), "/")
	if dir[0] == "." {
		dir = nil
	}
	var to []string
	if target != "." {
		to = strings.Split(target, "/")
	}
	// Drop the directories both paths share, but never the target's file.
	for len(dir) > 0 && len(to) > 1 && dir[0] == to[0] {
		dir, to = dir[1:], to[1:]
	}
	rel := strings.Repeat("../", len(dir)) + strings.Join(to, "/")
	if rel == "" {
		return "./"
	}
	return rel
}

// iconSource is the one source of the Keel icon.
const iconSource = "docs/images/keel.svg"

// writeIcons publishes the icon as the favicon, with PNGs for browsers and
// home screens that want bitmaps.
func (s *site) writeIcons() error {
	svg, err := os.ReadFile(filepath.Join(s.root, iconSource))
	if err != nil {
		return err
	}
	for name, data := range map[string][]byte{"favicon.svg": svg, iconSource: svg} {
		if err := s.write(name, data); err != nil {
			return err
		}
	}
	for name, size := range map[string]int{"favicon-32.png": 32, "apple-touch-icon.png": 180, "icon-512.png": 512} {
		data, err := svgicon.Render(svg, size)
		if err != nil {
			return err
		}
		if err := s.write(name, data); err != nil {
			return err
		}
	}
	return nil
}
