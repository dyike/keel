package markdown

import (
	"image"
	"reflect"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	east "github.com/yuin/goldmark/extension/ast"
)

const footnotePrefix = "#keel-footnote-"

func footnoteID(index int) string         { return footnotePrefix + strconv.Itoa(index) }
func footnoteRefID(index, ref int) string { return footnoteID(index) + "-ref-" + strconv.Itoa(ref) }
func backlinkSpan(n *east.FootnoteBacklink) span {
	label := "↩"
	if n.RefCount > 1 {
		label += strconv.Itoa(n.RefIndex + 1)
	}
	return span{text: " " + label, link: footnoteRefID(n.Index, n.RefIndex)}
}
func anchorFirst(bs []block, id string) bool {
	for i := range bs {
		b := &bs[i]
		if b.kind == paragraph || b.kind == heading || b.kind == codeBlock {
			b.anchor = id
			return true
		}
		if anchorFirst(b.children, id) {
			return true
		}
		for j := range b.items {
			if anchorFirst(b.items[j].blocks, id) {
				return true
			}
		}
	}
	return false
}

// References and footnotes have document-wide dependencies, including forward
// definitions. Goldmark resolves these in one context. Unchanged converted
// blocks retain their backing slice and view, so layout caches, selection and
// code controls survive this parse. Ordinary documents keep tail-only parsing.
func (d *Doc) updateContextual() {
	src := d.src
	if d.streaming {
		src = heal(src)
	}
	if d.contextual && d.parsedContext == src {
		return
	}
	bs := parse(src)
	d.parses++
	next := make([]chunk, len(bs))
	for i, b := range bs {
		if i < len(d.chunks) && len(d.chunks[i].blocks) == 1 {
			old := d.chunks[i]
			if sameBlock(old.blocks[0], b) {
				next[i] = old
				continue
			}
		}
		next[i] = chunk{blocks: []block{b}}
		if i < len(d.chunks) {
			preserveCodeViews(d.chunks[i].blocks, next[i].blocks)
		}
	}
	d.chunks = next
	d.contextual, d.parsedContext = true, src
}
func sameBlock(a, b block) bool {
	if a.kind != b.kind || a.level != b.level || a.anchor != b.anchor || a.lang != b.lang || a.code != b.code || a.ordered != b.ordered || a.start != b.start {
		return false
	}
	if !reflect.DeepEqual(a.spans, b.spans) || !reflect.DeepEqual(a.tbl, b.tbl) || len(a.children) != len(b.children) || len(a.items) != len(b.items) {
		return false
	}
	for i := range a.children {
		if !sameBlock(a.children[i], b.children[i]) {
			return false
		}
	}
	for i := range a.items {
		x, y := a.items[i], b.items[i]
		if !reflect.DeepEqual(x.task, y.task) || len(x.blocks) != len(y.blocks) {
			return false
		}
		for j := range x.blocks {
			if !sameBlock(x.blocks[j], y.blocks[j]) {
				return false
			}
		}
	}
	return true
}

func (d *Doc) followLink(url string) {
	if strings.HasPrefix(url, footnotePrefix) {
		d.pendingAnchor = url
		return
	}
	if d.onLink != nil {
		d.onLink(url)
	}
}
func (d *Doc) navigate(cx *el.Context, root el.Element, gtx core.C) {
	if d.pendingAnchor == "" {
		return
	}
	anchor := d.pendingAnchor
	found := false
	target := 0
	el.VisitWidgets(root, gtx.Metric, func(w core.Widget, bounds image.Rectangle) {
		if found {
			return
		}
		var r *richBlock
		switch w := w.(type) {
		case *richBlock:
			r = w
		case *codeBody:
			r = w.view.rich
		}
		if r == nil {
			return
		}
		if r.anchor == anchor {
			target = bounds.Min.Y
			found = true
			return
		}
		for i, rn := range r.runs {
			if rn.anchor != anchor {
				continue
			}
			target = bounds.Min.Y
			found = true
			for _, p := range r.rt.pieces {
				if p.run == i {
					target += p.rect.Min.Y
					break
				}
			}
			return
		}
	})
	if found {
		origin, viewport := cx.PaintGeometry()
		cx.ScrollBy(target + origin.Y - viewport.Min.Y - gtx.Dp(8))
		d.pendingAnchor = ""
	}
}
