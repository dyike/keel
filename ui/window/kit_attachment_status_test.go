package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestAttachmentLifecycleAgentValues(t *testing.T) {
	a := kit.Attachment("file", 1024)
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Root(a)})
	for _, tc := range []struct {
		status kit.AttachmentStatus
		value  string
	}{
		{kit.AttachmentStatusPending, "pending"}, {kit.AttachmentStatusProcessing, "processing"},
		{kit.AttachmentStatusFailed, "error"}, {kit.AttachmentStatusCanceled, "canceled"}, {kit.AttachmentStatusComplete, ""},
	} {
		a.SetStatus(tc.status)
		w.render()
		e := element(t, w, "file")
		if e.Role != "attachment" || e.Value != tc.value {
			t.Fatal("state semantics", tc.status, e)
		}
	}
}
