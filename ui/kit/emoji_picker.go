package kit

import (
	"slices"
	"strings"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// EmojiSelection reports both the exact output and its unmodified catalog base.
type EmojiSelection struct{ Emoji, Base, Name string }

// EmojiPickerView is a searchable, virtualized emoji panel. Data and recent
// choices stay local; applications own persistence and popup presentation.
type EmojiPickerView struct {
	sheetPanel                bool
	popup, open               bool
	popover                   *PopoverView
	sheet                     *SheetView
	preview                   string
	entries                   []EmojiEntry
	lookup                    map[string]int
	search                    []string
	groups                    []string
	recent                    []string
	matches                   []int
	query, group, value, name string
	tone, focus, columns      int
	cell, height              float32
	disabled, dirty           bool
	input                     *InputView
	list                      *VirtualListView
	root                      *el.RootWidget
	onChange                  func(string)
	onSelect                  func(EmojiSelection)
}

// EmojiPicker uses the complete default catalog when no entries are supplied.
// Data can replace it with a custom or localized catalog, including an empty one.
func EmojiPicker(entries ...EmojiEntry) *EmojiPickerView {
	v := &EmojiPickerView{height: 264, dirty: true, columns: 6}
	v.input = Input("").Clearable().OnChange(func(s string) { v.query = s; v.dirty = true; v.focus = 0 })
	v.list = VirtualList(0, 44, v.renderRow)
	v.root = el.Embed(el.ViewFunc(v.renderPanel))
	if len(entries) == 0 {
		entries = defaultEmojiCatalog()
	}
	v.Data(entries)
	return v
}
func (v *EmojiPickerView) Name(s string) *EmojiPickerView { v.name = s; return v }
func (v *EmojiPickerView) Value() string                  { return v.value }
func (v *EmojiPickerView) SetValue(s string) {
	if s == "" {
		v.value = ""
		return
	}
	if _, ok := v.lookup[s]; ok {
		v.value = s
	}
}
func (v *EmojiPickerView) OnChange(fn func(string)) *EmojiPickerView { v.onChange = fn; return v }

// OnSelect fires for every user choice, including choosing the same emoji twice.
func (v *EmojiPickerView) OnSelect(fn func(EmojiSelection)) *EmojiPickerView {
	v.onSelect = fn
	return v
}
func (v *EmojiPickerView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetOpen(false)
	}
}
func (v *EmojiPickerView) Height(dp float32) *EmojiPickerView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Data copies a catalog. Empty/duplicate base outputs are ignored; unsupported
// values/recents are removed. Assignment is silent.
func (v *EmojiPickerView) Data(entries []EmojiEntry) *EmojiPickerView {
	v.entries = nil
	v.search = nil
	v.groups = nil
	v.lookup = make(map[string]int)
	seenGroups := map[string]bool{}
	for _, e := range cloneEmojiEntries(entries) {
		if e.Emoji == "" {
			continue
		}
		if _, exists := v.lookup[e.Emoji]; exists {
			continue
		}
		if e.Name == "" {
			e.Name = e.Emoji
		}
		i := len(v.entries)
		v.lookup[e.Emoji] = i
		for tone, variant := range e.SkinTones {
			if variant != "" {
				if _, exists := v.lookup[variant]; !exists {
					v.lookup[variant] = i
				} else {
					e.SkinTones[tone] = ""
				}
			}
		}
		v.entries = append(v.entries, e)
		v.search = append(v.search, strings.ToLower(e.Name+" "+strings.Join(e.Keywords, " ")))
		if e.Group != "" && !seenGroups[e.Group] {
			seenGroups[e.Group] = true
			v.groups = append(v.groups, e.Group)
		}
	}
	if _, ok := v.lookup[v.value]; !ok {
		v.value = ""
	}
	if v.group != "" && v.group != emojiRecentGroup && !slices.Contains(v.groups, v.group) {
		v.group = ""
	}
	v.SetRecent(v.recent)
	v.focus = 0
	v.dirty = true
	return v
}
func (v *EmojiPickerView) SetQuery(s string) {
	v.query = s
	v.input.SetValue(s)
	v.focus = 0
	v.dirty = true
}
func (v *EmojiPickerView) Query() string { return v.query }
func (v *EmojiPickerView) SetSkinTone(tone int) {
	if tone >= 0 && tone <= 5 {
		v.tone = tone
	}
}

// SkinTone is zero for unmodified, 1..5 from light to dark.
func (v *EmojiPickerView) SkinTone() int    { return v.tone }
func (v *EmojiPickerView) Groups() []string { return slices.Clone(v.groups) }
func (v *EmojiPickerView) SetGroup(group string) {
	if group == "" || slices.Contains(v.groups, group) {
		v.group = group
		v.focus = 0
		v.dirty = true
	}
}
func (v *EmojiPickerView) ShowRecent()      { v.group = emojiRecentGroup; v.focus = 0; v.dirty = true }
func (v *EmojiPickerView) Recent() []string { return slices.Clone(v.recent) }
func (v *EmojiPickerView) SetRecent(recent []string) {
	out := make([]string, 0, min(18, len(recent)))
	for _, s := range recent {
		if i, ok := v.lookup[s]; ok {
			base := v.entries[i].Emoji
			if !slices.Contains(out, base) {
				out = append(out, base)
			}
			if len(out) == 18 {
				break
			}
		}
	}
	v.recent = out
	if v.group == emojiRecentGroup {
		v.dirty = true
	}
}

const emojiRecentGroup = "\x00recent"

func (v *EmojiPickerView) filter() {
	if !v.dirty {
		return
	}
	v.dirty = false
	v.matches = v.matches[:0]
	words := strings.Fields(strings.ToLower(strings.TrimSpace(v.query)))
	add := func(i int) {
		e := v.entries[i]
		for _, word := range words {
			if strings.Contains(v.search[i], word) || strings.Contains(e.Emoji, word) {
				continue
			}
			exact := false
			for _, s := range e.SkinTones {
				if s != "" && word == s {
					exact = true
					break
				}
			}
			if !exact {
				return
			}
		}
		v.matches = append(v.matches, i)
	}
	if v.group == emojiRecentGroup {
		for _, s := range v.recent {
			if i, ok := v.lookup[s]; ok {
				add(i)
			}
		}
	} else {
		for i, e := range v.entries {
			if v.group == "" || e.Group == v.group {
				add(i)
			}
		}
	}
	v.focus = min(max(0, v.focus), max(0, len(v.matches)-1))
	v.list.reveal = 0
	v.list.revealAlign = ScrollStart
}
func (v *EmojiPickerView) output(e EmojiEntry) string {
	q := strings.TrimSpace(v.query)
	for _, s := range e.SkinTones {
		if s != "" && q == s {
			return s
		}
	}
	return emojiOutput(e, v.tone)
}
func (v *EmojiPickerView) pick(index int) {
	if v.disabled || index < 0 || index >= len(v.entries) {
		return
	}
	e := v.entries[index]
	output := v.output(e)
	changed := output != v.value
	v.value = output
	if v.popup {
		v.SetOpen(false)
	}
	v.SetRecent(append([]string{e.Emoji}, v.recent...))
	if changed && v.onChange != nil {
		v.onChange(output)
	}
	if v.onSelect != nil {
		v.onSelect(EmojiSelection{output, e.Emoji, e.Name})
	}
}
func (v *EmojiPickerView) cellID(index int) string {
	return autoID("emoji", v) + ":" + v.entries[index].Emoji
}
func (v *EmojiPickerView) renderRow(cx *el.Context, row int) el.Element {
	box := el.Div().Row()
	for pos := row * v.columns; pos < min((row+1)*v.columns, len(v.matches)); pos++ {
		index := v.matches[pos]
		entry := v.entries[index]
		id := v.cellID(index)
		if cx.Hovered(id) || cx.Focused(id) {
			v.preview = entry.Name
		}
		cell := el.Div().ID(id).Role("gridcell").Name(entry.Name).Value(v.output(entry)).
			Selected(v.value == v.output(entry)).W(el.Dp(v.cell)).H(el.Dp(v.cell)).Center().
			Rounded(theme.RadiusMd).Focusable(pos == v.focus).CursorPointer().
			Hover(func(s *el.Style) { s.Bg(theme.Subtle) }).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			Child(el.Text(v.output(entry)).TextSize(theme.TextXl)).
			OnClick(func() {
				if index < len(v.entries) && v.entries[index].Emoji == entry.Emoji {
					v.focus = pos
					v.pick(index)
				}
			}).
			OnKey(func(e el.KeyEvent) bool {
				to := pos
				switch key.Name(e.Name) {
				case key.NameLeftArrow:
					to--
				case key.NameRightArrow:
					to++
				case key.NameUpArrow:
					to -= v.columns
				case key.NameDownArrow:
					to += v.columns
				case key.NameHome:
					to = 0
				case key.NameEnd:
					to = len(v.matches) - 1
				case key.NameReturn, key.NameSpace:
					if e.State == el.KeyPress {
						v.pick(index)
					}
					return true
				default:
					return false
				}
				if e.State == el.KeyPress && len(v.matches) > 0 {
					v.focus = min(max(to, 0), len(v.matches)-1)
					v.list.ScrollTo(cx, v.focus/v.columns)
					cx.Focus(v.cellID(v.matches[v.focus]))
				}
				return true
			})
		box.Child(cell)
	}
	return box
}
func (v *EmojiPickerView) panel(cx *el.Context) el.Element {
	box := el.Div().Disabled(v.disabled).W(el.Dp(360)).MaxW(el.Full).Child(el.Widget(v.root))
	if v.sheetPanel {
		box.W(el.Full)
	}
	return box
}
func (v *EmojiPickerView) Render(cx *el.Context) el.Element {
	if !v.popup {
		v.sheetPanel = false
		return v.panel(cx)
	}
	return v.renderPopup(cx)
}
func (v *EmojiPickerView) renderPanel(cx *el.Context) el.Element {
	text := locale.Current()
	v.input.SetLabel(text.EmojiSearch)
	v.input.Placeholder(text.EmojiSearch)
	v.filter()
	v.preview = ""
	width, _ := cx.ViewportSize()
	v.cell = max(float32(44), 44*cx.TextScale())
	columns := max(1, int(width/v.cell))
	if columns != v.columns {
		v.columns = columns
		v.list.reveal = v.focus / columns
	}
	v.list.rowH = v.cell
	v.list.SetCount((len(v.matches) + columns - 1) / columns)
	v.list.Height(v.height)
	tabs := el.Div().Row().Gap(theme.SpaceXs)
	tab := func(label, group string) {
		shown, name := emojiGroupLabel(label)
		b := Button(shown, func() { v.group = group; v.dirty = true; v.focus = 0 }).Size(44).Variant(ButtonGhost).Selected(v.group == group).Name(name)
		tabs.Child(b.Render(cx))
	}
	tab(text.EmojiAll, "")
	tab(text.EmojiRecent, emojiRecentGroup)
	for _, group := range v.groups {
		tab(group, group)
	}
	tones := el.Div().Row().Gap(theme.SpaceXs)
	for tone, symbol := range []string{"✋", "✋🏻", "✋🏼", "✋🏽", "✋🏾", "✋🏿"} {
		b := Button(symbol, func() { v.SetSkinTone(tone) }).Name(text.EmojiTones[tone]).Size(44).Variant(ButtonGhost).Selected(v.tone == tone)
		tones.Child(b.Render(cx))
	}
	name := v.name
	if name == "" {
		name = text.EmojiPicker
	}
	root := el.Div().Role("group").Name(name).Disabled(v.disabled).Gap(theme.SpaceSm).
		Child(v.input.Render(cx), el.Div().ScrollX().Child(tones))
	if !v.sheetPanel {
		root.Child(el.Div().ScrollX().Child(tabs))
	}
	if len(v.matches) == 0 {
		root.Child(el.Div().H(el.Dp(v.height)).Center().Child(el.Text(text.NoMatches).TextColor(theme.Muted)))
	} else {
		root.Child(el.Div().Role("grid").Name(text.EmojiResults).Child(v.list.Render(cx)))
	}
	if v.sheetPanel {
		root.Child(el.Div().ScrollX().Child(tabs))
	}
	label := v.preview
	if label == "" {
		label = text.EmojiChoose
	}
	root.Child(el.Text(label).MaxLines(2).TextColor(theme.Muted))
	return root
}
