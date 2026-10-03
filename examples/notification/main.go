package main

import (
	"flag"
	"fmt"
	"github.com/dyike/keel/native/notification"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
)

type demo struct {
	status   string
	busy     bool
	count    int
	notifier *kit.NotifierView
}

func (d *demo) done(operation string) func(error) {
	return func(err error) {
		core.Update(func() {
			d.busy = false
			if err != nil {
				d.status = operation + ": " + err.Error()
			} else {
				d.status = operation + "：请求已完成"
			}
		})
	}
}
func (d *demo) Render(cx *el.Context) el.Element {
	return el.Div().P(24).Gap(16).Child(
		el.Text("系统通知测试").Bold(),
		el.Text("先申请权限，再投递。相同 ID 会替换。切到其他应用后观察通知中心。"),
		el.Div().Disabled(d.busy).Gap(8).Child(
			kit.Button("申请通知权限", func() { d.busy = true; notification.RequestPermission(d.done("权限")) }).Render(cx),
			kit.Button("仅系统通知", func() { d.send(kit.NoticeSystemOnly) }).Render(cx),
			kit.Button("应用内和系统通知", func() { d.send(kit.NoticeInAppAndSystem) }).Render(cx),
			kit.Button("撤回任务通知", func() { d.notifier.DismissKey("demo") }).Render(cx),
			kit.Button("清除所有通知", func() { d.notifier.Clear() }).Render(cx),
		),
		el.Text(d.status),
		d.notifier.Render(cx),
	)
}
func (d *demo) send(mode kit.NoticeDelivery) {
	d.count++
	d.notifier.NotifyKey("demo", kit.Notice{Title: "Keel 通知", Body: fmt.Sprintf("第 %d 次投递；应用内 5 秒后消失，系统通知保留。", d.count), Delivery: mode})
}

type systemBackend struct{}

func (systemBackend) Post(id, title, body string, done func(error)) {
	notification.Post(notification.Message{ID: id, Title: title, Body: body}, done)
}
func (systemBackend) Remove(id string, done func(error)) { notification.Remove(id, done) }

func main() {
	check := flag.Bool("check", false, "print platform/bundle support without asking permission")
	flag.Parse()
	if *check {
		fmt.Println(notification.Available())
		return
	}
	d := &demo{status: fmt.Sprintf("当前进程支持：%v", notification.Available())}
	d.notifier = kit.Notifier().SystemBackend(systemBackend{}, func(r kit.NoticeSystemResult) {
		if r.Err != nil {
			d.status = r.Err.Error()
		} else {
			d.status = fmt.Sprintf("通知 %d 系统请求完成，撤回：%v", r.ID, r.Removing)
		}
	})
	window.Open(window.Options{Title: "Keel 系统通知", Width: 520, Height: 460, Content: el.Root(d)})
	window.Main()
}
