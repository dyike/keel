package sys

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type clipboardXValue struct {
	kind   xproto.Atom
	format byte
	data   []byte
}
type clipboardXTransfer struct {
	window                            xproto.Window
	selection, target, property, incr xproto.Atom
	next                              func() (xgb.Event, error)
	fetch                             func() (*xproto.GetPropertyReply, error)
	remove                            func() error
}

func (t clipboardXTransfer) receive() (clipboardXValue, error) {
	for {
		event, err := t.next()
		if err != nil {
			return clipboardXValue{}, err
		}
		if event == nil {
			return clipboardXValue{}, clipboardFormatError("X11 connection closed")
		}
		e, ok := event.(xproto.SelectionNotifyEvent)
		if !ok || e.Requestor != t.window || e.Selection != t.selection || e.Target != t.target {
			continue
		}
		if e.Property == xproto.AtomNone {
			return clipboardXValue{}, nil
		}
		if e.Property == t.property {
			break
		}
	}
	p, err := t.fetch()
	if err != nil {
		return clipboardXValue{}, err
	}
	if p == nil || p.BytesAfter != 0 || len(p.Value) > clipboardByteLimit {
		return clipboardXValue{}, clipboardFormatError("invalid or oversized X11 property")
	}
	if p.Type != t.incr {
		if p.Type == xproto.AtomNone {
			return clipboardXValue{}, clipboardFormatError("missing X11 property")
		}
		if err = t.remove(); err != nil {
			return clipboardXValue{}, err
		}
		return clipboardXValue{p.Type, p.Format, append([]byte{}, p.Value...)}, nil
	}
	if p.Format != 32 || len(p.Value) != 4 || xgb.Get32(p.Value) > clipboardByteLimit {
		return clipboardXValue{}, clipboardFormatError("invalid or oversized INCR header")
	}
	if err = t.remove(); err != nil {
		return clipboardXValue{}, err
	}
	var out clipboardXValue
	out.data = make([]byte, 0)
	for {
		event, err := t.next()
		if err != nil {
			return clipboardXValue{}, err
		}
		if event == nil {
			return clipboardXValue{}, clipboardFormatError("X11 connection closed during INCR")
		}
		e, ok := event.(xproto.PropertyNotifyEvent)
		if !ok || e.Window != t.window || e.Atom != t.property || e.State != xproto.PropertyNewValue {
			continue
		}
		p, err = t.fetch()
		if err != nil {
			return clipboardXValue{}, err
		}
		if p == nil || p.Type == xproto.AtomNone || p.Type == t.incr || p.BytesAfter != 0 || len(p.Value) > clipboardByteLimit-len(out.data) {
			return clipboardXValue{}, clipboardFormatError("invalid or oversized INCR chunk")
		}
		if out.kind == 0 {
			out.kind, out.format = p.Type, p.Format
		}
		if out.kind != p.Type || out.format != p.Format {
			return clipboardXValue{}, clipboardFormatError("INCR format changed")
		}
		out.data = append(out.data, p.Value...)
		if err = t.remove(); err != nil {
			return clipboardXValue{}, err
		}
		if len(p.Value) == 0 {
			return out, nil
		}
	}
}
