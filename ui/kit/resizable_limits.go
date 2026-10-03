package kit

// Max limits the two pane sizes in dp. Zero removes that pane's maximum.
// Negative/non-finite values are ignored independently. A maximum below Min
// uses Min. If both maxima leave spare space, it remains after the second pane.
func (v *ResizableView) Max(first, second float32) *ResizableView {
	if first >= 0 && finiteNumber(float64(first)) {
		v.max1 = first
	}
	if second >= 0 && finiteNumber(float64(second)) {
		v.max2 = second
	}
	return v
}
