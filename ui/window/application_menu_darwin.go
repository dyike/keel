//go:build darwin && !ios

package window

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/ebitengine/purego/objc"
)

func platformNativeApplicationMenu() bool { return true }
func platformInstallApplicationMenu(data []byte) {
	var model menuWire
	if json.Unmarshal(data, &model) != nil {
		return
	}
	appkit.MainAsync(func() { installAppKitMenu(model) })
}
func platformMenuEdit(action MenuAction) bool {
	appkit.MainAsync(func() { appKitEdit(string(action)) })
	return true
}

func platformDrawApplicationMenu(*Window) bool { return false }
func applicationMenuWindowEvent(*Window, any)  {}

func platformNativeMenuShortcuts() bool { return true }

const (
	nsEventTypeKeyDown         = 10
	nsEventModifierFlagShift   = 1 << 17
	nsEventModifierFlagControl = 1 << 18
	nsEventModifierFlagOption  = 1 << 19
	nsEventModifierFlagCommand = 1 << 20
	nsF1FunctionKey            = 0xF704
	nsControlStateValueOn      = 1
	nsControlStateValueOff     = 0
	menuModifierCommand        = 1
	menuModifierControl        = 2
	menuModifierOption         = 4
	menuModifierShift          = 8
)

var editKeys = map[string]string{"copy": "c", "cut": "x", "paste": "v", "select-all": "a", "undo": "z", "redo": "z"}

// appKitEdit sends the standard shortcut of an editing action to the key
// window's first responder. Main thread.
func appKitEdit(action string) {
	key, ok := editKeys[action]
	if !ok {
		return
	}
	window := appkit.Send(appkit.App(), "keyWindow")
	responder := appkit.Send(window, "firstResponder")
	if responder == 0 || !appkit.SendBool(responder, "respondsToSelector:", uintptr(appkit.Sel("keyDown:"))) {
		return
	}
	flags := uint(nsEventModifierFlagCommand)
	if action == "redo" {
		flags |= nsEventModifierFlagShift
	}
	uptime := appkit.MsgFloat(appkit.Send(appkit.Class("NSProcessInfo"), "processInfo"), appkit.Sel("systemUptime"))
	chars := appkit.String(key)
	event := appkit.MsgKeyEvent(appkit.Class("NSEvent"), appkit.Sel("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"),
		nsEventTypeKeyDown, appkit.Point{}, flags, uptime, int(appkit.Send(window, "windowNumber")), 0, chars, chars, false, 0)
	appkit.Send(responder, "keyDown:", uintptr(event))
}

// Menu targets by object, with the generation of the model that built them.
// Main thread only.
var menuTargets = map[appkit.ID]uint64{}
var currentMenuTarget appkit.ID

var menuActionsClass = sync.OnceValue(func() appkit.ID {
	return appkit.RegisterClass("KeelMenuActions", "NSObject", nil, []objc.MethodDef{
		appkit.Method("run:", func(self appkit.ID, _ objc.SEL, item appkit.ID) {
			dispatchNativeMenu(menuTargets[self], appkit.GoString(appkit.Send(item, "representedObject")))
		}),
		appkit.Method("edit:", func(_ appkit.ID, _ objc.SEL, item appkit.ID) {
			appKitEdit(appkit.GoString(appkit.Send(item, "representedObject")))
		}),
	})
})

var menuKeys = map[string]string{"⏎": "\r", "⌤": "\r", "⎋": "\033", "space": " ", "tab": "\t", "⌫": "\177", "⌦": "",
	"↑": "", "↓": "", "←": "", "→": "", "⇱": "", "⇲": "", "⇞": "", "⇟": ""}

// keyEquivalent maps a wire key name to AppKit's key equivalent.
func keyEquivalent(name string) string {
	if k, ok := menuKeys[name]; ok {
		return k
	}
	if rest, ok := strings.CutPrefix(name, "f"); ok && rest != "" {
		if n, err := strconv.Atoi(rest); err == nil && n >= 1 && n <= 35 {
			return string(rune(nsF1FunctionKey + n - 1))
		}
	}
	return name
}

// Main thread.
func installAppKitMenu(model menuWire) {
	app := appkit.App()
	if app == 0 {
		return
	}
	target := appkit.Send(appkit.Send(menuActionsClass(), "alloc"), "init")
	menuTargets[target] = model.Generation
	appkit.Send(app, "setWindowsMenu:", 0)
	appkit.Send(app, "setHelpMenu:", 0)
	appkit.Send(app, "setServicesMenu:", 0)
	bar := buildAppKitMenu(app, "", model.Items, target)
	appkit.Send(app, "setMainMenu:", uintptr(bar))
	appkit.Release(bar)
	if currentMenuTarget != 0 {
		delete(menuTargets, currentMenuTarget)
		appkit.Release(currentMenuTarget)
	}
	currentMenuTarget = target
}

// buildAppKitMenu returns an owned NSMenu.
func buildAppKitMenu(app appkit.ID, title string, items []menuWireItem, target appkit.ID) appkit.ID {
	menu := appkit.Send(appkit.Send(appkit.Class("NSMenu"), "alloc"), "initWithTitle:", uintptr(appkit.String(title)))
	appkit.Send(menu, "setAutoenablesItems:", 0)
	for _, entry := range items {
		if entry.Separator {
			appkit.Send(menu, "addItem:", uintptr(appkit.Send(appkit.Class("NSMenuItem"), "separatorItem")))
			continue
		}
		action, represented := appkit.Sel("run:"), entry.ID
		if entry.Action != "" {
			action, represented = appkit.Sel("edit:"), string(entry.Action)
		}
		item := appkit.Send(menu, "addItemWithTitle:action:keyEquivalent:", uintptr(appkit.String(entry.Title)), uintptr(action), uintptr(appkit.String(keyEquivalent(entry.Key))))
		appkit.Send(item, "setTarget:", uintptr(target))
		appkit.Send(item, "setRepresentedObject:", uintptr(appkit.String(represented)))
		var flags uintptr
		for bit, flag := range map[uint32]uintptr{menuModifierCommand: nsEventModifierFlagCommand, menuModifierControl: nsEventModifierFlagControl,
			menuModifierOption: nsEventModifierFlagOption, menuModifierShift: nsEventModifierFlagShift} {
			if entry.Modifiers&bit != 0 {
				flags |= flag
			}
		}
		appkit.Send(item, "setKeyEquivalentModifierMask:", flags)
		appkit.Send(item, "setEnabled:", appkit.Bool(!entry.Disabled))
		state := uintptr(nsControlStateValueOff)
		if entry.Checked {
			state = nsControlStateValueOn
		}
		appkit.Send(item, "setState:", state)
		if len(entry.Children) > 0 {
			sub := buildAppKitMenu(app, entry.Title, entry.Children, target)
			appkit.Send(item, "setSubmenu:", uintptr(sub))
			switch entry.Role {
			case MenuWindow:
				appkit.Send(app, "setWindowsMenu:", uintptr(sub))
			case MenuHelp:
				appkit.Send(app, "setHelpMenu:", uintptr(sub))
			case MenuServices:
				appkit.Send(app, "setServicesMenu:", uintptr(sub))
			}
			appkit.Release(sub)
		}
	}
	return menu
}
