package markdown

import (
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// YAML front matter is a block of metadata at the very start of a document,
// between --- lines, as static site generators and note apps write it:
//
//	---
//	title: Release notes
//	tags: [keel, gio]
//	---
//
// It is not part of the text: by default it is hidden, and FrontMatter and
// Meta read it. Only a closed block counts, so while one is still streaming
// it shows as written.

// splitFrontMatter cuts a leading front matter block from src: the block
// with its delimiter lines, its YAML, and the rest.
func splitFrontMatter(src string) (block, yaml, rest string) {
	first, _, ok := strings.Cut(src, "\n")
	if !ok || strings.TrimRight(first, " \r") != "---" {
		return "", "", src
	}
	off := len(first) + 1
	for off <= len(src) {
		line, _, more := strings.Cut(src[off:], "\n")
		end := off + len(line)
		if more {
			end++
		}
		if t := strings.TrimRight(line, " \r"); t == "---" || t == "..." {
			return src[:end], src[len(first)+1 : off], src[end:]
		}
		if !more {
			break
		}
		off = end
	}
	return "", "", src
}

// FrontMatter returns the YAML of the document's front matter, without the
// --- lines, or "" if it has none.
func (d *Doc) FrontMatter() string {
	_, yaml, _ := splitFrontMatter(d.src)
	return yaml
}

// Meta reads the front matter's top-level "key: value" lines, with quotes
// around a value removed. Nested and list values are kept as written; use
// FrontMatter with a YAML parser for those.
func (d *Doc) Meta() map[string]string {
	return frontMatterPairs(d.FrontMatter())
}

// ShowFrontMatter shows the front matter as a table of its keys at the top
// of the document instead of hiding it.
func (d *Doc) ShowFrontMatter(on bool) *Doc {
	if d.showFrontMatter != on {
		d.showFrontMatter = on
		d.extensionRevision++
	}
	return d
}

func frontMatterPairs(yaml string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(yaml, "\n") {
		if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '#' || line[0] == '-' {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
			v = v[1 : len(v)-1]
		}
		out[strings.TrimSpace(k)] = v
	}
	return out
}

// frontMatterTable draws the metadata, in the order the keys were written.
func (d *Doc) frontMatterTable(b *block) el.Element {
	if !d.showFrontMatter {
		return el.Div().Hidden(true)
	}
	pairs := frontMatterPairs(b.code)
	table := el.Div().Role("table").Name("Front matter").Gap(theme.SpaceXxs).P(theme.SpaceMd).Rounded(theme.RadiusMd).
		Border(1, theme.Border).Bg(theme.Subtle).TextSize(theme.TextSm)
	for _, line := range strings.Split(b.code, "\n") {
		k, _, ok := strings.Cut(line, ":")
		k = strings.TrimSpace(k)
		v, known := pairs[k]
		if !ok || !known || line[0] == ' ' {
			continue
		}
		table.Child(el.Div().Row().Gap(theme.SpaceMd).Child(
			el.Text(k).W(el.Dp(120)).NoShrink().TextColor(theme.Muted).MaxLines(1),
			el.Text(v).Grow().W(el.Dp(0)),
		))
	}
	return table
}
