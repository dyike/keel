package el

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/editorstyle"
	"github.com/dyike/keel/ui/internal/inputcontent"
)

func documentSyncEditor(st *elemState, d *InputDocument) {
	if !d.session.Composing() {
		st.inputComposition = key.Range{Start: -1, End: -1}
	}
	p := d.presentation()
	if st.editor.Text() != p.Text {
		st.editor.SetText(p.Text)
	}
	r := d.Selection()
	visible := InputRange{Start: p.DisplayOffset(r.Start, 0), End: p.DisplayOffset(r.End, 0)}
	runes, err := inputcontent.RuneRange(p.Text, visible)
	if err == nil {
		st.editor.SetCaret(runes.Start, runes.End)
	}
	st.lastText = p.Text
	st.inputDocument, st.inputDocumentRevision = d, d.revision
}
func documentReadSelection(st *elemState, d *InputDocument) {
	a, b := st.editor.Selection()
	r, err := inputcontent.ByteRange(d.presentation().Text, InputRange{Start: a, End: b})
	if err == nil {
		_ = documentSelectLayout(d, r)
	}
	documentSyncEditor(st, d)
}
func documentReplaceSelection(d *InputDocument, text string) error {
	return d.session.ReplaceSource(d.Selection(), text)
}

func (e *engine) inputDocumentAction(n *Node, st *elemState, action InputAction) {
	if !e.gtx.Enabled() || st.disabled {
		return
	}
	d := n.input.document
	switch action {
	case InputCopy, InputCut:
		if action == InputCut && n.input.readOnly {
			return
		}
		if text := d.session.SelectedText(); text != "" {
			e.gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
			if action == InputCut {
				_ = documentReplaceSelection(d, "")
			}
		}
	case InputPaste:
		if !n.input.readOnly {
			e.requestInputPaste(n, st)
		}
	case InputSelectAll:
		_ = d.session.SelectSource(InputRange{Start: 0, End: len(d.Content().Text())})
	case InputUndo, InputRedo:
		if !n.input.readOnly {
			if action == InputUndo {
				d.session.Undo()
			} else {
				d.session.Redo()
			}
		}
	}
	documentSyncEditor(st, d)
}

func (e *engine) inputDocumentEvents(n *Node, st *elemState) {
	spec, d, ed := n.input, n.input.document, &st.editor
	prepareInputTokenHits(st, len(d.Content().Tokens()))
	if st.inputDocument != d || st.inputDocumentRevision != d.revision {
		documentSyncEditor(st, d)
	} else {
		documentReadSelection(st, d)
	}
	ed.Mask, ed.Filter, ed.MaxLen = 0, "", 0
	if spec.readOnly || !e.gtx.Enabled() {
		d.session.EndComposition()
		st.inputComposition = key.Range{Start: -1, End: -1}
	}
	before := d.Content()
	beforeRevision := d.revision
	filters := []event.Filter{key.FocusFilter{Target: ed}, transfer.TargetFilter{Target: ed, Type: "application/text"}}
	for _, name := range []key.Name{"C", "X", "V", "Z", "Y"} {
		filters = append(filters, key.Filter{Focus: ed, Name: name, Required: key.ModShortcut, Optional: key.ModShift})
	}
	for _, name := range []key.Name{key.NameDeleteBackward, key.NameDeleteForward, key.NameLeftArrow, key.NameRightArrow, key.NameEnter, key.NameReturn} {
		optional := key.ModShift
		if name != key.NameEnter && name != key.NameReturn {
			optional |= key.ModShortcutAlt
		}
		filters = append(filters, key.Filter{Focus: ed, Name: name, Optional: optional})
	}
	for {
		ev, ok := e.gtx.Event(filters...)
		if !ok {
			break
		}
		switch ev := ev.(type) {
		case key.FocusEvent:
			if ev.Focus && !spec.readOnly {
				e.gtx.Execute(key.SoftKeyboardCmd{Show: true})
			}
			if !ev.Focus {
				d.session.EndComposition()
			}
		case key.CompositionEvent:
			if !spec.readOnly {
				st.inputComposition = key.Range(ev)
				if ev.Start < 0 {
					d.session.EndComposition()
				} else {
					d.session.BeginComposition()
				}
			}
		case key.SelectionEvent:
			r, err := inputcontent.ByteRange(d.Content().Presentation().Text, InputRange{Start: ev.Start, End: ev.End})
			if err == nil {
				_ = d.session.SelectDisplay(r)
			}
		case key.EditEvent:
			if spec.readOnly {
				continue
			}
			r, err := inputcontent.ByteRange(d.Content().Presentation().Text, InputRange{Start: ev.Range.Start, End: ev.Range.End})
			if err == nil {
				text := ev.Text
				if !spec.multiline {
					text = strings.ReplaceAll(text, "\n", " ")
				}
				_ = d.session.ReplaceDisplay(r, text)
			}
		case transfer.DataEvent:
			reader := ev.Open()
			data, err := io.ReadAll(io.LimitReader(reader, (16<<20)+1))
			reader.Close()
			if err == nil && len(data) <= 16<<20 && !spec.readOnly {
				text := string(data)
				if !spec.multiline {
					text = strings.ReplaceAll(text, "\n", " ")
				}
				consumed := false
				if spec.onPaste != nil {
					core.Call(e.gtx, func() { consumed = spec.onPaste(core.ClipboardData{Text: text}) })
				}
				if !consumed {
					_ = documentReplaceSelection(d, text)
				}
			}
		case key.Event:
			if ev.State != key.Press {
				continue
			}
			switch ev.Name {
			case "C":
				e.inputDocumentAction(n, st, InputCopy)
			case "X":
				e.inputDocumentAction(n, st, InputCut)
			case "V":
				e.inputDocumentAction(n, st, InputPaste)
			case "Z", "Y":
				if !spec.readOnly {
					if ev.Name == "Y" || ev.Modifiers.Contain(key.ModShift) {
						d.session.Redo()
					} else {
						d.session.Undo()
					}
				}
			case key.NameEnter, key.NameReturn:
				if spec.multiline && !spec.readOnly {
					_ = documentReplaceSelection(d, "\n")
				} else if !spec.multiline && spec.onSubmit != nil {
					core.Call(e.gtx, func() { spec.onSubmit(d.Content().Text()) })
				}
			case key.NameLeftArrow, key.NameRightArrow:
				delta := 1
				if ev.Name == key.NameLeftArrow {
					delta = -1
				}
				a, b := ed.Selection()
				if a != b && !ev.Modifiers.Contain(key.ModShift) {
					if delta < 0 {
						ed.SetCaret(min(a, b), min(a, b))
					} else {
						ed.SetCaret(max(a, b), max(a, b))
					}
				} else {
					endDelta := delta
					if ev.Modifiers.Contain(key.ModShift) {
						endDelta = 0
					}
					if ev.Modifiers.Contain(key.ModShortcutAlt) {
						next := documentWordEdge(ed.Text(), a, delta)
						anchor := next
						if endDelta == 0 {
							anchor = b
						}
						ed.SetCaret(next, anchor)
					} else {
						ed.MoveCaret(delta, endDelta)
					}
					a, b = ed.Selection()
					p := d.presentation()
					if r, err := inputcontent.ByteRange(p.Text, InputRange{Start: a, End: b}); err == nil {
						if a == b {
							at := p.SourceOffset(r.Start, delta)
							_ = d.session.SelectSource(InputRange{Start: at, End: at})
						} else {
							_ = documentSelectLayout(d, r)
						}
						documentSyncEditor(st, d)
					}
				}
				documentReadSelection(st, d)
			case key.NameDeleteBackward, key.NameDeleteForward:
				if spec.readOnly {
					continue
				}
				if ed.SelectionLen() == 0 {
					delta := 1
					if ev.Name == key.NameDeleteBackward {
						delta = -1
					}
					if ev.Modifiers.Contain(key.ModShortcutAlt) {
						a, _ := ed.Selection()
						ed.SetCaret(documentWordEdge(ed.Text(), a, delta), a)
					} else {
						ed.MoveCaret(delta, 0)
					}
					documentReadSelection(st, d)
				}
				_ = documentReplaceSelection(d, "")
			}
		}
		documentSyncEditor(st, d)
	}
	// Gio still owns pointer selection, scrolling and vertical/word navigation.
	// Text-changing keys and platform edits above use explicit ranges.
	for {
		ev, ok := ed.Update(e.gtx)
		if !ok {
			break
		}
		if _, ok := ev.(widget.SelectEvent); ok {
			documentReadSelection(st, d)
		}
	}
	_, _ = e.inputPasteEvent(n, st)
	for _, action := range st.inputActions {
		e.inputDocumentAction(n, st, action)
	}
	st.inputActions = nil
	documentReadSelection(st, d)
	if d.revision == beforeRevision && !inputcontent.Equal(before, d.Content()) {
		d.revision++
		st.inputDocumentRevision = d.revision
		if spec.bind != nil {
			*spec.bind = d.Content().Text()
		}
		if spec.onChange != nil {
			core.Call(e.gtx, func() { spec.onChange(d.Content().Text()) })
		}
	}
	e.inputTokenClicks(n, st)
}

func (e *engine) inputDocumentIME(n *Node, st *elemState, g layout.Context) {
	inputTokenHitAreas(st, g)
	if !g.Enabled() || !g.Focused(&st.editor) {
		return
	}
	text, start, end := documentPlatformSelection(n.input.document)
	ca, cb := documentCompositionLayout(n.input.document, st.inputComposition.Start, st.inputComposition.End)
	bounds := editorstyle.Composition(g, &st.editor, key.Range{Start: ca, End: cb}, *n.textStyle.color)
	g.Execute(key.SelectionCmd{Tag: &st.editor, Range: key.Range{Start: start, End: end}, Caret: st.caret.InputMethodCaret(&st.editor), CompositionBounds: bounds})
	g.Execute(key.SnippetCmd{Tag: &st.editor, Snippet: key.Snippet{Range: key.Range{Start: 0, End: utf8.RuneCountInString(text)}, Text: text}})
}

// Word navigation stops at whitespace transitions, as Gio's editor does.
func documentWordEdge(text string, at, direction int) int {
	runes := []rune(text)
	at = max(0, min(len(runes), at))
	index := at
	if direction < 0 {
		index--
	}
	if index < 0 || index >= len(runes) {
		return at
	}
	space := unicode.IsSpace(runes[index])
	for index >= 0 && index < len(runes) && unicode.IsSpace(runes[index]) == space {
		at += direction
		index += direction
	}
	return at
}
