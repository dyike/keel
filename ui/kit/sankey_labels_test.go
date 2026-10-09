package kit

import (
	"fmt"
	"gioui.org/io/input"
	"github.com/dyike/keel/ui/el"
	"image"
	"strings"
	"testing"
)

func TestSankeyDenseLabelsMeasuredBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, width := range []int{180, 600} {
			nodes := make([]SankeyNode, 30)
			links := []SankeyLink{}
			for i := range nodes {
				nodes[i].Name = fmt.Sprintf("node-%02d", i)
				if i > 0 {
					links = append(links, SankeyLink{Source: 0, Target: i, Value: float64(i)})
				}
			}
			v := SankeyChart(nodes, links).Height(180).Labels(func(_ *el.Context, d SankeyTooltip) el.Element {
				return el.Div().H(el.Dp(36)).Role("note").Name("label:" + d.Node.Name).Child(el.Text(d.Node.Name))
			})
			h := renderView(v, width, scale)
			h.Frame()
			labels := []image.Rectangle{}
			seen := map[string]bool{}
			var walk func(input.SemanticNode)
			walk = func(n input.SemanticNode) {
				if strings.HasPrefix(n.Desc.Label, "label:") {
					if !seen[n.Desc.Label] {
						labels = append(labels, n.Desc.Bounds)
						seen[n.Desc.Label] = true
					}
				} else if strings.HasPrefix(n.Desc.Label, "node-") && n.Desc.Description != "" {
					seen[n.Desc.Label] = true
				}
				for _, c := range n.Children {
					walk(c)
				}
			}
			for _, n := range h.Router.AppendSemantics(nil) {
				walk(n)
			}
			if len(labels) == 0 || len(labels) >= len(nodes) {
				t.Fatal("dense labels not thinned", scale, width, len(labels))
			}
			for i, a := range labels {
				for _, b := range labels[i+1:] {
					if a.Overlaps(b) {
						t.Fatal("custom label overlap", a, b)
					}
				}
			}
			// Every original node is still represented in the figure even when its
			// text label is suppressed. Data and hover are unchanged.
			for _, node := range nodes {
				if !shown(h, node.Name) {
					t.Fatal("node semantics lost", node.Name)
				}
			}
			target := -1
			g := v.layout(v.width, v.height)
			for i, r := range g.nodes {
				if i > 0 && r.y+r.h/2 > 20 && r.y+r.h/2 < v.height-25 && !shown(h, "label:"+nodes[i].Name) {
					target = i
					break
				}
			}
			if target < 0 {
				t.Fatal("no suppressed interior label to exercise")
			}
			v.hover = target
			h.Frame()
			if !shown(h, "label:"+nodes[target].Name) {
				t.Fatal("hovered node label lost priority", scale, width)
			}
		}
	}
}

func TestSankeyLabelPlacementBoundsAndReset(t *testing.T) {
	var p sankeyLabelPlacement
	p.reset(image.Rect(10, 20, 210, 120), 2)
	if !p.take(image.Rect(20, 30, 80, 60)) || p.take(image.Rect(81, 30, 100, 60)) || p.take(image.Rect(0, 0, 30, 40)) {
		t.Fatal("collision or clipping accepted")
	}
	if !p.take(image.Rect(82, 30, 100, 60)) {
		t.Fatal("separated label rejected")
	}
	p.reset(image.Rect(10, 20, 210, 120), 2)
	if !p.take(image.Rect(20, 30, 80, 60)) {
		t.Fatal("previous frame retained")
	}
}
