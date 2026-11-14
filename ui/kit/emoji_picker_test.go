package kit

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func testEmojis() []EmojiEntry {
	return []EmojiEntry{
		{Emoji: "😀", Name: "Grinning face", Group: "Faces", Keywords: []string{"smile", "开心"}},
		{Emoji: "👍", Name: "Thumbs up", Group: "People", Keywords: []string{"yes"}, SkinTones: [5]string{"👍🏻", "👍🏼", "👍🏽", "👍🏾", "👍🏿"}},
		{Emoji: "❤️", Name: "Red heart", Group: "Symbols"},
		{Emoji: "🐈‍⬛", Name: "Black cat", Group: "Animals"},
	}
}

func TestEmojiPickerCatalogCompleteAndIndependent(t *testing.T) {
	entries := DefaultEmojis()
	seen := map[string]bool{}
	groups := map[string]bool{}
	for _, e := range entries {
		if e.Emoji == "" || e.Name == "" || e.Group == "" {
			t.Fatal(e)
		}
		groups[e.Group] = true
		for _, s := range append([]string{e.Emoji}, e.SkinTones[:]...) {
			if s == "" {
				continue
			}
			if seen[s] {
				t.Fatal("duplicate", s)
			}
			seen[s] = true
		}
	}
	if len(seen) != emojiCatalogSequenceCount || len(groups) != 9 {
		t.Fatal(len(seen), len(groups))
	}
	for _, s := range []string{"❤️", "👍🏽", "👨‍👩‍👧‍👦", "🫱🏻‍🫲🏿", "🇨🇳", "🐈‍⬛"} {
		if !seen[s] {
			t.Fatal("missing full sequence", s)
		}
	}
	entries[0].Name = "changed"
	if DefaultEmojis()[0].Name == "changed" {
		t.Fatal("default catalog mutated")
	}
}
func TestEmojiPickerSelectionToneRecentsAndSilentSetters(t *testing.T) {
	v := EmojiPicker(testEmojis()...)
	changes, selects := 0, 0
	v.OnChange(func(string) { changes++ }).OnSelect(func(s EmojiSelection) {
		selects++
		if s.Base != "👍" || s.Emoji != "👍🏽" {
			t.Fatal(s)
		}
	})
	v.SetSkinTone(3)
	v.SetValue("❤️")
	v.SetRecent([]string{"😀", "👍🏻", "😀", "unknown"})
	if changes != 0 || selects != 0 || len(v.Recent()) != 2 || v.Recent()[1] != "👍" {
		t.Fatal("assignment", v.Recent())
	}
	v.SetQuery("thumbs")
	h := renderView(v, 320, 1)
	click(t, h, "Thumbs up")
	click(t, h, "Thumbs up")
	if v.Value() != "👍🏽" || changes != 1 || selects != 2 || v.Recent()[0] != "👍" {
		t.Fatal(v.Value(), changes, selects, v.Recent())
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "Thumbs up")
	if selects != 2 {
		t.Fatal("disabled")
	}
	v.SetValue("😀")
	if v.Value() != "😀" || changes != 1 {
		t.Fatal("silent disabled assignment")
	}
	recent := v.Recent()
	recent[0] = "changed"
	if v.Recent()[0] == "changed" {
		t.Fatal("mutable recents")
	}
}
func TestEmojiPickerSearchGroupsAndDataOwnership(t *testing.T) {
	data := testEmojis()
	v := EmojiPicker(data...)
	data[0].Keywords[0] = "changed"
	v.SetQuery("SMILE 开心")
	v.filter()
	if len(v.matches) != 1 || v.entries[v.matches[0]].Emoji != "😀" {
		t.Fatal(v.matches)
	}
	v.SetQuery("👍🏿")
	v.filter()
	if len(v.matches) != 1 || v.output(v.entries[v.matches[0]]) != "👍🏿" {
		t.Fatal("exact variant search")
	}
	v.SetQuery("")
	v.SetGroup("Animals")
	v.filter()
	if len(v.matches) != 1 || v.entries[v.matches[0]].Emoji != "🐈‍⬛" {
		t.Fatal("group")
	}
	v.Data(nil)
	v.filter()
	if len(v.matches) != 0 || len(v.Groups()) != 0 {
		t.Fatal("empty data")
	}
}
func TestEmojiPickerTypedSearchAndAncestorDisabled(t *testing.T) {
	v := EmojiPicker(testEmojis()...)
	disabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 320, 1)
	clickClass(t, h, "Editor", locale.Current().EmojiSearch)
	typeAt(h, "black")
	h.Frame()
	if v.Query() != "black" || len(v.matches) != 1 {
		t.Fatal(v.Query(), v.matches)
	}
	disabled = true
	h.Frame()
	click(t, h, "Black cat")
	if v.Value() != "" {
		t.Fatal("ancestor disabled")
	}
	disabled = false
	h.Frame()
	click(t, h, "Black cat")
	if v.Value() != "🐈‍⬛" {
		t.Fatal(v.Value())
	}
}
func TestEmojiPickerKeyboardMovesAndSelects(t *testing.T) {
	v := EmojiPicker(testEmojis()...)
	h := renderView(v, 320, 1)
	click(t, h, "Grinning face")
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if v.Value() != "👍" {
		t.Fatal("keyboard", v.Value())
	}
}

func TestEmojiPickerVirtualGridEndAndSearchReset(t *testing.T) {
	v := EmojiPicker()
	h := renderView(v, 320, 1)
	if len(v.matches) < 1000 {
		t.Fatal("default catalog too small")
	}
	click(t, h, v.entries[v.matches[0]].Name)
	h.Key(key.NameEnd, 0)
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	last := v.entries[v.matches[len(v.matches)-1]].Emoji
	if v.Value() != last {
		t.Fatal("offscreen keyboard target", v.Value(), last)
	}
	v.SetQuery("grinning face")
	h.Frame()
	h.Frame()
	if len(v.matches) == 0 {
		t.Fatal("search")
	}
	click(t, h, v.entries[v.matches[0]].Name)
	if v.Value() == last {
		t.Fatal("shortened list retained old scroll")
	}
}
func TestEmojiPickerGridUsesContainerWidth(t *testing.T) {
	v := EmojiPicker()
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(190)).Child(v.Render(cx)) }), 900, 1)
	if v.columns != 4 {
		t.Fatal("used window rather than container", v.columns)
	}
	first := bounds(h, v.entries[0].Name)
	if first.Empty() || first.Max.X > 190 {
		t.Fatal(first)
	}
}
