package kit

import (
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestAttachmentLifecycleAndLegacyTransitions(t *testing.T) {
	a := Attachment("file", 0)
	if !a.Status().IsComplete() {
		t.Fatal("default")
	}
	for _, s := range []AttachmentStatus{AttachmentStatusPending, AttachmentStatusUploading, AttachmentStatusProcessing, AttachmentStatusFailed, AttachmentStatusComplete, AttachmentStatusCanceled} {
		a.SetStatus(s)
		if a.Status() != s {
			t.Fatal("set status", s, a.Status())
		}
		a.SetStatus(255)
		if a.Status() != s {
			t.Fatal("invalid state")
		}
	}
	a.SetProgress(.7)
	if !a.Status().IsUploading() || !a.Status().IsInProgress() {
		t.Fatal("legacy progress")
	}
	a.SetError("offline")
	if !a.Status().IsFailed() {
		t.Fatal("legacy error")
	}
	a.SetError("")
	if !a.Status().IsUploading() || a.progress != .7 {
		t.Fatal("error clear lost upload")
	}
	a.SetStatus(AttachmentStatusProcessing)
	a.SetError("offline")
	a.SetError("")
	if !a.Status().IsProcessing() {
		t.Fatal("error clear lost processing")
	}
	a.SetProgress(-1)
	if !a.Status().IsComplete() || a.Status().IsInProgress() {
		t.Fatal("legacy complete")
	}
}

func TestAttachmentStateActionsAndLocalization(t *testing.T) {
	old := locale.Current()
	defer locale.Apply(old)
	for _, lang := range []locale.Strings{locale.Chinese(), locale.English()} {
		locale.Apply(lang)
		opens, cancels, retries := 0, 0, 0
		a := Attachment("file", 0).OnOpen(func() { opens++ }).OnCancel(func() { cancels++ }).OnRetry(func() { retries++ })
		a.SetStatus(AttachmentStatusPending)
		h := renderView(a, 300, 1)
		if !shown(h, lang.AttachmentPending) {
			t.Fatal("pending label")
		}
		click(t, h, "file")
		if opens != 0 || shown(h, lang.Name(lang.Cancel, "file")) {
			t.Fatal("pending actions")
		}
		a.SetStatus(AttachmentStatusProcessing)
		h.Frame()
		if !shown(h, lang.AttachmentProcessing) {
			t.Fatal("processing label")
		}
		click(t, h, "file")
		if opens != 0 {
			t.Fatal("processing opened")
		}
		click(t, h, lang.Name(lang.Cancel, "file"))
		if cancels != 1 || a.Status() != AttachmentStatusCanceled {
			t.Fatal("processing cancel")
		}
		click(t, h, lang.Name(lang.Retry, "file"))
		if retries != 1 || !a.Status().IsUploading() || a.progress != 0 {
			t.Fatal("cancel retry")
		}
		a.SetStatus(AttachmentStatusFailed)
		h.Frame()
		if !shown(h, lang.AttachmentFailed) {
			t.Fatal("reasonless failure")
		}
		click(t, h, lang.Name(lang.Retry, "file"))
		if retries != 2 || !a.Status().IsUploading() {
			t.Fatal("failed retry")
		}
		a.SetStatus(AttachmentStatusComplete)
		h.Frame()
		click(t, h, "file")
		if opens != 1 {
			t.Fatal("completed open")
		}
	}
}
