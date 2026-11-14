package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/dyike/keel/native/hotkey"
	"github.com/dyike/keel/native/screen"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

// State is accessed under the frame lock. Native calls run off that lock.
type hotkeyDemo struct {
	register              func(string, func()) (func() error, error)
	displays              func() ([]screen.Display, error)
	unregister            func() error
	closed, busy, loading bool
	generation            uint64
	hits                  int
	status, displayText   string
}

func newHotkeyDemo() *hotkeyDemo {
	return &hotkeyDemo{register: hotkey.Register, displays: screen.Displays}
}
func globalAccelerator() string {
	if runtime.GOOS == "darwin" {
		return "cmd+shift+k"
	}
	return "ctrl+shift+k"
}
func releaseHotkey(fn func() error) {
	if fn != nil {
		if err := fn(); err != nil {
			log.Printf("unregister demo hotkey: %v", err)
		}
	}
}
func (d *hotkeyDemo) close() {
	d.closed = true
	d.generation++
	fn := d.unregister
	d.unregister = nil
	go releaseHotkey(fn)
}
func (d *hotkeyDemo) toggle() {
	if d.closed || d.busy {
		return
	}
	d.busy = true
	generation := d.generation
	if fn := d.unregister; fn != nil {
		d.unregister = nil
		go func() {
			err := fn()
			core.Update(func() {
				if d.closed || d.generation != generation {
					if err != nil {
						log.Printf("unregister demo hotkey: %v", err)
					}
					return
				}
				d.busy = false
				if err != nil {
					d.unregister = fn
					d.status = err.Error()
					return
				}
				d.unregister = nil
				d.status = ""
			})
		}()
		return
	}
	d.generation++
	generation = d.generation
	register := d.register
	go func() {
		fn, err := register(globalAccelerator(), func() {
			core.UpdateIf(func() bool {
				if d.closed || d.generation != generation || d.unregister == nil {
					return false
				}
				d.hits++
				return true
			})
		})
		core.Update(func() {
			if d.closed || d.generation != generation {
				go releaseHotkey(fn)
				return
			}
			d.busy = false
			if err != nil {
				d.status = err.Error()
				return
			}
			d.unregister = fn
			d.status = ""
		})
	}()
}
func (d *hotkeyDemo) refresh() {
	if d.closed || d.loading {
		return
	}
	d.loading = true
	displays := d.displays
	go func() {
		ds, err := displays()
		core.Update(func() {
			if d.closed {
				return
			}
			d.loading = false
			if err != nil {
				d.displayText = err.Error()
				return
			}
			var lines []string
			for _, s := range ds {
				lines = append(lines, fmt.Sprintf(demoText("Display %d: %.0f × %.0f pt, primary=%t", "显示器 %d：%.0f × %.0f pt，主屏=%t"), s.ID, s.Width, s.Height, s.Primary))
			}
			d.displayText = strings.Join(lines, "\n")
			if d.displayText == "" {
				d.displayText = demoText("No displays reported", "未检测到显示器")
			}
		})
	}()
}
func (d *hotkeyDemo) Render(cx *el.Context) el.Element {
	label := demoText("Enable global shortcut", "启用全局快捷键")
	if d.unregister != nil {
		label = demoText("Disable global shortcut", "停用全局快捷键")
	}
	toggle := kit.Button(label, d.toggle)
	toggle.SetDisabled(d.busy)
	refresh := kit.Button(demoText("Refresh displays", "刷新显示器"), d.refresh)
	refresh.SetDisabled(d.loading)
	return el.Div().P(24).Gap(16).Child(
		el.Text(demoText("Global shortcuts & displays", "全局快捷键与显示器")).TextSize(22).Bold(),
		el.Text(globalAccelerator()).Bold(),
		el.Text(demoText("Enable the shortcut, switch to another app, then press it. Closing this preview unregisters it. Unsupported platforms or occupied shortcuts report an error.", "启用后切到其他应用按下快捷键。关闭此预览会注销快捷键；平台不支持或快捷键被占用时显示错误。")).TextColor(theme.Muted),
		toggle.Render(cx),
		el.Text(fmt.Sprintf(demoText("Triggered %d times", "已触发 %d 次"), d.hits)),
		el.Text(d.status).TextColor(theme.Danger),
		refresh.Render(cx), el.Text(d.displayText),
	)
}

func init() {
	registerSection("hotkey", "foundations", func() core.Widget {
		var preview *window.Window
		open := func() {
			if preview != nil && !preview.Closed() {
				preview.Raise()
				return
			}
			demo := newHotkeyDemo()
			preview = window.Open(window.Options{Title: demoText("Keel Native Capabilities", "Keel 原生能力"), Width: 640, Height: 480, Content: el.Root(demo), OnClose: func() { demo.close(); preview = nil }})
		}
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).Child(
				el.Text(demoText("Global shortcuts & displays", "全局快捷键与显示器")).TextSize(22).Bold(),
				el.Text(demoText("A global shortcut works even while another application has focus. Open the preview to enable it or query display information; opening this page alone registers nothing.", "全局快捷键在其他应用获得焦点时仍可触发。打开预览后可启用快捷键或查询显示器信息；仅浏览本页不会注册快捷键。")).TextColor(theme.Muted),
				kit.Button(demoText("Open native capabilities preview", "打开原生能力预览"), open).Render(cx),
			)
		}))
	})
}
