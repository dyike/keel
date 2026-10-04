package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
	"time"
)

func TestTimeFieldQuickTypingInWindow(t *testing.T) {
	v := kit.TimeField("Clock").Seconds().Hour12(false).SegmentKeys(true)
	w := openTest(t, kitPage(v))
	w.click(element(t, w, "时 Clock").center())
	w.render()
	if err := w.typeText("09"); err != nil {
		t.Fatal(err)
	}
	w.render()
	w.render()
	if err := w.typeText("45"); err != nil {
		t.Fatal(err)
	}
	w.render()
	w.render()
	if err := w.typeText("30"); err != nil {
		t.Fatal(err)
	}
	if v.Value() != 9*time.Hour+45*time.Minute+30*time.Second {
		t.Fatal("typing did not advance across segments", v.Value())
	}
}
