// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android && cgo && !purego) || freebsd || openbsd) && !nox11

package app

// Gio's Xlib calls through cgo, behind the functions os_x11.go uses;
// os_x11_purego.go loads the same libraries without cgo. Keel patch: moved
// from os_x11.go.

/*
#cgo freebsd openbsd CFLAGS: -I/usr/X11R6/include -I/usr/local/include
#cgo freebsd openbsd LDFLAGS: -L/usr/X11R6/lib -L/usr/local/lib
#cgo freebsd openbsd LDFLAGS: -lX11 -lxkbcommon -lxkbcommon-x11 -lX11-xcb -lXcursor -lXfixes
#cgo linux pkg-config: x11 xkbcommon xkbcommon-x11 x11-xcb xcursor xfixes

#include <stdlib.h>
#include <locale.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
#include <X11/Xresource.h>
#include <X11/XKBlib.h>
#include <X11/Xlib-xcb.h>
#include <X11/extensions/Xfixes.h>
#include <X11/Xcursor/Xcursor.h>
#include <xkbcommon/xkbcommon-x11.h>

*/
import "C"

import (
	"fmt"
	"unsafe"
)

// The structures and values os_x11.go defines must match the headers.
func init() {
	for _, c := range [...]struct {
		name      string
		got, want uintptr
	}{
		{"xAnyEvent", unsafe.Sizeof(xAnyEvent{}), unsafe.Sizeof(C.XAnyEvent{})},
		{"xAnyEvent.Type", unsafe.Offsetof(xAnyEvent{}.Type), unsafe.Offsetof(C.XAnyEvent{}._type)},
		{"xAnyEvent.Serial", unsafe.Offsetof(xAnyEvent{}.Serial), unsafe.Offsetof(C.XAnyEvent{}.serial)},
		{"xAnyEvent.SendEvent", unsafe.Offsetof(xAnyEvent{}.SendEvent), unsafe.Offsetof(C.XAnyEvent{}.send_event)},
		{"xAnyEvent.Display", unsafe.Offsetof(xAnyEvent{}.Display), unsafe.Offsetof(C.XAnyEvent{}.display)},
		{"xAnyEvent.Window", unsafe.Offsetof(xAnyEvent{}.Window), unsafe.Offsetof(C.XAnyEvent{}.window)},
		{"xKeyEvent", unsafe.Sizeof(xKeyEvent{}), unsafe.Sizeof(C.XKeyEvent{})},
		{"xKeyEvent.Type", unsafe.Offsetof(xKeyEvent{}.Type), unsafe.Offsetof(C.XKeyEvent{}._type)},
		{"xKeyEvent.Serial", unsafe.Offsetof(xKeyEvent{}.Serial), unsafe.Offsetof(C.XKeyEvent{}.serial)},
		{"xKeyEvent.SendEvent", unsafe.Offsetof(xKeyEvent{}.SendEvent), unsafe.Offsetof(C.XKeyEvent{}.send_event)},
		{"xKeyEvent.Display", unsafe.Offsetof(xKeyEvent{}.Display), unsafe.Offsetof(C.XKeyEvent{}.display)},
		{"xKeyEvent.Window", unsafe.Offsetof(xKeyEvent{}.Window), unsafe.Offsetof(C.XKeyEvent{}.window)},
		{"xKeyEvent.Root", unsafe.Offsetof(xKeyEvent{}.Root), unsafe.Offsetof(C.XKeyEvent{}.root)},
		{"xKeyEvent.Subwindow", unsafe.Offsetof(xKeyEvent{}.Subwindow), unsafe.Offsetof(C.XKeyEvent{}.subwindow)},
		{"xKeyEvent.Time", unsafe.Offsetof(xKeyEvent{}.Time), unsafe.Offsetof(C.XKeyEvent{}.time)},
		{"xKeyEvent.X", unsafe.Offsetof(xKeyEvent{}.X), unsafe.Offsetof(C.XKeyEvent{}.x)},
		{"xKeyEvent.Y", unsafe.Offsetof(xKeyEvent{}.Y), unsafe.Offsetof(C.XKeyEvent{}.y)},
		{"xKeyEvent.XRoot", unsafe.Offsetof(xKeyEvent{}.XRoot), unsafe.Offsetof(C.XKeyEvent{}.x_root)},
		{"xKeyEvent.YRoot", unsafe.Offsetof(xKeyEvent{}.YRoot), unsafe.Offsetof(C.XKeyEvent{}.y_root)},
		{"xKeyEvent.State", unsafe.Offsetof(xKeyEvent{}.State), unsafe.Offsetof(C.XKeyEvent{}.state)},
		{"xKeyEvent.Keycode", unsafe.Offsetof(xKeyEvent{}.Keycode), unsafe.Offsetof(C.XKeyEvent{}.keycode)},
		{"xKeyEvent.SameScreen", unsafe.Offsetof(xKeyEvent{}.SameScreen), unsafe.Offsetof(C.XKeyEvent{}.same_screen)},
		{"xButtonEvent", unsafe.Sizeof(xButtonEvent{}), unsafe.Sizeof(C.XButtonEvent{})},
		{"xButtonEvent.Type", unsafe.Offsetof(xButtonEvent{}.Type), unsafe.Offsetof(C.XButtonEvent{}._type)},
		{"xButtonEvent.Display", unsafe.Offsetof(xButtonEvent{}.Display), unsafe.Offsetof(C.XButtonEvent{}.display)},
		{"xButtonEvent.Window", unsafe.Offsetof(xButtonEvent{}.Window), unsafe.Offsetof(C.XButtonEvent{}.window)},
		{"xButtonEvent.Time", unsafe.Offsetof(xButtonEvent{}.Time), unsafe.Offsetof(C.XButtonEvent{}.time)},
		{"xButtonEvent.X", unsafe.Offsetof(xButtonEvent{}.X), unsafe.Offsetof(C.XButtonEvent{}.x)},
		{"xButtonEvent.Y", unsafe.Offsetof(xButtonEvent{}.Y), unsafe.Offsetof(C.XButtonEvent{}.y)},
		{"xButtonEvent.State", unsafe.Offsetof(xButtonEvent{}.State), unsafe.Offsetof(C.XButtonEvent{}.state)},
		{"xButtonEvent.Button", unsafe.Offsetof(xButtonEvent{}.Button), unsafe.Offsetof(C.XButtonEvent{}.button)},
		{"xButtonEvent.SameScreen", unsafe.Offsetof(xButtonEvent{}.SameScreen), unsafe.Offsetof(C.XButtonEvent{}.same_screen)},
		{"xMotionEvent", unsafe.Sizeof(xMotionEvent{}), unsafe.Sizeof(C.XMotionEvent{})},
		{"xMotionEvent.Type", unsafe.Offsetof(xMotionEvent{}.Type), unsafe.Offsetof(C.XMotionEvent{}._type)},
		{"xMotionEvent.Display", unsafe.Offsetof(xMotionEvent{}.Display), unsafe.Offsetof(C.XMotionEvent{}.display)},
		{"xMotionEvent.Time", unsafe.Offsetof(xMotionEvent{}.Time), unsafe.Offsetof(C.XMotionEvent{}.time)},
		{"xMotionEvent.X", unsafe.Offsetof(xMotionEvent{}.X), unsafe.Offsetof(C.XMotionEvent{}.x)},
		{"xMotionEvent.Y", unsafe.Offsetof(xMotionEvent{}.Y), unsafe.Offsetof(C.XMotionEvent{}.y)},
		{"xMotionEvent.State", unsafe.Offsetof(xMotionEvent{}.State), unsafe.Offsetof(C.XMotionEvent{}.state)},
		{"xMotionEvent.IsHint", unsafe.Offsetof(xMotionEvent{}.IsHint), unsafe.Offsetof(C.XMotionEvent{}.is_hint)},
		{"xMotionEvent.SameScreen", unsafe.Offsetof(xMotionEvent{}.SameScreen), unsafe.Offsetof(C.XMotionEvent{}.same_screen)},
		{"xExposeEvent", unsafe.Sizeof(xExposeEvent{}), unsafe.Sizeof(C.XExposeEvent{})},
		{"xExposeEvent.Type", unsafe.Offsetof(xExposeEvent{}.Type), unsafe.Offsetof(C.XExposeEvent{}._type)},
		{"xExposeEvent.Window", unsafe.Offsetof(xExposeEvent{}.Window), unsafe.Offsetof(C.XExposeEvent{}.window)},
		{"xExposeEvent.X", unsafe.Offsetof(xExposeEvent{}.X), unsafe.Offsetof(C.XExposeEvent{}.x)},
		{"xExposeEvent.Y", unsafe.Offsetof(xExposeEvent{}.Y), unsafe.Offsetof(C.XExposeEvent{}.y)},
		{"xExposeEvent.Width", unsafe.Offsetof(xExposeEvent{}.Width), unsafe.Offsetof(C.XExposeEvent{}.width)},
		{"xExposeEvent.Height", unsafe.Offsetof(xExposeEvent{}.Height), unsafe.Offsetof(C.XExposeEvent{}.height)},
		{"xExposeEvent.Count", unsafe.Offsetof(xExposeEvent{}.Count), unsafe.Offsetof(C.XExposeEvent{}.count)},
		{"xConfigureEvent", unsafe.Sizeof(xConfigureEvent{}), unsafe.Sizeof(C.XConfigureEvent{})},
		{"xConfigureEvent.Type", unsafe.Offsetof(xConfigureEvent{}.Type), unsafe.Offsetof(C.XConfigureEvent{}._type)},
		{"xConfigureEvent.Event", unsafe.Offsetof(xConfigureEvent{}.Event), unsafe.Offsetof(C.XConfigureEvent{}.event)},
		{"xConfigureEvent.Window", unsafe.Offsetof(xConfigureEvent{}.Window), unsafe.Offsetof(C.XConfigureEvent{}.window)},
		{"xConfigureEvent.X", unsafe.Offsetof(xConfigureEvent{}.X), unsafe.Offsetof(C.XConfigureEvent{}.x)},
		{"xConfigureEvent.Y", unsafe.Offsetof(xConfigureEvent{}.Y), unsafe.Offsetof(C.XConfigureEvent{}.y)},
		{"xConfigureEvent.Width", unsafe.Offsetof(xConfigureEvent{}.Width), unsafe.Offsetof(C.XConfigureEvent{}.width)},
		{"xConfigureEvent.Height", unsafe.Offsetof(xConfigureEvent{}.Height), unsafe.Offsetof(C.XConfigureEvent{}.height)},
		{"xConfigureEvent.BorderWidth", unsafe.Offsetof(xConfigureEvent{}.BorderWidth), unsafe.Offsetof(C.XConfigureEvent{}.border_width)},
		{"xConfigureEvent.Above", unsafe.Offsetof(xConfigureEvent{}.Above), unsafe.Offsetof(C.XConfigureEvent{}.above)},
		{"xConfigureEvent.OverrideRedirect", unsafe.Offsetof(xConfigureEvent{}.OverrideRedirect), unsafe.Offsetof(C.XConfigureEvent{}.override_redirect)},
		{"xSelectionEvent", unsafe.Sizeof(xSelectionEvent{}), unsafe.Sizeof(C.XSelectionEvent{})},
		{"xSelectionEvent.Type", unsafe.Offsetof(xSelectionEvent{}.Type), unsafe.Offsetof(C.XSelectionEvent{}._type)},
		{"xSelectionEvent.Display", unsafe.Offsetof(xSelectionEvent{}.Display), unsafe.Offsetof(C.XSelectionEvent{}.display)},
		{"xSelectionEvent.Requestor", unsafe.Offsetof(xSelectionEvent{}.Requestor), unsafe.Offsetof(C.XSelectionEvent{}.requestor)},
		{"xSelectionEvent.Selection", unsafe.Offsetof(xSelectionEvent{}.Selection), unsafe.Offsetof(C.XSelectionEvent{}.selection)},
		{"xSelectionEvent.Target", unsafe.Offsetof(xSelectionEvent{}.Target), unsafe.Offsetof(C.XSelectionEvent{}.target)},
		{"xSelectionEvent.Property", unsafe.Offsetof(xSelectionEvent{}.Property), unsafe.Offsetof(C.XSelectionEvent{}.property)},
		{"xSelectionEvent.Time", unsafe.Offsetof(xSelectionEvent{}.Time), unsafe.Offsetof(C.XSelectionEvent{}.time)},
		{"xSelectionRequestEvent", unsafe.Sizeof(xSelectionRequestEvent{}), unsafe.Sizeof(C.XSelectionRequestEvent{})},
		{"xSelectionRequestEvent.Type", unsafe.Offsetof(xSelectionRequestEvent{}.Type), unsafe.Offsetof(C.XSelectionRequestEvent{}._type)},
		{"xSelectionRequestEvent.Display", unsafe.Offsetof(xSelectionRequestEvent{}.Display), unsafe.Offsetof(C.XSelectionRequestEvent{}.display)},
		{"xSelectionRequestEvent.Owner", unsafe.Offsetof(xSelectionRequestEvent{}.Owner), unsafe.Offsetof(C.XSelectionRequestEvent{}.owner)},
		{"xSelectionRequestEvent.Requestor", unsafe.Offsetof(xSelectionRequestEvent{}.Requestor), unsafe.Offsetof(C.XSelectionRequestEvent{}.requestor)},
		{"xSelectionRequestEvent.Selection", unsafe.Offsetof(xSelectionRequestEvent{}.Selection), unsafe.Offsetof(C.XSelectionRequestEvent{}.selection)},
		{"xSelectionRequestEvent.Target", unsafe.Offsetof(xSelectionRequestEvent{}.Target), unsafe.Offsetof(C.XSelectionRequestEvent{}.target)},
		{"xSelectionRequestEvent.Property", unsafe.Offsetof(xSelectionRequestEvent{}.Property), unsafe.Offsetof(C.XSelectionRequestEvent{}.property)},
		{"xSelectionRequestEvent.Time", unsafe.Offsetof(xSelectionRequestEvent{}.Time), unsafe.Offsetof(C.XSelectionRequestEvent{}.time)},
		{"xClientMessageEvent", unsafe.Sizeof(xClientMessageEvent{}), unsafe.Sizeof(C.XClientMessageEvent{})},
		{"xClientMessageEvent.Type", unsafe.Offsetof(xClientMessageEvent{}.Type), unsafe.Offsetof(C.XClientMessageEvent{}._type)},
		{"xClientMessageEvent.Display", unsafe.Offsetof(xClientMessageEvent{}.Display), unsafe.Offsetof(C.XClientMessageEvent{}.display)},
		{"xClientMessageEvent.Window", unsafe.Offsetof(xClientMessageEvent{}.Window), unsafe.Offsetof(C.XClientMessageEvent{}.window)},
		{"xClientMessageEvent.MessageType", unsafe.Offsetof(xClientMessageEvent{}.MessageType), unsafe.Offsetof(C.XClientMessageEvent{}.message_type)},
		{"xClientMessageEvent.Format", unsafe.Offsetof(xClientMessageEvent{}.Format), unsafe.Offsetof(C.XClientMessageEvent{}.format)},
		{"xClientMessageEvent.Data", unsafe.Offsetof(xClientMessageEvent{}.Data), unsafe.Offsetof(C.XClientMessageEvent{}.data)},
		{"xkbAnyEvent", unsafe.Sizeof(xkbAnyEvent{}), unsafe.Sizeof(C.XkbAnyEvent{})},
		{"xkbAnyEvent.Type", unsafe.Offsetof(xkbAnyEvent{}.Type), unsafe.Offsetof(C.XkbAnyEvent{}._type)},
		{"xkbAnyEvent.Time", unsafe.Offsetof(xkbAnyEvent{}.Time), unsafe.Offsetof(C.XkbAnyEvent{}.time)},
		{"xkbAnyEvent.XkbType", unsafe.Offsetof(xkbAnyEvent{}.XkbType), unsafe.Offsetof(C.XkbAnyEvent{}.xkb_type)},
		{"xkbAnyEvent.Device", unsafe.Offsetof(xkbAnyEvent{}.Device), unsafe.Offsetof(C.XkbAnyEvent{}.device)},
		{"xkbStateNotifyEvent", unsafe.Sizeof(xkbStateNotifyEvent{}), unsafe.Sizeof(C.XkbStateNotifyEvent{})},
		{"xkbStateNotifyEvent.Type", unsafe.Offsetof(xkbStateNotifyEvent{}.Type), unsafe.Offsetof(C.XkbStateNotifyEvent{}._type)},
		{"xkbStateNotifyEvent.Time", unsafe.Offsetof(xkbStateNotifyEvent{}.Time), unsafe.Offsetof(C.XkbStateNotifyEvent{}.time)},
		{"xkbStateNotifyEvent.XkbType", unsafe.Offsetof(xkbStateNotifyEvent{}.XkbType), unsafe.Offsetof(C.XkbStateNotifyEvent{}.xkb_type)},
		{"xkbStateNotifyEvent.Device", unsafe.Offsetof(xkbStateNotifyEvent{}.Device), unsafe.Offsetof(C.XkbStateNotifyEvent{}.device)},
		{"xkbStateNotifyEvent.BaseGroup", unsafe.Offsetof(xkbStateNotifyEvent{}.BaseGroup), unsafe.Offsetof(C.XkbStateNotifyEvent{}.base_group)},
		{"xkbStateNotifyEvent.LatchedGroup", unsafe.Offsetof(xkbStateNotifyEvent{}.LatchedGroup), unsafe.Offsetof(C.XkbStateNotifyEvent{}.latched_group)},
		{"xkbStateNotifyEvent.LockedGroup", unsafe.Offsetof(xkbStateNotifyEvent{}.LockedGroup), unsafe.Offsetof(C.XkbStateNotifyEvent{}.locked_group)},
		{"xkbStateNotifyEvent.BaseMods", unsafe.Offsetof(xkbStateNotifyEvent{}.BaseMods), unsafe.Offsetof(C.XkbStateNotifyEvent{}.base_mods)},
		{"xkbStateNotifyEvent.LatchedMods", unsafe.Offsetof(xkbStateNotifyEvent{}.LatchedMods), unsafe.Offsetof(C.XkbStateNotifyEvent{}.latched_mods)},
		{"xkbStateNotifyEvent.LockedMods", unsafe.Offsetof(xkbStateNotifyEvent{}.LockedMods), unsafe.Offsetof(C.XkbStateNotifyEvent{}.locked_mods)},
		{"xkbStateNotifyEvent.Keycode", unsafe.Offsetof(xkbStateNotifyEvent{}.Keycode), unsafe.Offsetof(C.XkbStateNotifyEvent{}.keycode)},
		{"xSizeHints", unsafe.Sizeof(xSizeHints{}), unsafe.Sizeof(C.XSizeHints{})},
		{"xSizeHints.Flags", unsafe.Offsetof(xSizeHints{}.Flags), unsafe.Offsetof(C.XSizeHints{}.flags)},
		{"xSizeHints.MinWidth", unsafe.Offsetof(xSizeHints{}.MinWidth), unsafe.Offsetof(C.XSizeHints{}.min_width)},
		{"xSizeHints.MinHeight", unsafe.Offsetof(xSizeHints{}.MinHeight), unsafe.Offsetof(C.XSizeHints{}.min_height)},
		{"xSizeHints.MaxWidth", unsafe.Offsetof(xSizeHints{}.MaxWidth), unsafe.Offsetof(C.XSizeHints{}.max_width)},
		{"xSizeHints.MaxHeight", unsafe.Offsetof(xSizeHints{}.MaxHeight), unsafe.Offsetof(C.XSizeHints{}.max_height)},
		{"xSizeHints.WinGravity", unsafe.Offsetof(xSizeHints{}.WinGravity), unsafe.Offsetof(C.XSizeHints{}.win_gravity)},
		{"xTextProperty", unsafe.Sizeof(xTextProperty{}), unsafe.Sizeof(C.XTextProperty{})},
		{"xTextProperty.Value", unsafe.Offsetof(xTextProperty{}.Value), unsafe.Offsetof(C.XTextProperty{}.value)},
		{"xTextProperty.Encoding", unsafe.Offsetof(xTextProperty{}.Encoding), unsafe.Offsetof(C.XTextProperty{}.encoding)},
		{"xTextProperty.Format", unsafe.Offsetof(xTextProperty{}.Format), unsafe.Offsetof(C.XTextProperty{}.format)},
		{"xTextProperty.Nitems", unsafe.Offsetof(xTextProperty{}.Nitems), unsafe.Offsetof(C.XTextProperty{}.nitems)},
		{"xWindowAttributes", unsafe.Sizeof(xWindowAttributes{}), unsafe.Sizeof(C.XWindowAttributes{})},
		{"xWindowAttributes.BorderWidth", unsafe.Offsetof(xWindowAttributes{}.BorderWidth), unsafe.Offsetof(C.XWindowAttributes{}.border_width)},
		{"xWindowAttributes.Visual", unsafe.Offsetof(xWindowAttributes{}.Visual), unsafe.Offsetof(C.XWindowAttributes{}.visual)},
		{"xWindowAttributes.Root", unsafe.Offsetof(xWindowAttributes{}.Root), unsafe.Offsetof(C.XWindowAttributes{}.root)},
		{"xWindowAttributes.Screen", unsafe.Offsetof(xWindowAttributes{}.Screen), unsafe.Offsetof(C.XWindowAttributes{}.screen)},
		{"xSetWindowAttributes", unsafe.Sizeof(xSetWindowAttributes{}), unsafe.Sizeof(C.XSetWindowAttributes{})},
		{"xSetWindowAttributes.BackgroundPixmap", unsafe.Offsetof(xSetWindowAttributes{}.BackgroundPixmap), unsafe.Offsetof(C.XSetWindowAttributes{}.background_pixmap)},
		{"xSetWindowAttributes.EventMask", unsafe.Offsetof(xSetWindowAttributes{}.EventMask), unsafe.Offsetof(C.XSetWindowAttributes{}.event_mask)},
		{"xSetWindowAttributes.OverrideRedirect", unsafe.Offsetof(xSetWindowAttributes{}.OverrideRedirect), unsafe.Offsetof(C.XSetWindowAttributes{}.override_redirect)},
		{"xSetWindowAttributes.Cursor", unsafe.Offsetof(xSetWindowAttributes{}.Cursor), unsafe.Offsetof(C.XSetWindowAttributes{}.cursor)},
		{"xWMHints", unsafe.Sizeof(xWMHints{}), unsafe.Sizeof(C.XWMHints{})},
		{"xWMHints.Flags", unsafe.Offsetof(xWMHints{}.Flags), unsafe.Offsetof(C.XWMHints{}.flags)},
		{"xWMHints.Input", unsafe.Offsetof(xWMHints{}.Input), unsafe.Offsetof(C.XWMHints{}.input)},
		{"xWMHints.WindowGroup", unsafe.Offsetof(xWMHints{}.WindowGroup), unsafe.Offsetof(C.XWMHints{}.window_group)},
		{"xClassHint", unsafe.Sizeof(xClassHint{}), unsafe.Sizeof(C.XClassHint{})},
		{"xClassHint.ResName", unsafe.Offsetof(xClassHint{}.ResName), unsafe.Offsetof(C.XClassHint{}.res_name)},
		{"xClassHint.ResClass", unsafe.Offsetof(xClassHint{}.ResClass), unsafe.Offsetof(C.XClassHint{}.res_class)},
		{"xrmValue", unsafe.Sizeof(xrmValue{}), unsafe.Sizeof(C.XrmValue{})},
		{"xrmValue.Size", unsafe.Offsetof(xrmValue{}.Size), unsafe.Offsetof(C.XrmValue{}.size)},
		{"xrmValue.Addr", unsafe.Offsetof(xrmValue{}.Addr), unsafe.Offsetof(C.XrmValue{}.addr)},
		{"xEvent", unsafe.Sizeof(xEvent{}), unsafe.Sizeof(C.XEvent{})},
	} {
		if c.got != c.want {
			panic(fmt.Sprintf("x11: %s is %d, the headers say %d", c.name, c.got, c.want))
		}
	}
	for _, c := range [...]struct {
		name      string
		want, got int64
	}{
		{"False", C.False, xFalse},
		{"True", C.True, xTrue},
		{"None", C.None, xNone},
		{"CurrentTime", C.CurrentTime, xCurrentTime},
		{"NoEventMask", C.NoEventMask, xNoEventMask},
		{"KeyPress", C.KeyPress, xKeyPress},
		{"KeyRelease", C.KeyRelease, xKeyRelease},
		{"ButtonPress", C.ButtonPress, xButtonPress},
		{"ButtonRelease", C.ButtonRelease, xButtonRelease},
		{"MotionNotify", C.MotionNotify, xMotionNotify},
		{"FocusIn", C.FocusIn, xFocusIn},
		{"FocusOut", C.FocusOut, xFocusOut},
		{"Expose", C.Expose, xExpose},
		{"ConfigureNotify", C.ConfigureNotify, xConfigureNotify},
		{"SelectionRequest", C.SelectionRequest, xSelectionRequest},
		{"SelectionNotify", C.SelectionNotify, xSelectionNotify},
		{"ClientMessage", C.ClientMessage, xClientMessage},
		{"KeyPressMask", C.KeyPressMask, xKeyPressMask},
		{"KeyReleaseMask", C.KeyReleaseMask, xKeyReleaseMask},
		{"ButtonPressMask", C.ButtonPressMask, xButtonPressMask},
		{"ButtonReleaseMask", C.ButtonReleaseMask, xButtonReleaseMask},
		{"PointerMotionMask", C.PointerMotionMask, xPointerMotionMask},
		{"ExposureMask", C.ExposureMask, xExposureMask},
		{"StructureNotifyMask", C.StructureNotifyMask, xStructureNotifyMask},
		{"SubstructureNotifyMask", C.SubstructureNotifyMask, xSubstructureNotifyMask},
		{"SubstructureRedirectMask", C.SubstructureRedirectMask, xSubstructureRedirectMask},
		{"FocusChangeMask", C.FocusChangeMask, xFocusChangeMask},
		{"CWBackPixmap", C.CWBackPixmap, xCWBackPixmap},
		{"CWOverrideRedirect", C.CWOverrideRedirect, xCWOverrideRedirect},
		{"CWEventMask", C.CWEventMask, xCWEventMask},
		{"CopyFromParent", C.CopyFromParent, xCopyFromParent},
		{"InputOutput", C.InputOutput, xInputOutput},
		{"PMinSize", C.PMinSize, xPMinSize},
		{"PMaxSize", C.PMaxSize, xPMaxSize},
		{"InputHint", C.InputHint, xInputHint},
		{"PropModeReplace", C.PropModeReplace, xPropModeReplace},
		{"Button1", C.Button1, xButton1},
		{"Button2", C.Button2, xButton2},
		{"Button3", C.Button3, xButton3},
		{"Button4", C.Button4, xButton4},
		{"Button5", C.Button5, xButton5},
		{"XkbUseCoreKbd", C.XkbUseCoreKbd, xkbUseCoreKbd},
		{"XkbMajorVersion", C.XkbMajorVersion, xkbMajorVersion},
		{"XkbMinorVersion", C.XkbMinorVersion, xkbMinorVersion},
		{"XkbNewKeyboardNotify", C.XkbNewKeyboardNotify, xkbNewKeyboardNotify},
		{"XkbMapNotify", C.XkbMapNotify, xkbMapNotify},
		{"XkbStateNotify", C.XkbStateNotify, xkbStateNotify},
		{"XkbNewKeyboardNotifyMask", C.XkbNewKeyboardNotifyMask, xkbNewKeyboardNotifyMask},
		{"XkbMapNotifyMask", C.XkbMapNotifyMask, xkbMapNotifyMask},
		{"XkbStateNotifyMask", C.XkbStateNotifyMask, xkbStateNotifyMask},
	} {
		if c.got != c.want {
			panic(fmt.Sprintf("x11: %s is %d, the headers say %d", c.name, c.got, c.want))
		}
	}
}

func dpy(d unsafe.Pointer) *C.Display { return (*C.Display)(d) }

func cchar(b []byte) *C.char { return (*C.char)(unsafe.Pointer(unsafe.SliceData(b))) }

func cev(ev *xEvent) *C.XEvent { return (*C.XEvent)(unsafe.Pointer(ev)) }

func loadX11() error                 { return nil }
func xInitThreads() int32            { return int32(C.XInitThreads()) }
func xrmInitialize()                 { C.XrmInitialize() }
func xOpenDisplay() unsafe.Pointer   { return unsafe.Pointer(C.XOpenDisplay(nil)) }
func xCloseDisplay(d unsafe.Pointer) { C.XCloseDisplay(dpy(d)) }

func xkbQueryExtension(d unsafe.Pointer, major, minor int32) (int32, bool) {
	maj, min := C.int(major), C.int(minor)
	var base C.int
	ok := C.XkbQueryExtension(dpy(d), nil, &base, nil, &maj, &min) == C.True
	return int32(base), ok
}

func xkbSelectEvents(d unsafe.Pointer, device, affect, values uint32) int32 {
	return int32(C.XkbSelectEvents(dpy(d), C.uint(device), C.uint(affect), C.uint(values)))
}

func xDefaultRootWindow(d unsafe.Pointer) xWindow { return xWindow(C.XDefaultRootWindow(dpy(d))) }
func xDefaultScreen(d unsafe.Pointer) int32       { return int32(C.XDefaultScreen(dpy(d))) }
func xDisplayWidth(d unsafe.Pointer, screen int32) int32 {
	return int32(C.XDisplayWidth(dpy(d), C.int(screen)))
}
func xDisplayHeight(d unsafe.Pointer, screen int32) int32 {
	return int32(C.XDisplayHeight(dpy(d), C.int(screen)))
}

func xCreateWindow(d unsafe.Pointer, parent xWindow, x, y int32, width, height, border uint32, depth int32, class uint32, mask uintptr, attrs *xSetWindowAttributes) xWindow {
	return xWindow(C.XCreateWindow(dpy(d), C.Window(parent), C.int(x), C.int(y), C.uint(width), C.uint(height), C.uint(border), C.int(depth), C.uint(class), nil, C.ulong(mask), (*C.XSetWindowAttributes)(unsafe.Pointer(attrs))))
}

func xDestroyWindow(d unsafe.Pointer, w xWindow) { C.XDestroyWindow(dpy(d), C.Window(w)) }
func xMapWindow(d unsafe.Pointer, w xWindow)     { C.XMapWindow(dpy(d), C.Window(w)) }
func xMapRaised(d unsafe.Pointer, w xWindow)     { C.XMapRaised(dpy(d), C.Window(w)) }
func xIconifyWindow(d unsafe.Pointer, w xWindow, screen int32) {
	C.XIconifyWindow(dpy(d), C.Window(w), C.int(screen))
}
func xResizeWindow(d unsafe.Pointer, w xWindow, width, height uint32) {
	C.XResizeWindow(dpy(d), C.Window(w), C.uint(width), C.uint(height))
}
func xMoveResizeWindow(d unsafe.Pointer, w xWindow, x, y int32, width, height uint32) {
	C.XMoveResizeWindow(dpy(d), C.Window(w), C.int(x), C.int(y), C.uint(width), C.uint(height))
}
func xGetWindowAttributes(d unsafe.Pointer, w xWindow, attrs *xWindowAttributes) {
	C.XGetWindowAttributes(dpy(d), C.Window(w), (*C.XWindowAttributes)(unsafe.Pointer(attrs)))
}
func xSetWMNormalHints(d unsafe.Pointer, w xWindow, h *xSizeHints) {
	C.XSetWMNormalHints(dpy(d), C.Window(w), (*C.XSizeHints)(unsafe.Pointer(h)))
}
func xSetWMHints(d unsafe.Pointer, w xWindow, h *xWMHints) {
	C.XSetWMHints(dpy(d), C.Window(w), (*C.XWMHints)(unsafe.Pointer(h)))
}

// xSetClassHint takes NUL-terminated strings.
// The strings are copied to C memory: cgo rejects a struct of Go pointers.
func xSetClassHint(d unsafe.Pointer, w xWindow, name, class []byte) {
	cname := C.CString(string(name[:len(name)-1]))
	defer C.free(unsafe.Pointer(cname))
	cclass := C.CString(string(class[:len(class)-1]))
	defer C.free(unsafe.Pointer(cclass))
	h := C.XClassHint{cname, cclass}
	C.XSetClassHint(dpy(d), C.Window(w), &h)
}

func xSetWMProtocols(d unsafe.Pointer, w xWindow, atoms []xAtom) {
	C.XSetWMProtocols(dpy(d), C.Window(w), (*C.Atom)(unsafe.Pointer(unsafe.SliceData(atoms))), C.int(len(atoms)))
}

// xStoreName takes a NUL-terminated name.
func xStoreName(d unsafe.Pointer, w xWindow, name []byte) {
	C.XStoreName(dpy(d), C.Window(w), cchar(name))
}

func xSetTextProperty(d unsafe.Pointer, w xWindow, value []byte, encoding xAtom, format int32, prop xAtom) {
	// Copied to C memory: cgo rejects a struct holding a Go pointer.
	cvalue := C.CString(string(value))
	defer C.free(unsafe.Pointer(cvalue))
	p := C.XTextProperty{
		value:    (*C.uchar)(unsafe.Pointer(cvalue)),
		encoding: C.Atom(encoding),
		format:   C.int(format),
		nitems:   C.ulong(len(value)),
	}
	C.XSetTextProperty(dpy(d), C.Window(w), &p, C.Atom(prop))
}

func xGetTextProperty(d unsafe.Pointer, w xWindow, p *xTextProperty, prop xAtom) int32 {
	return int32(C.XGetTextProperty(dpy(d), C.Window(w), (*C.XTextProperty)(unsafe.Pointer(p)), C.Atom(prop)))
}

// xInternAtom takes a NUL-terminated name.
func xInternAtom(d unsafe.Pointer, name []byte, onlyIfExists int32) xAtom {
	return xAtom(C.XInternAtom(dpy(d), cchar(name), C.Bool(onlyIfExists)))
}

func xDeleteProperty(d unsafe.Pointer, w xWindow, prop xAtom) {
	C.XDeleteProperty(dpy(d), C.Window(w), C.Atom(prop))
}

func xConvertSelection(d unsafe.Pointer, sel, target, prop xAtom, requestor xWindow, time uintptr) {
	C.XConvertSelection(dpy(d), C.Atom(sel), C.Atom(target), C.Atom(prop), C.Window(requestor), C.Time(time))
}

func xSetSelectionOwner(d unsafe.Pointer, sel xAtom, owner xWindow, time uintptr) {
	C.XSetSelectionOwner(dpy(d), C.Atom(sel), C.Window(owner), C.Time(time))
}

func xChangeProperty(d unsafe.Pointer, w xWindow, prop, typ xAtom, format, mode int32, data unsafe.Pointer, n int32) {
	C.XChangeProperty(dpy(d), C.Window(w), C.Atom(prop), C.Atom(typ), C.int(format), C.int(mode), (*C.uchar)(data), C.int(n))
}

func xSendEvent(d unsafe.Pointer, w xWindow, propagate int32, mask int, ev *xEvent) {
	C.XSendEvent(dpy(d), C.Window(w), C.Bool(propagate), C.long(mask), cev(ev))
}

func xPending(d unsafe.Pointer) int32          { return int32(C.XPending(dpy(d))) }
func xNextEvent(d unsafe.Pointer, ev *xEvent)  { C.XNextEvent(dpy(d), cev(ev)) }
func xFilterEvent(ev *xEvent, w xWindow) int32 { return int32(C.XFilterEvent(cev(ev), C.Window(w))) }
func xConnectionNumber(d unsafe.Pointer) int32 { return int32(C.XConnectionNumber(dpy(d))) }

func xTranslateCoordinates(d unsafe.Pointer, src, dst xWindow, x, y int32) (int32, int32) {
	var dx, dy C.int
	var child C.Window
	C.XTranslateCoordinates(dpy(d), C.Window(src), C.Window(dst), C.int(x), C.int(y), &dx, &dy, &child)
	return int32(dx), int32(dy)
}

func xFixesHideCursor(d unsafe.Pointer, w xWindow) { C.XFixesHideCursor(dpy(d), C.Window(w)) }

// xcursorLibraryLoadCursor takes a NUL-terminated name.
func xcursorLibraryLoadCursor(d unsafe.Pointer, name []byte) uintptr {
	return uintptr(C.XcursorLibraryLoadCursor(dpy(d), cchar(name)))
}

func xDefineCursor(d unsafe.Pointer, w xWindow, c uintptr) {
	C.XDefineCursor(dpy(d), C.Window(w), C.Cursor(c))
}

func xDisplayString(d unsafe.Pointer) string { return C.GoString(C.XDisplayString(dpy(d))) }

func xResourceManagerString(d unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.XResourceManagerString(dpy(d)))
}

func xrmGetStringDatabase(s unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.XrmGetStringDatabase((*C.char)(s)))
}

func xrmGetResource(db unsafe.Pointer, name, class string) (string, xrmValue, bool) {
	var t *C.char
	var v C.XrmValue
	cname, cclass := cString(name), cString(class)
	if C.XrmGetResource(C.XrmDatabase(db), cchar(cname), cchar(cclass), &t, &v) == C.False || t == nil {
		return "", xrmValue{}, false
	}
	return C.GoString(t), xrmValue{Size: uint32(v.size), Addr: unsafe.Pointer(v.addr)}, true
}

func xrmDestroyDatabase(db unsafe.Pointer) { C.XrmDestroyDatabase(C.XrmDatabase(db)) }

func xGetXCBConnection(d unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.XGetXCBConnection(dpy(d)))
}

func xkbX11GetCoreKeyboardDeviceID(xcb unsafe.Pointer) int32 {
	return int32(C.xkb_x11_get_core_keyboard_device_id((*C.xcb_connection_t)(xcb)))
}

func xkbX11KeymapNewFromDevice(ctx, xcb unsafe.Pointer, dev int32) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_x11_keymap_new_from_device((*C.struct_xkb_context)(ctx), (*C.xcb_connection_t)(xcb), C.int32_t(dev), C.XKB_KEYMAP_COMPILE_NO_FLAGS))
}

func xkbX11StateNewFromDevice(km, xcb unsafe.Pointer, dev int32) unsafe.Pointer {
	return unsafe.Pointer(C.xkb_x11_state_new_from_device((*C.struct_xkb_keymap)(km), (*C.xcb_connection_t)(xcb), C.int32_t(dev)))
}

func xkbKeymapUnref(km unsafe.Pointer) { C.xkb_keymap_unref((*C.struct_xkb_keymap)(km)) }
