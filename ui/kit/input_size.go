package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// InputSize sets the field text and frame density. Medium preserves defaults.
type InputSize uint8

const (
	InputSizeMedium InputSize = iota
	InputSizeXSmall
	InputSizeSmall
	InputSizeLarge
)

func (v *InputView) Size(size InputSize) *InputView {
	if size <= InputSizeLarge {
		v.size = size
	}
	return v
}
func (v *InputView) applySize(text *el.InputEl, box *el.DivEl) {
	if v.size == InputSizeMedium {
		if v.multiline && v.minRows == 0 {
			text.MinH(el.Sp(float32(v.rows) * 22))
		}
		return
	}
	height, font, horizontal, vertical := float32(28), float32(theme.TextSm), float32(8), float32(3)
	switch v.size {
	case InputSizeXSmall:
		height, font, horizontal, vertical = 24, theme.TextXs, 6, 2
	case InputSizeLarge:
		height, font, horizontal, vertical = 40, theme.TextBody, 12, 6
	}
	text.TextSize(font)
	box.MinH(el.Dp(height)).Px(horizontal).Py(vertical)
	if v.multiline {
		box.Py(vertical * 2)
		if v.minRows == 0 {
			text.MinH(el.Sp(float32(v.rows) * font * 1.6))
		}
	}
}
