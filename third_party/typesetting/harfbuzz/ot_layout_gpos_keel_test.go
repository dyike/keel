package harfbuzz

import "testing"

// Keel patch test: an attachment chain pointing before the buffer is
// ignored, as in HarfBuzz, instead of indexing pos[-1].
func TestPropagateAttachmentOffsetsBeforeBuffer(t *testing.T) {
	pos := make([]GlyphPosition, 3)
	pos[0].attachChain = -1
	pos[0].attachType = attachTypeMark
	pos[0].XOffset = 7
	propagateAttachmentOffsets(pos, 0, LeftToRight)
	if pos[0].XOffset != 7 || pos[0].attachChain != 0 {
		t.Fatalf("offset %d chain %d", pos[0].XOffset, pos[0].attachChain)
	}
}
