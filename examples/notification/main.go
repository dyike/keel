package main

import (
	"flag"
	"fmt"
	"github.com/dyike/keel/native/notification"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
	"runtime"
	"strings"
)

type demo struct {
	status   string
	events   []string
	busy     bool
	count    int
	notifier *kit.NotifierView
}

func (d *demo) record(message string) {
	d.status = message
	d.events = append(d.events, message)
	if len(d.events) > 12 {
		d.events = append([]string(nil), d.events[len(d.events)-12:]...)
	}
}

func (d *demo) done(operation string) func(error) {
	return func(err error) {
		core.Update(func() {
			d.busy = false
			if err != nil {
				d.record(operation + ": " + err.Error())
			} else {
				d.record(operation + text(": request completed", "：请求已完成"))
			}
		})
	}
}
func (d *demo) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(16).ScrollY().Child(
		el.Text(text("System notification test", "系统通知测试")).Bold(),
		el.Text(text("Request permission before sending. Reusing an ID replaces the notification. Switch to another app to check Notification Center.", "先申请权限，再投递。相同 ID 会替换。切到其他应用后观察通知中心。")),
		el.Div().Disabled(d.busy).Gap(8).Child(
			kit.Button(text("Request notification permission", "申请通知权限"), func() { d.busy = true; notification.RequestPermission(d.done(text("Permission", "权限"))) }).Render(cx),
			kit.Button(text("System notification only", "仅系统通知"), func() { d.send(kit.NoticeSystemOnly) }).Render(cx),
			kit.Button(text("In-app and system notification", "应用内和系统通知"), func() { d.send(kit.NoticeInAppAndSystem) }).Render(cx),
			kit.Button(text("Dismiss task notification", "撤回任务通知"), func() { d.notifier.DismissKey("demo") }).Render(cx),
			kit.Button(text("Clear all notifications", "清除所有通知"), func() { d.notifier.Clear() }).Render(cx),
			kit.Button(text("Native click callback", "原生点击回调"), func() {
				d.busy = true
				notification.Post(notification.Message{ID: "keel.native.click-demo", Title: text("Click this notification", "点击此通知"), Body: text("This is also shown in the foreground. Click it to display the callback result.", "前台也请求显示；点击后示例会显示回调结果。"), OnClick: func() {
					core.Update(func() {
						d.record(text("Received system notification click callback", "收到系统通知点击回调"))
					})
				}}, d.done(text("Clickable notification", "可点击通知")))
			}).Render(cx),
		),
		el.Text(d.status),
		el.Text(text("Recent events (oldest first)", "最近事件（从旧到新）")).Bold(),
		el.Text(strings.Join(d.events, "\n")),
		d.notifier.Render(cx),
	)
}
func (d *demo) send(mode kit.NoticeDelivery) {
	d.count++
	d.notifier.NotifyKey("demo", kit.Notice{Title: text("Keel Notification", "Keel 通知"), Body: fmt.Sprintf(text("Delivery %d; the in-app notice disappears after 5 seconds, while the system notification remains.", "第 %d 次投递；应用内 5 秒后消失，系统通知保留。"), d.count), Delivery: mode, OnClick: func() { d.record(text("Opened the related task", "已打开对应任务")) }, OnClose: func() { d.record(text("In-app notification closed", "应用内通知已关闭")) }})
}

type systemBackend struct{}

func (systemBackend) Post(id, title, body string, done func(error)) {
	notification.Post(notification.Message{ID: id, Title: title, Body: body}, done)
}
func (systemBackend) Remove(id string, done func(error)) { notification.Remove(id, done) }

type interactiveBackend struct{ systemBackend }

func (interactiveBackend) PostInteractive(id, title, body string, activated func(), done func(error)) {
	notification.Post(notification.Message{ID: id, Title: title, Body: body, OnClick: activated}, done)
}

func (interactiveBackend) PostActivated(id, title, body string, activated func(kit.NoticeActivation), done func(error)) {
	notification.Post(notification.Message{ID: id, Title: title, Body: body, OnActivate: func(activation notification.Activation) { activated(kit.NoticeActivation{Token: activation.Token}) }}, done)
}

func main() {
	check := flag.Bool("check", false, "print platform/bundle support without asking permission")
	language := flag.String("lang", "en", "example language: en or zh-CN")
	flag.Parse()
	applyLanguage(*language)
	if *check {
		fmt.Println(notification.Available())
		return
	}
	d := &demo{status: fmt.Sprintf(text("Available in this process: %v", "当前进程支持：%v"), notification.Available())}
	var backend kit.NoticeSystemBackend = systemBackend{}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		backend = interactiveBackend{}
	}
	d.notifier = kit.Notifier().SystemBackend(backend, func(r kit.NoticeSystemResult) {
		if r.Err != nil {
			d.record(r.Err.Error())
		} else {
			d.record(fmt.Sprintf(text("System request for notification %d completed, removing=%v", "通知 %d 系统请求完成，撤回：%v"), r.ID, r.Removing))
		}
	})
	w := window.Open(window.Options{Title: text("Keel System Notifications", "Keel 系统通知"), Width: 520, Height: 460, Content: el.Root(d)})
	// OnSystemActivation runs first and carries the token; OnSystemActivate
	// then raises the window with it.
	token := ""
	d.notifier.OnSystemActivation(func(activation kit.NoticeActivation) { token = activation.Token })
	d.notifier.OnSystemActivate(func() {
		if !w.Closed() {
			if token != "" {
				d.record(text("Requested window activation with a system token", "用系统激活令牌请求窗口置前"))
			} else {
				d.record(text("Requested window activation", "请求窗口置前"))
			}
			w.Activate(token)
			token = ""
		}
	})
	window.Main()
}
