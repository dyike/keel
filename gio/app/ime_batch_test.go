// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"testing"
	"unicode/utf8"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/io/key"
)

type imeBatchDriver struct {
	driver
	edits int
	last  key.EditEvent
}

func (d *imeBatchDriver) ProcessEvent(e event.Event) {
	if edit, ok := e.(key.EditEvent); ok {
		d.edits++
		d.last = edit
	}
}

func imeBatchCallbacks() (*callbacks, *imeBatchDriver) {
	d := new(imeBatchDriver)
	w := &Window{driver: d}
	w.imeState = editorState{
		EditorState: input.EditorState{
			Snippet: key.Snippet{Range: key.Range{Start: 0, End: 5}, Text: "A🙂niZ"},
		},
		compose: key.Range{Start: 2, End: 4},
	}
	w.imeState.Selection.Range = key.Range{Start: 4, End: 4}
	return &callbacks{w: w}, d
}

func TestIMEBatchPreeditReplacesOnce(t *testing.T) {
	for _, tc := range []struct {
		name               string
		update             textInputUpdate
		text               string
		compose, selection key.Range
	}{
		{"preedit", textInputUpdate{preeditSet: true, preedit: "你好", preeditSelection: key.Range{Start: 6, End: 6}}, "A🙂你好Z", key.Range{Start: 2, End: 4}, key.Range{Start: 4, End: 4}},
		{"commit", textInputUpdate{commitSet: true, commit: "你好"}, "A🙂你好Z", key.Range{Start: -1, End: -1}, key.Range{Start: 4, End: 4}},
		{"commit-and-preedit", textInputUpdate{commitSet: true, commit: "你好", preeditSet: true, preedit: "shi", preeditSelection: key.Range{Start: 3, End: 3}}, "A🙂你好shiZ", key.Range{Start: 4, End: 7}, key.Range{Start: 7, End: 7}},
		{"cancel", textInputUpdate{}, "A🙂Z", key.Range{Start: -1, End: -1}, key.Range{Start: 2, End: 2}},
		{"cursor-in-preedit", textInputUpdate{preeditSet: true, preedit: "你好", preeditSelection: key.Range{Start: 3, End: 3}}, "A🙂你好Z", key.Range{Start: 2, End: 4}, key.Range{Start: 3, End: 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, d := imeBatchCallbacks()
			s := textInputState{pending: tc.update}
			s.apply(c)
			got := c.EditorState()
			if got.Snippet.Text != tc.text || got.compose != tc.compose || got.Selection.Range != tc.selection {
				t.Fatalf("wrong IME state: %+v", got)
			}
			if d.edits != 1 || d.last.Range != (key.Range{Start: 2, End: 4}) {
				t.Fatalf("intermediate clear escaped to input bindings: edits=%d last=%+v", d.edits, d.last)
			}
		})
	}
}

func TestIMEBatchSerialAndSurrounding(t *testing.T) {
	c, _ := imeBatchCallbacks()
	s := textInputState{serial: 42, pending: textInputUpdate{preeditSet: true, preedit: "你好", preeditSelection: key.Range{Start: 6, End: 6}}}
	if s.applyDone(c, 41) || !s.stale || !s.dirty {
		t.Fatal("stale serial must apply the edit without sending protocol state")
	}
	if c.EditorState().Snippet.Text != "A🙂你好Z" {
		t.Fatal("stale done dropped text")
	}
	s.pending = textInputUpdate{preeditSet: true, preedit: "你好吗", preeditSelection: key.Range{Start: 9, End: 9}}
	if !s.applyDone(c, 42) || s.stale {
		t.Fatal("matching serial did not release deferred state")
	}
	snapshot, status := textInputSurrounding(c.EditorState())
	if status != surroundingReady || snapshot.text != "A🙂Z" || snapshot.cursor != len("A🙂") {
		t.Fatalf("preedit leaked into surrounding text: %+v", snapshot)
	}
}

func BenchmarkIMEBatchPreedit(b *testing.B) {
	c, d := imeBatchCallbacks()
	s := textInputState{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		text := "nihao"
		if i%2 != 0 {
			text = "nihaoshijie"
		}
		s.pending = textInputUpdate{preeditSet: true, preedit: text, preeditSelection: key.Range{Start: len(text), End: len(text)}}
		s.apply(c)
		if c.EditorState().compose.End != 2+utf8.RuneCountInString(text) {
			b.Fatal("bad composition")
		}
	}
	b.ReportMetric(float64(d.edits)/float64(b.N), "edits/key")
}
