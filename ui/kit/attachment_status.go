package kit

// AttachmentStatus describes an attachment's application-controlled lifecycle.
// Canceled is Keel's additional state, retained for the cancel/retry workflow.
type AttachmentStatus uint8

const (
	AttachmentStatusComplete AttachmentStatus = iota
	AttachmentStatusPending
	AttachmentStatusUploading
	AttachmentStatusProcessing
	AttachmentStatusFailed
	AttachmentStatusCanceled
)

func (s AttachmentStatus) IsPending() bool    { return s == AttachmentStatusPending }
func (s AttachmentStatus) IsUploading() bool  { return s == AttachmentStatusUploading }
func (s AttachmentStatus) IsProcessing() bool { return s == AttachmentStatusProcessing }
func (s AttachmentStatus) IsFailed() bool     { return s == AttachmentStatusFailed }
func (s AttachmentStatus) IsComplete() bool   { return s == AttachmentStatusComplete }
func (s AttachmentStatus) IsInProgress() bool { return s.IsUploading() || s.IsProcessing() }

// Status reads the effective state. An error temporarily overlays the base
// state, so SetError("") restores the state that preceded the error.
func (v *AttachmentView) Status() AttachmentStatus {
	if v.canceled {
		return AttachmentStatusCanceled
	}
	if v.err != "" {
		return AttachmentStatusFailed
	}
	return v.status
}

// SetStatus changes state without invoking callbacks. It clears any previous
// error/cancellation. Uploading retains a known fraction, or starts at zero;
// other states clear the fraction. Invalid states are ignored.
func (v *AttachmentView) SetStatus(status AttachmentStatus) {
	if status > AttachmentStatusCanceled {
		return
	}
	v.status = status
	v.err = ""
	v.canceled = status == AttachmentStatusCanceled
	if status == AttachmentStatusUploading {
		v.progress = max(v.progress, 0)
	} else {
		v.progress = -1
	}
}
