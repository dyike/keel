package main

import (
	"image"
	"sort"

	"gioui.org/layout"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"
)

type demoSection struct {
	name, category string
	build          func() core.Widget
}

var demoSections []demoSection

func registerSection(name, category string, build func() core.Widget) {
	demoSections = append(demoSections, demoSection{name, category, build})
}
func sectionContent(name string) (core.Widget, bool) {
	sort.Slice(demoSections, func(i, j int) bool { return demoSections[i].name < demoSections[j].name })
	var content []core.Widget
	for _, s := range demoSections {
		if name == "all" || s.name == name || s.category == name {
			content = append(content, &translatedSection{build: s.build})
		}
	}
	// Preserve FillsWindow for a standalone el.Root section.
	if len(content) == 1 {
		return content[0], true
	}
	if len(content) == 0 {
		return nil, false
	}
	return column(content), true
}

// column stacks several sections. A section that fills its window (an
// el.Root) gets a fixed height here, so it does not take all the space.
type column []core.Widget

func (c column) Layout(gtx core.C) core.D {
	children := make([]layout.FlexChild, 0, 2*len(c))
	for i, w := range c {
		w := w
		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: 24}.Layout))
		}
		children = append(children, layout.Rigid(func(gtx core.C) core.D {
			if f, ok := w.(interface{ FillsWindow() bool }); ok && f.FillsWindow() {
				gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, gtx.Dp(560)))
			}
			return w.Layout(gtx)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// Standalone examples also rebuild their application text when a control
// changes the locale (for example, Settings' language selector).
type translatedSection struct {
	build    func() core.Widget
	widget   core.Widget
	revision uint64
}

func (s *translatedSection) current() core.Widget {
	if s.widget == nil || s.revision != locale.Revision() {
		s.widget = s.build()
		s.revision = locale.Revision()
	}
	return s.widget
}

func (s *translatedSection) Layout(gtx core.C) core.D { return s.current().Layout(gtx) }

func (s *translatedSection) FillsWindow() bool {
	f, ok := s.current().(interface{ FillsWindow() bool })
	return ok && f.FillsWindow()
}
