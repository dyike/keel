//go:build darwin && !ios

package sys

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/dyike/keel/native"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type darwinTray struct{ item, target id }

func TrayMenuCapabilities() TrayMenuFeatures {
	return TrayMenuFeatures{Icons: true, TemplateIcons: true, DismissAnimation: true, CancelTracking: true}
}
func (t *darwinTray) CancelMenuTracking(animated bool) error {
	mainAsync(func() {
		if t.item == 0 {
			return
		}
		selector := "cancelTrackingWithoutAnimation"
		if animated {
			selector = "cancelTracking"
		}
		send(send(t.item, "menu"), selector)
	})
	return nil
}

var darwinTrays struct {
	once            sync.Once
	class           objc.Class
	err             error
	targets         map[id]func(uint32) // AppKit main thread only
	statusItem      func(id, objc.SEL, float64) id
	imageSize       func(id, objc.SEL, struct{ Width, Height float64 })
	imageDimensions func(id, objc.SEL) struct{ Width, Height float64 }
}

func initDarwinTray() {
	darwinTrays.once.Do(func() {
		load()
		purego.RegisterFunc(&darwinTrays.statusItem, msgSendAddr)
		purego.RegisterFunc(&darwinTrays.imageSize, msgSendAddr)
		purego.RegisterFunc(&darwinTrays.imageDimensions, msgSendAddr)
		darwinTrays.targets = make(map[id]func(uint32))
		darwinTrays.class, darwinTrays.err = objc.RegisterClass("KeelTrayTarget", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
			{Cmd: sel("keelTrayAction:"), Fn: func(target id, _ objc.SEL, sender id) {
				if event := darwinTrays.targets[target]; event != nil {
					event(uint32(send(sender, "tag")))
				}
			}},
		})
	})
}

func TrayOpen(options TrayOptions) (TrayBackend, error) {
	load()
	var tray *darwinTray
	var err error
	onMain(func() {
		initDarwinTray()
		if darwinTrays.err != nil {
			err = fmt.Errorf("%w: %v", native.ErrFailed, darwinTrays.err)
			return
		}
		send(class("NSApplication"), "sharedApplication")
		bar := send(class("NSStatusBar"), "systemStatusBar")
		item := darwinTrays.statusItem(bar, sel("statusItemWithLength:"), -1)
		if item == 0 {
			err = native.ErrFailed
			return
		}
		send(item, "retain")
		target := send(send(id(darwinTrays.class), "alloc"), "init")
		tray = &darwinTray{item: item, target: target}
		darwinTrays.targets[target] = options.Event
		button := send(item, "button")
		send(button, "setTarget:", uintptr(target))
		send(button, "setAction:", uintptr(sel("keelTrayAction:")))
		send(button, "setTag:", 0)
		tray.apply(options)
	})
	if err != nil {
		return nil, err
	}
	return tray, nil
}

// Changes are queued, never synchronously waiting for AppKit while a UI callback
// might hold the frame lock. FIFO main-queue dispatch also orders Close last.
func (t *darwinTray) Update(options TrayOptions) error {
	mainAsync(func() {
		if t.item != 0 {
			t.apply(options)
		}
	})
	return nil
}

func (t *darwinTray) apply(options TrayOptions) {
	data := send(class("NSData"), "dataWithBytes:length:", uintptr(unsafe.Pointer(&options.PNG[0])), uintptr(len(options.PNG)))
	runtime.KeepAlive(options.PNG)
	image := send(send(class("NSImage"), "alloc"), "initWithData:", uintptr(data))
	if image != 0 {
		var template uintptr
		if options.Template {
			template = 1
		}
		send(image, "setTemplate:", template)
		darwinTrays.imageSize(image, sel("setSize:"), struct{ Width, Height float64 }{18, 18})
		send(send(t.item, "button"), "setImage:", uintptr(image))
		release(image)
	}
	send(send(t.item, "button"), "setToolTip:", uintptr(nsString(options.Tooltip)))
	var menu id
	if len(options.Menu) != 0 {
		menu = t.menu(options.Menu)
	}
	send(t.item, "setMenu:", uintptr(menu))
	release(menu)
}

func (t *darwinTray) menu(items []TrayMenuItem) id {
	menu := send(send(class("NSMenu"), "alloc"), "initWithTitle:", uintptr(nsString("")))
	send(menu, "setAutoenablesItems:", 0)
	for _, item := range items {
		if item.Separator {
			send(menu, "addItem:", uintptr(send(class("NSMenuItem"), "separatorItem")))
			continue
		}
		entry := send(send(class("NSMenuItem"), "alloc"), "initWithTitle:action:keyEquivalent:", uintptr(nsString(item.Title)), uintptr(sel("keelTrayAction:")), uintptr(nsString("")))
		send(entry, "setTarget:", uintptr(t.target))
		send(entry, "setTag:", uintptr(item.ID))
		var enabled, checked uintptr
		if !item.Disabled {
			enabled = 1
		}
		if item.Checked {
			checked = 1
		}
		send(entry, "setEnabled:", enabled)
		send(entry, "setState:", checked)
		if len(item.PNG) > 0 {
			data := send(class("NSData"), "dataWithBytes:length:", uintptr(unsafe.Pointer(&item.PNG[0])), uintptr(len(item.PNG)))
			runtime.KeepAlive(item.PNG)
			image := send(send(class("NSImage"), "alloc"), "initWithData:", uintptr(data))
			if image != 0 {
				var template uintptr
				if item.Template {
					template = 1
				}
				send(image, "setTemplate:", template)
				size := darwinTrays.imageDimensions(image, sel("size"))
				if side := max(size.Width, size.Height); side > 0 {
					size.Width, size.Height = size.Width*16/side, size.Height*16/side
					darwinTrays.imageSize(image, sel("setSize:"), size)
				}
				send(entry, "setImage:", uintptr(image))
				release(image)
			}
		}
		if len(item.Children) != 0 {
			child := t.menu(item.Children)
			send(entry, "setSubmenu:", uintptr(child))
			release(child)
		}
		send(menu, "addItem:", uintptr(entry))
		release(entry)
	}
	return menu
}

func (t *darwinTray) Close() error {
	mainAsync(func() {
		if t.item == 0 {
			return
		}
		delete(darwinTrays.targets, t.target)
		send(send(class("NSStatusBar"), "systemStatusBar"), "removeStatusItem:", uintptr(t.item))
		release(t.item)
		release(t.target)
		t.item, t.target = 0, 0
	})
	return nil
}

func TrayMain() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	load()
	if !isMainThread() {
		panic("tray.Main must run on the main OS thread")
	}
	withPool(func() { send(send(class("NSApplication"), "sharedApplication"), "run") })
}
