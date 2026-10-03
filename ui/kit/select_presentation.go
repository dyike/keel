package kit

import "github.com/dyike/keel/ui/el"

// SelectItemContext describes a visible option. Index addresses Entries, not
// the filtered/grouped list. Custom content should be presentational; the row
// retains selection, keyboard behavior and accessible option semantics.
type SelectItemContext struct {
	Option           SelectOption
	Index            int
	Selected, Active bool
}

func (v *SelectView) RenderItem(fn func(*el.Context, SelectItemContext) el.Element) *SelectView {
	v.renderItem = fn
	return v
}

// RenderValue replaces selected text, leaving the empty hint unchanged. The
// supplied entries are an owned snapshot; nil output uses the ordinary labels.
func (v *SelectView) RenderValue(fn func(*el.Context, []SelectOption) el.Element) *SelectView {
	v.renderValue = fn
	return v
}
func (v *SelectView) Empty(view el.View) *SelectView        { v.empty = view; return v }
func (v *SelectView) TitlePrefix(prefix string) *SelectView { v.titlePrefix = prefix; return v }
func (v *SelectView) Clearable(on bool) *SelectView         { v.clearable = on; return v }

// Match replaces the local search predicate. Nil restores label/value matching.
// Reapply it after changing captured search data to invalidate cached results.
func (v *SelectView) Match(fn func(SelectOption, string) bool) *SelectView {
	v.match = fn
	v.cached = false
	return v
}

// MenuWidth sets popup width in dp; zero restores anchor width. The overlay
// still constrains it to the window. Negative/non-finite values are ignored.
func (v *SelectView) MenuWidth(dp float32) *SelectView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.menuWidth = dp
	}
	return v
}

// MenuMaxHeight includes search and padding. Positive values have a 64dp floor;
// zero restores the 240dp default. Window bounds always take precedence.
func (v *SelectView) MenuMaxHeight(dp float32) *SelectView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.menuHeight = dp
	}
	return v
}

// Size scales field height, spacing, text, navigation icon and default row height.
// Zero restores the theme size. Values below 20dp are clamped to 20dp.
func (v *SelectView) Size(dp float32) *SelectView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.height = dp
		if dp > 0 {
			v.height = max(20, dp)
		}
		v.virtual.reveal = v.active
	}
	return v
}

// RowHeight sets the uniform option/group height; zero follows Size.
func (v *SelectView) RowHeight(dp float32) *SelectView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.rowHeight = dp
		if dp > 0 {
			v.rowHeight = max(16, dp)
		}
		v.virtual.reveal = v.active
	}
	return v
}
func (v *SelectView) Appearance(on bool) *SelectView { v.plain = !on; return v }
func (v *SelectView) clearSelection() {
	if v.disabled || len(v.Values()) == 0 {
		return
	}
	old := v.value
	v.SetValue("")
	v.open = false
	v.query = ""
	v.active = -1
	v.err = ""
	if v.multiple && v.onValues != nil {
		v.onValues(nil)
	}
	if old != "" && v.onChange != nil {
		v.onChange("")
	}
}
