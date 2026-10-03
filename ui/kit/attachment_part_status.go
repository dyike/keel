package kit

import (
	"image/color"
	"strconv"

	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// PartStatus overrides the default title animation or description presentation.
// Only Title and Description accept overrides; other parts and invalid states
// are ignored. The attachment lifecycle, media and action availability stay owned
// by SetStatus/SetProgress/SetError.
func (v *AttachmentView) PartStatus(part AttachmentPart, status AttachmentStatus) *AttachmentView {
	if status > AttachmentStatusCanceled {
		return v
	}
	switch part {
	case AttachmentPartTitle:
		v.titleStatus = &status
	case AttachmentPartDescription:
		v.descriptionStatus = &status
	}
	return v
}

// ClearPartStatus restores inheritance from the attachment's effective status.
func (v *AttachmentView) ClearPartStatus(part AttachmentPart) *AttachmentView {
	switch part {
	case AttachmentPartTitle:
		v.titleStatus = nil
	case AttachmentPartDescription:
		v.descriptionStatus = nil
	}
	return v
}

// Description replaces automatic status/size text, retaining the description's
// effective status color. An empty string displays no description text.
func (v *AttachmentView) Description(text string) *AttachmentView { v.description = &text; return v }

// ClearDescription restores the automatic localized status/size text.
func (v *AttachmentView) ClearDescription() *AttachmentView { v.description = nil; return v }

func (v *AttachmentView) partStatus(part AttachmentPart) AttachmentStatus {
	if part == AttachmentPartTitle && v.titleStatus != nil {
		return *v.titleStatus
	}
	if part == AttachmentPartDescription && v.descriptionStatus != nil {
		return *v.descriptionStatus
	}
	return v.Status()
}

func (v *AttachmentView) statusDetail(status AttachmentStatus) (state, detail string, tint color.NRGBA) {
	text := locale.Current()
	detail, tint = FileSize(v.size), theme.Muted
	switch status {
	case AttachmentStatusCanceled:
		state, detail = "canceled", text.Canceled
	case AttachmentStatusPending:
		state, detail = "pending", text.AttachmentPending
	case AttachmentStatusProcessing:
		state, detail = "processing", text.AttachmentProcessing
	case AttachmentStatusFailed:
		state, detail, tint = "error", v.err, theme.DangerText
		if detail == "" {
			detail = text.AttachmentFailed
		}
	case AttachmentStatusUploading:
		pct := strconv.Itoa(int(max(v.progress, 0)*100+.5)) + "%"
		state, detail = text.Uploading+" "+pct, text.Uploading+" "+pct
	}
	return
}
