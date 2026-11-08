//go:build !js

package ibus

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/godbus/dbus/v5"
	"image"
	"testing"
)

type latencyObject struct {
	dbus.BusObject
	call func(string, dbus.Flags, ...any) *dbus.Call
}

func (o latencyObject) Call(method string, flags dbus.Flags, args ...any) *dbus.Call {
	return o.call(method, flags, args...)
}

type earlyPreeditHandler struct{ text string }

func (h *earlyPreeditHandler) Event(text string, _ key.Range, _, _ bool) { h.text = text }

func TestPreeditBeforeProcessKeyReply(t *testing.T) {
	h := new(earlyPreeditHandler)
	ibus := &IBus{h: h, ibusConn: new(dbus.Conn)}
	ibus.inputContextObject = latencyObject{call: func(_ string, _ dbus.Flags, _ ...any) *dbus.Call {
		(*messageHandler)(ibus).DeliverSignal(inputContextIface, "UpdatePreeditText", &dbus.Signal{Body: []any{dbus.MakeVariant([]any{"IBusText", map[string]dbus.Variant{}, "ni"}), uint32(2), true}})
		return &dbus.Call{Body: []any{true}}
	}}
	if !ibus.ProcessKey('n', 42, 0) || h.text != "ni" {
		t.Fatal("first preedit was dropped before the key reply")
	}
}

func TestCursorUpdatesDoNotWaitOrRepeat(t *testing.T) {
	calls := 0
	ibus := &IBus{ibusConn: new(dbus.Conn)}
	ibus.inputContextObject = latencyObject{call: func(_ string, flags dbus.Flags, _ ...any) *dbus.Call {
		calls++
		if flags&dbus.FlagNoReplyExpected == 0 {
			t.Fatal("cursor update waits for a DBus reply on the UI thread")
		}
		return new(dbus.Call)
	}}
	ibus.SetCursorLocation(image.Rect(10, 20, 11, 40))
	ibus.SetCursorLocation(image.Rect(10, 20, 11, 40))
	ibus.SetCursorLocation(image.Rect(11, 20, 12, 40))
	if calls != 2 {
		t.Fatalf("repeated cursor position sent %d requests", calls)
	}
}
