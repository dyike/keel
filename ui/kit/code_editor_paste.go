package kit

import (
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/dyike/keel/third_party/gio/io/clipboard"

	"github.com/dyike/keel/third_party/gio/io/transfer"
	"github.com/dyike/keel/ui/core"
)

type codePasteRequest struct {
	revision        uint64
	sels            []codeSel
	primary         int
	data            core.ClipboardData
	err             error
	ready, fallback bool
}

// OnPaste intercepts clipboard contents before multi-selection text insertion.
// Returning true consumes the paste. Read-only and disabled editors ignore it.
func (v *CodeEditorView) OnPaste(fn func(core.ClipboardData) bool) *CodeEditorView {
	v.onPaste = fn
	return v
}

// PasteReader supplies an asynchronous rich clipboard source. Nil uses Gio text.
func (v *CodeEditorView) PasteReader(reader core.ClipboardReader) *CodeEditorView {
	v.pasteReader = reader
	v.pendingPaste = nil
	return v
}

// OnPasteError reports a rich read failure before falling back to Gio text, or
// a text read failure before rejecting that paste. Runs on the UI thread.
func (v *CodeEditorView) OnPasteError(fn func(error)) *CodeEditorView { v.onPasteError = fn; return v }

func (v *CodeEditorView) requestPaste(gtx core.C) {
	if v.readOnly || v.disabled || !gtx.Enabled() {
		return
	}
	r := &codePasteRequest{revision: v.buf.revision, sels: slices.Clone(v.sels), primary: v.prim}
	v.pendingPaste = r
	if v.pasteReader != nil {
		var once sync.Once
		v.pasteReader(func(data core.ClipboardData, err error) {
			once.Do(func() {
				core.Update(func() {
					if v.pendingPaste == r {
						r.data, r.err, r.ready = data, err, true
					}
				})
			})
		})
	} else {
		r.fallback = true
		gtx.Execute(clipboard.ReadCmd{Tag: v})
	}
}
func (v *CodeEditorView) pasteRequestValid(r *codePasteRequest) bool {
	return !v.readOnly && !v.disabled && r.revision == v.buf.revision && r.primary == v.prim && slices.Equal(r.sels, v.sels)
}
func readCodePaste(event transfer.DataEvent) (string, error) {
	reader := event.Open()
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, (16<<20)+1))
	if len(data) > 16<<20 {
		return "", fmt.Errorf("clipboard text exceeds 16MiB")
	}
	return string(data), err
}
func (v *CodeEditorView) pasteCompletion(gtx core.C) {
	r := v.pendingPaste
	if r == nil {
		return
	}

	if !gtx.Enabled() || !v.pasteRequestValid(r) {
		v.pendingPaste = nil
		return
	}
	if !r.ready {
		return
	}
	if r.err != nil {
		if v.onPasteError != nil {
			core.Call(gtx, func() { v.onPasteError(r.err) })
		}
		if !r.fallback && v.pendingPaste == r && v.pasteRequestValid(r) {
			r.err, r.ready, r.fallback = nil, false, true
			gtx.Execute(clipboard.ReadCmd{Tag: v})
			return
		}
		if v.pendingPaste == r {
			v.pendingPaste = nil
		}
		return
	}
	v.pendingPaste = nil
	consumed := false
	if v.onPaste != nil {
		core.Call(gtx, func() { consumed = v.onPaste(r.data) })
	}
	if !consumed && v.pasteRequestValid(r) && r.data.Text != "" {
		v.paste(gtx, r.data.Text)
	}
}
