package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type matrixImage struct{ Name, File string }

// Each case builds fresh state so themes and viewports cannot inherit layout.
func renderMatrix(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	old := theme.Current()
	defer theme.Apply(old)
	sort.Slice(demoSections, func(i, j int) bool { return demoSections[i].name < demoSections[j].name })
	var images []matrixImage
	for _, section := range demoSections {
		for _, mode := range []string{"light", "dark"} {
			for _, width := range []int{320, 680} {
				for _, scale := range []int{1, 2} {
					if mode == "light" {
						theme.Apply(theme.Light())
					} else {
						theme.Apply(theme.Dark())
					}
					name := fmt.Sprintf("%s-%s-%d-%dx", section.name, mode, width, scale)
					file := name + ".png"
					if err := window.ScreenshotAtScale(section.build(), width, 1040, float32(scale), filepath.Join(dir, file)); err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}
					images = append(images, matrixImage{Name: name, File: file})
				}
			}
		}
		fmt.Println("rendered", section.name)
	}
	f, err := os.Create(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	err = matrixHTML.Execute(f, images)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

var matrixHTML = template.Must(template.New("matrix").Parse(`<!doctype html><meta charset="utf-8"><title>Keel component matrix</title>
<style>body{font:14px system-ui;background:#ddd}main{display:flex;flex-wrap:wrap;gap:16px}figure{margin:0}img{width:320px;height:auto;display:block}figcaption{padding:8px 0}</style>
<h1>Keel component matrix</h1><p>First-frame screenshots; inspect interactions separately. Each image links to its full resolution.</p><main>
{{range .}}<figure><figcaption>{{.Name}}</figcaption><a href="{{.File}}"><img loading="lazy" src="{{.File}}" alt="{{.Name}}"></a></figure>{{end}}</main>`))
