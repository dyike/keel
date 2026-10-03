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
	status string
	busy   bool
	count  int
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
			kit.Button("投递或替换", func() {
				d.busy = true
				d.count++
				notification.Post(notification.Message{ID: "keel.notification.demo", Title: "Keel 通知", Body: fmt.Sprintf("第 %d 次投递", d.count)}, d.done("投递"))
			}).Render(cx),
			kit.Button("撤回", func() { d.busy = true; notification.Remove("keel.notification.demo", d.done("撤回")) }).Render(cx),
		),
		el.Text(d.status),
	)
}
func main() {
	check := flag.Bool("check", false, "print platform/bundle support without asking permission")
	flag.Parse()
	if *check {
		fmt.Println(notification.Available())
		return
	}
	window.Open(window.Options{Title: "Keel 系统通知", Width: 520, Height: 360, Content: el.Root(&demo{status: fmt.Sprintf("当前进程支持：%v", notification.Available())})})
	window.Main()
}
