package el

import (
	"image"
	"strings"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDocumentEditorKeepsWheelOffset(t *testing.T) {
	var doc InputDocument
	_ = doc.SetText(strings.Repeat("中文 line of text\n", 40))
	var field *InputEl
	root := Root(viewFunc(func(*Context) Element {
		field = TextArea().ID("draft").Document(&doc).AutoGrow(2, 4).W(Dp(240)).Scrollbars(ScrollbarAlways)
		return Div().Items(Start).Child(field)
	}))
	h := uitest.New(root)
	h.Frame()
	st := root.store.states[field.n.key]
	sc, ok := any(&st.editor).(editorViewport)
	if !ok {
		t.Skip("upstream Gio has native scrolling only")
	}
	h.Scroll(40, 30, 60)
	for range 4 {
		h.Frame()
	}
	if sc.ScrollOffset().Y <= 0 {
		t.Fatal("document editor reset wheel offset to caret")
	}
	if st.editor.Text() != doc.presentation().Text {
		t.Fatal("scroll changed text")
	}
	h.Click(40, 30)
	h.Key(key.NameEnd, key.ModShortcut)
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: len([]rune(doc.Content().Text())), End: len([]rune(doc.Content().Text()))}, Text: "tail"})
	h.Frame()
	h.Frame()
	if !strings.HasSuffix(doc.Content().Text(), "tail") {
		t.Fatalf("editing failed after scrolling: %q", doc.Content().Text())
	}
	if sc.ScrollOffset().Y < sc.ScrollBounds().Max.Y-5 {
		t.Fatal("editing did not reveal caret")
	}
	_ = doc.SetText("short")
	h.Frame()
	h.Frame()
	if sc.ScrollOffset() != (image.Point{}) {
		t.Fatal("short document kept offset")
	}
}
