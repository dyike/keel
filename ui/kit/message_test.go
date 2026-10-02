package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMessageUserActionsRetryReactionsAndOwnership(t *testing.T) {
	copied, retried, reacted := 0, 0, 0
	actions := []el.View{nil, Button("复制消息", func() { copied++ })}
	reactions := []MessageReaction{{Name: "赞", Count: 2}}
	msg := Message("我", el.ViewFunc(func(*el.Context) el.Element { return el.Text("订单已提交") })).User().Actions(actions...).Reactions(reactions...).OnReaction(func(i int, on bool) {
		if i != 0 {
			t.Fatal(i)
		}
		reacted++
	})
	msg.OnRetry(func() { retried++ })
	actions[1] = Button("错误的按钮", nil)
	reactions[0].Count = 999
	msg.SetState(MessageFailed, "网络中断")
	h := renderView(msg, 400, 1)
	if shown(h, "错误的按钮") || !shown(h, "赞 2") {
		t.Fatal("retained caller slices")
	}
	click(t, h, "复制消息")
	click(t, h, "赞 2")
	if copied != 1 || reacted != 1 || !shown(h, "赞 3") {
		t.Fatal("user message actions unavailable")
	}
	click(t, h, "赞 3")
	if !shown(h, "赞 2") {
		t.Fatal("reaction could not be removed")
	}
	click(t, h, "重试")
	if retried != 1 || msg.state != MessageSending || shown(h, "重试") || !shown(h, "发送中") {
		t.Fatal("retry did not transition once")
	}
	msg.SetState(MessageReady, "")
	msg.SetDisabled(true)
	h.Frame()
	click(t, h, "复制消息")
	click(t, h, "赞 2")
	if copied != 1 || reacted != 2 {
		t.Fatal("disabled actions fired")
	}
}
