package core

// ReloadState is implemented by window content whose UI state survives a
// keel run reload: el's roots keep scroll offsets and the text of inputs.
// ui/window calls both methods under the frame lock; applications need not.
type ReloadState interface {
	// SaveReloadState encodes the state of the last frame drawn.
	SaveReloadState() []byte
	// RestoreReloadState applies saved state to the matching elements of the
	// next frame, before it is laid out.
	RestoreReloadState([]byte)
}
