package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
)

// Inline places the palette in normal layout and opens it when enabled. Running
// an inline command leaves it visible. SetValue still controls its visibility.
// Switching presentation invalidates outstanding search requests.
func (v *CommandView) Inline(on bool) *CommandView {
	if v.inline != on {
		v.inline = on
		v.SetValue(on)
	}
	return v
}

// Searchable toggles the search field and local filtering. Without search all
// items are shown and OnSearch is not invoked. Changing mode clears the query.
func (v *CommandView) Searchable(on bool) *CommandView {
	if v.nonsearchable == !on {
		return v
	}
	v.nonsearchable = !on
	v.query = ""
	v.cached = false
	v.request++
	v.loading = false
	v.searchError = ""
	v.active = -1
	if v.open {
		v.searchChanged()
		v.pendingFocus = !v.inline
	}
	return v
}

// Header adds content above the optional search field. Nil removes it.
func (v *CommandView) Header(content el.View) *CommandView { v.header = content; return v }

// Footer adds content below every result state. Nil removes it.
func (v *CommandView) Footer(content el.View) *CommandView { v.footer = content; return v }

// Empty replaces the no-matches content. Nil restores the default message.
func (v *CommandView) Empty(content el.View) *CommandView { v.empty = content; return v }

// RenderItem replaces a visible command's presentation, including its shortcut.
// Nil content uses the default row. The outer row retains selection and disabled
// semantics. Reuse stateful children; nested actions do not execute the command.
func (v *CommandView) RenderItem(fn func(CommandItem, bool) el.View) *CommandView {
	v.renderItem = fn
	return v
}

// RowHeight sets a uniform virtual slot height in dp. Zero restores 36dp;
// negative or non-finite values are ignored. Group headings share this height.
func (v *CommandView) RowHeight(dp float32) *CommandView {
	if dp < 0 || math.IsNaN(float64(dp)) || math.IsInf(float64(dp), 0) {
		return v
	}
	v.rowHeight = dp
	v.list.rowH = v.itemHeight()
	if v.active >= 0 {
		v.list.reveal = v.active
	}
	return v
}
func (v *CommandView) itemHeight() float32 {
	if v.rowHeight == 0 {
		return 36
	}
	return max(5, v.rowHeight)
}

// FocusID is the current search input or non-searchable frame's focus target.
func (v *CommandView) FocusID() string {
	id := autoID("command", v)
	if v.nonsearchable {
		return id + "/panel"
	}
	return id + "/search"
}

// Focus explicitly focuses an open palette. Inline palettes never steal focus
// merely because they render.
func (v *CommandView) Focus(cx *el.Context) {
	if v.open && !v.disabled {
		cx.Focus(v.FocusID())
	}
}
func (v *CommandView) confirmActive() {
	if v.active >= 0 && v.active < len(v.rows) && !v.rowDisabled(v.active) {
		v.run(v.rows[v.active])
	}
}
