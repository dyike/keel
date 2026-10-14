package kit

import "testing"

// Keep layout/shaping measurements separate from the optional asynchronous
// whole-document highlighter measured by BenchmarkCodeEditorLargeFileFrame.
func BenchmarkCodeEditorPlainLargeFileFrame(b *testing.B) {
	ed := CodeEditor(bigSource(200000))
	h := editorHarness(ed)
	h.Frame()
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		ed.SetCursor(i*997%200000, 0)
		h.Frame()
	}
}
