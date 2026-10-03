package kit

// AttachmentSize selects a coordinated card, media, text and control scale.
type AttachmentSize uint8

const (
	AttachmentSizeMedium AttachmentSize = iota
	AttachmentSizeXSmall
	AttachmentSizeSmall
	AttachmentSizeLarge
)

// Size sets the density. Medium is the default; explicit part styles apply last.
func (v *AttachmentView) Size(size AttachmentSize) *AttachmentView {
	if size <= AttachmentSizeLarge {
		v.density = size
	}
	return v
}

type attachmentMetrics struct{ width, height, media, title, description, padding, gap, action float32 }

func (v *AttachmentView) metrics() attachmentMetrics {
	switch v.density {
	case AttachmentSizeXSmall:
		return attachmentMetrics{176, 40, 28, 11, 10, 5, 6, 20}
	case AttachmentSizeSmall:
		return attachmentMetrics{200, 48, 32, 12, 11, 7, 8, 22}
	case AttachmentSizeLarge:
		return attachmentMetrics{272, 64, 44, 14, 12, 9, 12, 28}
	default:
		return attachmentMetrics{232, 56, 38, 13, 11, 8, 10, 24}
	}
}
