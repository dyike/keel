package window

import (
	"fmt"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestEmojiPickerAgentSelection(t *testing.T) {
	p := kit.EmojiPicker(kit.EmojiEntry{Emoji: "😀", Name: "Grinning face", Group: "Faces"}, kit.EmojiEntry{Emoji: "❤️", Name: "Red heart", Group: "Symbols"}).Name("Choose reaction")
	w := openTest(t, Options{Width: 320, Height: 700, Content: el.Root(p)})
	if e := element(t, w, "Choose reaction"); e.Role != "group" {
		t.Fatal(e)
	}
	cell := element(t, w, "Grinning face")
	if cell.Role != "gridcell" || cell.Width < 44 || cell.Height < 44 {
		t.Fatal(cell)
	}
	w.click(cell.center())
	if p.Value() != "😀" {
		t.Fatal(p.Value())
	}
	if err := w.press("right"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if p.Value() != "❤️" {
		t.Fatal("keyboard", p.Value())
	}
	p.SetDisabled(true)
	if !element(t, w, "Red heart").Disabled {
		t.Fatal("missing disabled state")
	}
}
func TestEmojiPickerAgentPopupSelectCloses(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	for _, width := range []int{390, 900} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			p := kit.EmojiPicker(kit.EmojiEntry{Emoji: "😀", Name: "Grinning face"}).Popup(true).Name("Pick reaction")
			w := openTest(t, Options{Width: width, Height: 800, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Items(el.Start).Child(p.Render(cx)) }))})
			w.click(element(t, w, "Pick reaction").center())
			if !p.Open() {
				t.Fatal("did not open")
			}
			w.click(element(t, w, "Grinning face").center())
			if p.Open() || p.Value() != "😀" {
				t.Fatal("selection failed", p.Open(), p.Value())
			}
			p.SetOpen(true)
			w.snapshot()
			if err := w.press("escape"); err != nil {
				t.Fatal(err)
			}
			if p.Open() {
				t.Fatal("escape did not dismiss")
			}
		})
	}
}
