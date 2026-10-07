package kit

import (
	"image/color"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
)

type pendingCodeHighlighter struct {
	started chan string
	release chan struct{}
}

func (h *pendingCodeHighlighter) Language(name string) string { return name }
func (h *pendingCodeHighlighter) Highlight(src string, opts core.HighlightOptions) ([]core.CodeToken, bool) {
	h.started <- opts.Language
	<-h.release
	return []core.CodeToken{{Text: src, Color: color.NRGBA{R: 255, A: 255}}}, true
}

func TestCodeEditorLanguageDiscardsPendingHighlight(t *testing.T) {
	previous := core.CurrentHighlighter()
	defer core.SetHighlighter(previous)
	highlighter := &pendingCodeHighlighter{started: make(chan string, 1), release: make(chan struct{})}
	core.SetHighlighter(highlighter)
	ed := CodeEditor("source").Language("go")
	h := editorHarness(ed)
	select {
	case <-highlighter.started:
	case <-time.After(time.Second):
		t.Fatal("highlight did not start")
	}
	ed.Language("")
	close(highlighter.release)
	deadline := time.Now().Add(time.Second)
	for ed.hlRunning && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if ed.hlRunning || len(ed.buf.spans(0)) != 0 || ed.hlRev != 0 {
		t.Fatal("old language's highlight was applied")
	}
}

func TestCodeEditorLanguageClearsSpansAndKeepsSameLanguageCache(t *testing.T) {
	ed := CodeEditor("source").Language("go")
	ed.buf.lines.at(0).spans = []codeSpan{{start: 0, end: 6, color: color.NRGBA{R: 255, A: 255}}}
	ed.hlRev = ed.buf.revision
	ed.Language("go")
	if ed.hlRev != ed.buf.revision || len(ed.buf.spans(0)) != 1 {
		t.Fatal("unchanged language invalidated highlighting")
	}
	ed.Language("")
	if len(ed.buf.spans(0)) != 0 {
		t.Fatal("plain text retained syntax highlighting")
	}
}
