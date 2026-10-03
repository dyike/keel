package kit

import (
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestAttachmentPartStatusInheritanceAndActions(t *testing.T) {
	retries := 0
	a := Attachment("override.txt", 1024).OnRetry(func() { retries++ })
	a.SetError("Network failed")
	a.Description("Previous upload completed").PartStatus(AttachmentPartDescription, AttachmentStatusComplete).PartStatus(AttachmentPartTitle, AttachmentStatusProcessing)
	h := renderView(a, 280, 1)
	if !shown(h, "Previous upload completed") || a.Status() != AttachmentStatusFailed {
		t.Fatal("override changed lifecycle")
	}
	if a.titleShimmer.disabled {
		t.Fatal("title did not use override")
	}
	a.SetStatus(AttachmentStatusComplete)
	h.Frame()
	if a.titleShimmer.disabled {
		t.Fatal("parent change discarded title override")
	}
	a.ClearPartStatus(AttachmentPartTitle)
	h.Frame()
	if !a.titleShimmer.disabled {
		t.Fatal("title did not resume inheritance")
	}
	a.ClearDescription()
	h.Frame()
	if !shown(h, "1.0 KB") {
		t.Fatal("automatic description not restored")
	}
	a.SetError("Again failed")
	h.Frame()
	if !shown(h, "1.0 KB") {
		t.Fatal("description override lost")
	}
	a.ClearPartStatus(AttachmentPartDescription)
	h.Frame()
	if !shown(h, "Again failed") {
		t.Fatal("description inheritance not restored")
	}
	a.PartStatus(AttachmentPartDescription, AttachmentStatusUploading)
	h.Frame()
	if !shown(h, locale.Current().Uploading+" 0%") {
		t.Fatal("unknown progress should display zero")
	}
	a.PartStatus(AttachmentPartDescription, AttachmentStatus(255)).PartStatus(AttachmentPartRoot, AttachmentStatusComplete)
	h.Frame()
	if a.Status() != AttachmentStatusFailed || !shown(h, locale.Current().Uploading+" 0%") {
		t.Fatal("invalid override changed state")
	}
	click(t, h, locale.Current().Name(locale.Current().Retry, "override.txt"))
	if retries != 1 || a.Status() != AttachmentStatusUploading {
		t.Fatal("override interfered with retry")
	}
}
