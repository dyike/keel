package kit

import (
	"io"
	"strings"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/clipboard"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/ui/core"
)

func (v *TerminalView) send(gtx core.C, data string) {
	if v.disabled || data == "" || v.onData == nil {
		return
	}
	v.follow = true
	fn := v.onData
	core.Call(gtx, func() {
		if !v.disabled {
			fn(data)
		}
	})
}
func (v *TerminalView) point(p f32.Point) codePos {
	row := min(max(0, int(p.Y+v.scrollY)/max(1, v.painter.metrics.lh)), v.lineCount()-1)
	line := v.line(row)
	col := min(max(0, int(p.X)/max(1, v.painter.metrics.space)), len(line))
	if col < len(line) && line[col].width == 0 && col > 0 {
		col--
	}
	return codePos{row, col}
}
func (v *TerminalView) updateInput(gtx core.C, maxScroll float32) {
	if !gtx.Enabled() {
		return
	}
	if d := v.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, 1, pointer.ScrollRange{}, pointer.ScrollRange{Min: -int(v.scrollY), Max: int(maxScroll - v.scrollY)}); d != 0 {
		v.scrollY = min(max(v.scrollY+float32(d), 0), maxScroll)
		v.follow = v.scrollY >= maxScroll
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: v, Kinds: pointer.Press | pointer.Move | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			break
		}
		e := ev.(pointer.Event)
		switch e.Kind {
		case pointer.Press:
			if e.Source == pointer.Touch {
				v.touch = true
				v.touchStart = e.Position
				v.touchID = e.PointerID
				continue
			}
			if e.Source != pointer.Mouse || !e.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			gtx.Execute(key.FocusCmd{Tag: v})
			v.tabExit = false
			v.anchor, v.end = v.point(e.Position), v.point(e.Position)
			v.dragging = true
			v.selected = false
		case pointer.Move, pointer.Drag:
			if e.Source == pointer.Touch && e.PointerID == v.touchID {
				d := e.Position.Sub(v.touchStart)
				if d.X*d.X+d.Y*d.Y > float32(gtx.Dp(8)*gtx.Dp(8)) {
					v.touch = false
				}
				continue
			}
			if v.dragging {
				v.end = v.point(e.Position)
				v.selected = v.anchor != v.end
			}
		case pointer.Release, pointer.Cancel:
			if e.Source == pointer.Touch && e.PointerID == v.touchID {
				if v.touch && e.Kind == pointer.Release {
					gtx.Execute(key.FocusCmd{Tag: v})
					gtx.Execute(key.SoftKeyboardCmd{Show: true})
				}
				v.touch = false
			}
			v.dragging = false
		}
	}
	filters := []event.Filter{
		key.FocusFilter{Target: v},
		key.Filter{Focus: v, Optional: key.ModCtrl | key.ModShift | key.ModAlt | key.ModSuper},
		transfer.TargetFilter{Target: v, Type: "application/text"},
	}
	if !v.tabExit {
		filters = append(filters, key.Filter{Focus: v, Name: key.NameTab, Optional: key.ModShift})
	}
	edited := false
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		switch e := ev.(type) {
		case key.FocusEvent:
			if e.Focus {
				v.imeText = "\x00"
				v.imePos = codePos{-1, -1}
				gtx.Execute(key.SoftKeyboardCmd{Show: true})
			} else {
				v.draft = nil
				v.draftSelection = key.Range{}
				v.composition = key.Range{Start: -1, End: -1}
				v.pastePending = false
				v.tabExit = false
			}
		case key.SnippetEvent:
			v.imeText = "\x00"
		case key.SelectionEvent:
			v.draftSelection = key.Range{Start: min(max(e.Start, 0), len(v.draft)), End: min(max(e.End, 0), len(v.draft))}
		case key.CompositionEvent:
			v.composition = key.Range(e)
		case key.EditEvent:
			a, z := min(e.Range.Start, e.Range.End), max(e.Range.Start, e.Range.End)
			a, z = min(max(a, 0), len(v.draft)), min(max(z, 0), len(v.draft))
			if len(v.draft)+len(e.Text) > 1<<20 {
				continue
			}
			next := append([]rune(nil), v.draft[:a]...)
			next = append(next, []rune(e.Text)...)
			next = append(next, v.draft[z:]...)
			v.draft = next
			at := a + len([]rune(e.Text))
			v.draftSelection = key.Range{Start: at, End: at}
			edited = true
			v.tabExit = false
		case key.Event:
			if e.State != key.Press {
				continue
			}
			copyKey := e.Name == "C" && (e.Modifiers.Contain(key.ModCtrl|key.ModShift) || e.Modifiers.Contain(key.ModSuper))
			pasteKey := e.Name == "V" && (e.Modifiers.Contain(key.ModCtrl|key.ModShift) || e.Modifiers.Contain(key.ModSuper))
			if copyKey {
				if text := v.Selection(); text != "" {
					gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
				}
				continue
			}
			if pasteKey {
				v.pastePending = true
				gtx.Execute(clipboard.ReadCmd{Tag: v})
				continue
			}
			if v.composition.Start >= 0 && v.composition.End > v.composition.Start {
				continue
			}
			v.tabExit = e.Name == key.NameEscape
			v.send(gtx, terminalKey(e, v.model.applicationCursor))
		case transfer.DataEvent:
			reader := e.Open()
			data, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
			reader.Close()
			if v.pastePending && err == nil && len(data) <= 1<<20 {
				s := string(data)
				if v.model.bracketedPaste {
					s = "\x1b[200~" + s + "\x1b[201~"
				}
				v.send(gtx, s)
			}
			v.pastePending = false
		}
	}
	if (edited || len(v.draft) > 0) && (v.composition.Start < 0 || v.composition.Start == v.composition.End) {
		v.send(gtx, string(v.draft))
		v.draft = nil
		v.draftSelection = key.Range{}
	}
}
func terminalKey(e key.Event, application bool) string {
	if e.Modifiers.Contain(key.ModCtrl) {
		if len(e.Name) == 1 {
			b := byte(e.Name[0])
			if b >= 'A' && b <= 'Z' || b >= '[' && b <= '_' {
				return string([]byte{b & 31})
			}
		}
		if e.Name == key.NameSpace {
			return "\x00"
		}
	}
	prefix := "\x1b["
	if application {
		prefix = "\x1bO"
	}
	switch e.Name {
	case key.NameReturn, key.NameEnter:
		return "\r"
	case key.NameDeleteBackward:
		return "\x7f"
	case key.NameTab:
		if e.Modifiers.Contain(key.ModShift) {
			return "\x1b[Z"
		}
		return "\t"
	case key.NameEscape:
		return "\x1b"
	case key.NameUpArrow:
		return prefix + "A"
	case key.NameDownArrow:
		return prefix + "B"
	case key.NameRightArrow:
		return prefix + "C"
	case key.NameLeftArrow:
		return prefix + "D"
	case key.NameHome:
		return prefix + "H"
	case key.NameEnd:
		return prefix + "F"
	case key.NameDeleteForward:
		return "\x1b[3~"
	case key.NamePageUp:
		return "\x1b[5~"
	case key.NamePageDown:
		return "\x1b[6~"
	case key.NameF1:
		return "\x1bOP"
	case key.NameF2:
		return "\x1bOQ"
	case key.NameF3:
		return "\x1bOR"
	case key.NameF4:
		return "\x1bOS"
	case key.NameF5:
		return "\x1b[15~"
	case key.NameF6:
		return "\x1b[17~"
	case key.NameF7:
		return "\x1b[18~"
	case key.NameF8:
		return "\x1b[19~"
	case key.NameF9:
		return "\x1b[20~"
	case key.NameF10:
		return "\x1b[21~"
	case key.NameF11:
		return "\x1b[23~"
	case key.NameF12:
		return "\x1b[24~"
	}
	return ""
}
