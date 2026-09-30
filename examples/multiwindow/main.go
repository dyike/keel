package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
	"github.com/dyike/keel/window"
)

func main() {
	kit := keel.NewWithOptions(keel.Options{Name: "Keel multi-window"})
	name := ui.Input(ui.InputOptions{Label: "你的名字", Placeholder: "例如：小明"})
	result := ui.Text("主窗口和设置窗口拥有各自的状态。")
	shortcutHint := ui.Text("⌘ + , 打开设置，也可点击按钮。")
	var settings *window.Window
	var count int
	var bindSettingsShortcut func(*window.Window)

	// This function is called only on the UI thread, by both the button and shortcut.
	openSettings := func() {
		if settings != nil {
			host := settings.Host()
			// Creation is pending: repeated shortcut presses must not create duplicates.
			if host == nil {
				return
			}
			if host.Ctx().Err() == nil && !host.CloseRequested() {
				if err := settings.Show(); err != nil {
					log.Print(err)
				}
				return
			}
		}
		count++
		var err error
		settings, err = kit.Window.New(keel.WindowOptions{Title: "设置", Width: 460, Height: 360, UI: settingsPage(count)})
		if err != nil {
			log.Print(err)
			return
		}
		bindSettingsShortcut(settings)
	}
	bindSettingsShortcut = func(win *window.Window) {
		if err := win.OnReady(func() {
			if _, err := win.RegisterShortcut("super+comma", openSettings); err != nil {
				log.Print(err)
			}
		}); err != nil {
			log.Print(err)
		}
	}
	openButton := ui.Button("打开设置窗口", openSettings)

	page := ui.NewPage("主窗口", ui.Column(
		ui.Heading("Keel · Go-Gui"),
		ui.Card(name, ui.Row(ui.Button("打招呼", func() {
			value := strings.TrimSpace(name.Value())
			if value == "" {
				result.SetText("请先输入名字。")
				return
			}
			result.SetText("你好，" + value + "！")
		}), openButton), result, shortcutHint),
	))
	mainWindow, err := kit.Window.New(keel.WindowOptions{Title: "Keel 主窗口", Width: 720, Height: 480, UI: page})
	if err != nil {
		log.Fatal(err)
	}

	// Cmd+, is an application shortcut: it runs directly on the UI thread.
	if _, err := mainWindow.RegisterShortcut("super+comma", openSettings); err != nil {
		log.Fatal(err)
	}

	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
}

func settingsPage(number int) *ui.Page {
	enabled := ui.Checkbox(ui.CheckboxOptions{Label: "启用提示", Checked: true})
	note := ui.Input(ui.InputOptions{Label: "设置备注", Value: fmt.Sprintf("设置窗口 #%d", number)})
	status := ui.Text("设置尚未保存。")
	return ui.NewPage("设置", ui.Column(ui.Heading("设置"), enabled, note,
		ui.Button("保存", func() { status.SetText(fmt.Sprintf("已保存：%s，提示=%t", note.Value(), enabled.Value())) }), status))
}
