package kit

import (
	"github.com/dyike/keel/ui/el"
	"strconv"
	"testing"
)

func TestMessageAllPartStyles(t *testing.T) {
	for _, scale := range []int{1, 2} {
		msg := Message("Author", Label("Body")).Header(Label("Header")).Footer(Label("Footer")).Actions(Button("Action", nil)).Reactions(MessageReaction{Name: "Like", Count: 1})
		msg.SetState(MessageSending, "")
		seen := map[MessagePart]int{}
		for part := MessagePartRoot; part <= MessagePartReactions; part++ {
			msg.PartStyle(part, func(e *el.DivEl) {
				seen[part]++
				e.Px(3)
				if part != MessagePartRoot {
					e.Name("part-" + strconv.Itoa(int(part)))
				}
			})
		}
		msg.PartStyle(MessagePart(255), func(*el.DivEl) { t.Fatal("invalid part called") })
		h := renderView(msg, 260, scale)
		for part := MessagePartRoot; part <= MessagePartReactions; part++ {
			if seen[part] == 0 {
				t.Fatal("part not styled", part)
			}
			if part != MessagePartRoot {
				b := bounds(h, "part-"+strconv.Itoa(int(part)))
				if b.Dx() <= 0 || b.Min.X < 0 || b.Max.X > 260*scale {
					t.Fatal("part layout", part, b)
				}
			}
		}
		msg.PartStyle(MessagePartFooter, nil)
		h.Frame()
		if shown(h, "part-"+strconv.Itoa(int(MessagePartFooter))) {
			t.Fatal("style not cleared")
		}
	}
}
