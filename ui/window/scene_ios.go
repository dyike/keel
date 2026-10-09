//go:build ios

package window

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Gio v0.10's app delegate creates a window before a scene exists. When
// explicitly selected in Info.plist, move that work to the scene delegate.
// This compatibility adapter depends on Gio v0.10's Objective-C class names.
// It runs from init, as early as the cgo version did, so the class and the
// replaced method exist before UIKit finishes launching.
func init() {
	if prepareIOSScene() != 0 {
		panic("window: Gio's iOS application delegate is incompatible with KeelSceneDelegate")
	}
}

func sel(name string) objc.SEL { return objc.RegisterName(name) }

func nsString(s string) objc.ID {
	b := append([]byte(s), 0)
	return objc.ID(objc.GetClass("NSString")).Send(sel("stringWithUTF8String:"), &b[0])
}

func dict(d objc.ID, key string) objc.ID { return d.Send(sel("objectForKey:"), nsString(key)) }

func symbol(name string) uintptr {
	p, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if err != nil {
		return 0
	}
	return p
}

// Scene delegate windows by delegate object (retained). UIKit calls the
// delegate on the main thread only.
var sceneWindows sync.Map

func prepareIOSScene() int {
	if _, err := purego.Dlopen("/System/Library/Frameworks/UIKit.framework/UIKit", purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		return 0 // Not an app process; nothing to adapt.
	}
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(sel("new"))
	defer pool.Send(sel("drain"))
	info := objc.ID(objc.GetClass("NSBundle")).Send(sel("mainBundle")).Send(sel("infoDictionary"))
	configs := dict(dict(dict(info, "UIApplicationSceneManifest"), "UISceneConfigurations"), "UIWindowSceneSessionRoleApplication")
	selected := false
	if configs != 0 {
		count := int(configs.Send(sel("count")))
		for i := 0; i < count; i++ {
			name := dict(configs.Send(sel("objectAtIndex:"), uintptr(i)), "UISceneDelegateClassName")
			if name != 0 && byte(name.Send(sel("isEqualToString:"), nsString("KeelSceneDelegate"))) != 0 {
				selected = true
			}
		}
	}
	if !selected {
		return 0
	}
	delegate := objc.GetClass("_gioAppDelegate")
	launch := sel("application:didFinishLaunchingWithOptions:")
	var method uintptr
	if delegate != 0 {
		method, _, _ = purego.SyscallN(symbol("class_getInstanceMethod"), uintptr(delegate), uintptr(launch))
	}
	if method == 0 || objc.GetClass("GioViewController") == 0 {
		return 1
	}
	registerSceneDelegate()
	types, _, _ := purego.SyscallN(symbol("method_getTypeEncoding"), method)
	imp := objc.NewIMP(func(_ objc.ID, _ objc.SEL, _, _ objc.ID) bool { return true })
	purego.SyscallN(symbol("class_replaceMethod"), uintptr(delegate), uintptr(launch), uintptr(imp), types)
	return 0
}

func registerSceneDelegate() {
	var protocols []*objc.Protocol
	if p := objc.GetProtocol("UIWindowSceneDelegate"); p != nil {
		protocols = append(protocols, p)
	}
	_, err := objc.RegisterClass("KeelSceneDelegate", objc.GetClass("UIResponder"), protocols, nil, []objc.MethodDef{
		{Cmd: sel("scene:willConnectToSession:options:"), Fn: func(self objc.ID, _ objc.SEL, scene, _, _ objc.ID) {
			if byte(scene.Send(sel("isKindOfClass:"), objc.GetClass("UIWindowScene"))) == 0 {
				return
			}
			window := objc.ID(objc.GetClass("UIWindow")).Send(sel("alloc")).Send(sel("initWithWindowScene:"), scene)
			setSceneWindow(self, window)
			window.Send(sel("release"))
			// Gio registers a display link in NSRunLoop.currentMode while
			// loading its view. Run that work after UIKit starts the main
			// run loop.
			self.Send(sel("performSelector:withObject:afterDelay:"), sel("showWindow"), objc.ID(0), float64(0))
		}},
		{Cmd: sel("showWindow"), Fn: func(self objc.ID, _ objc.SEL) {
			window := sceneWindow(self)
			controller := objc.ID(objc.GetClass("GioViewController")).Send(sel("alloc")).Send(sel("init"))
			window.Send(sel("setRootViewController:"), controller)
			controller.Send(sel("release"))
			window.Send(sel("makeKeyAndVisible"))
		}},
		// The window property UIKit reads from a scene delegate.
		{Cmd: sel("window"), Fn: func(self objc.ID, _ objc.SEL) objc.ID { return sceneWindow(self) }},
		{Cmd: sel("setWindow:"), Fn: func(self objc.ID, _ objc.SEL, window objc.ID) { setSceneWindow(self, window) }},
	})
	if err != nil {
		panic("window: cannot register KeelSceneDelegate: " + err.Error())
	}
}

func sceneWindow(delegate objc.ID) objc.ID {
	if w, ok := sceneWindows.Load(delegate); ok {
		return w.(objc.ID)
	}
	return 0
}

// setSceneWindow keeps a strong reference, as the property did.
func setSceneWindow(delegate, window objc.ID) {
	if window != 0 {
		window.Send(sel("retain"))
	}
	if old := sceneWindow(delegate); old != 0 {
		old.Send(sel("release"))
	}
	if window == 0 {
		sceneWindows.Delete(delegate)
	} else {
		sceneWindows.Store(delegate, window)
	}
}
