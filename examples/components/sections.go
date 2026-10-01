package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"sort"
)

type demoSection struct {
	name, category string
	build          func() core.Widget
}

var demoSections []demoSection

func registerSection(name, category string, build func() core.Widget) {
	demoSections = append(demoSections, demoSection{name, category, build})
}
func sectionContent(name string, base core.Widget) (core.Widget, bool) {
	sort.Slice(demoSections, func(i, j int) bool { return demoSections[i].name < demoSections[j].name })
	var content []core.Widget
	for _, s := range demoSections {
		if name == "all" || s.name == name || s.category == name {
			content = append(content, s.build())
		}
	}
	if name == "all" {
		content = append(content, base)
	}
	if len(content) == 0 {
		return nil, false
	}
	return layout.Column(content...), true
}
