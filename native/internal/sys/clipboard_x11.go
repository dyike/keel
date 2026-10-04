package sys

import (
	"fmt"
	"os"
	"time"

	"github.com/dyike/keel/native"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func readX11Clipboard() (raw []byte, err error) {
	deadline := time.Now().Add(5 * time.Second)
	defer func() {
		if err != nil && !time.Now().Before(deadline) {
			err = fmt.Errorf("%w: X11 clipboard", native.ErrTimeout)
		}
	}()
	c, closeConn, err := openClipboardX11(os.Getenv("DISPLAY"), deadline)
	if err != nil {
		return nil, err
	}
	defer closeConn()
	setup := xproto.Setup(c)
	if c.DefaultScreen < 0 || c.DefaultScreen >= len(setup.Roots) {
		return nil, clipboardFormatError("invalid X11 screen")
	}
	root := setup.Roots[c.DefaultScreen].Root
	window, err := xproto.NewWindowId(c)
	if err != nil {
		return nil, err
	}
	if err = xproto.CreateWindowChecked(c, 0, window, root, 0, 0, 1, 1, 0, xproto.WindowClassInputOnly, 0, xproto.CwEventMask, []uint32{xproto.EventMaskPropertyChange}).Check(); err != nil {
		return nil, err
	}
	atoms := make(map[string]xproto.Atom)
	atom := func(name string) (xproto.Atom, error) {
		if a := atoms[name]; a != 0 {
			return a, nil
		}
		reply, err := xproto.InternAtom(c, false, uint16(len(name)), name).Reply()
		if err != nil {
			return 0, err
		}
		if reply == nil {
			return 0, clipboardFormatError("X11 connection closed")
		}
		atoms[name] = reply.Atom
		return reply.Atom, nil
	}
	names := []string{"CLIPBOARD", "TARGETS", "TIMESTAMP", "INCR", "UTF8_STRING", "text/plain;charset=utf-8", "text/plain", "STRING", "text/uri-list", "x-special/gnome-copied-files", "image/png", "image/jpeg", "image/tiff", "image/bmp", "image/webp"}
	for _, name := range names {
		if _, err = atom(name); err != nil {
			return nil, err
		}
	}
	owner, err := xproto.GetSelectionOwner(c, atoms["CLIPBOARD"]).Reply()
	if err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, clipboardFormatError("X11 connection closed")
	}
	if owner.Owner == 0 {
		return []byte(`{"text":"","images":[],"files":[]}`), nil
	}
	sameOwner := func() error {
		current, err := xproto.GetSelectionOwner(c, atoms["CLIPBOARD"]).Reply()
		if err != nil {
			return err
		}
		if current == nil || current.Owner != owner.Owner {
			return clipboardFormatError("X11 clipboard owner changed")
		}
		return nil
	}
	serial := 0
	request := func(name string) (clipboardXValue, error) {
		if err := sameOwner(); err != nil {
			return clipboardXValue{}, err
		}
		property, err := atom(fmt.Sprintf("_KEEL_CLIPBOARD_SNAPSHOT_%d", serial))
		serial++
		if err != nil {
			return clipboardXValue{}, err
		}
		target := atoms[name]
		if err = xproto.ConvertSelectionChecked(c, window, atoms["CLIPBOARD"], target, property, xproto.TimeCurrentTime).Check(); err != nil {
			return clipboardXValue{}, err
		}
		transfer := clipboardXTransfer{window: window, selection: atoms["CLIPBOARD"], target: target, property: property, incr: atoms["INCR"],
			next: func() (xgb.Event, error) { event, err := c.WaitForEvent(); return event, err },
			fetch: func() (*xproto.GetPropertyReply, error) {
				return xproto.GetProperty(c, false, window, property, xproto.GetPropertyTypeAny, 0, clipboardByteLimit/4+1).Reply()
			},
			remove: func() error { return xproto.DeletePropertyChecked(c, window, property).Check() },
		}
		return transfer.receive()
	}
	value, err := request("TARGETS")
	if err != nil {
		return nil, err
	}
	var targets map[string]bool
	if value.data != nil {
		if value.kind != xproto.AtomAtom || value.format != 32 || len(value.data)%4 != 0 || len(value.data) > 4096 {
			return nil, clipboardFormatError("invalid X11 TARGETS")
		}
		offered := make(map[xproto.Atom]bool)
		for i := 0; i < len(value.data); i += 4 {
			offered[xproto.Atom(xgb.Get32(value.data[i:]))] = true
		}
		targets = make(map[string]bool)
		for _, name := range names {
			targets[name] = offered[atoms[name]]
		}
	}
	timestamp := func() ([]byte, error) {
		v, err := request("TIMESTAMP")
		if err != nil {
			return nil, err
		}
		if v.kind != xproto.AtomInteger || v.format != 32 || len(v.data) != 4 {
			return nil, clipboardFormatError("invalid X11 TIMESTAMP")
		}
		return v.data, nil
	}
	var before []byte
	if targets["TIMESTAMP"] {
		before, err = timestamp()
		if err != nil {
			return nil, err
		}
	}
	raw, err = collectXClipboard(targets, func(name string) ([]byte, error) {
		v, err := request(name)
		if err != nil || v.data == nil {
			return nil, err
		}
		validType := v.kind == atoms[name]
		if name == "text/plain" || name == "text/plain;charset=utf-8" {
			validType = validType || v.kind == atoms["UTF8_STRING"]
		}
		if !validType || v.format != 8 {
			return nil, clipboardFormatError("unexpected X11 target encoding")
		}
		return v.data, nil
	})
	if err != nil {
		return nil, err
	}
	if err = sameOwner(); err != nil {
		return nil, err
	}
	if before != nil {
		after, err := timestamp()
		if err != nil {
			return nil, err
		}
		if string(before) != string(after) {
			return nil, clipboardFormatError("X11 clipboard changed while reading")
		}
	}
	if err = sameOwner(); err != nil {
		return nil, err
	}
	return raw, nil
}
