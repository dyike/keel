package main

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
	"testing"
)

func TestConversationLayout(t *testing.T) {
	app := newApp()
	w := window.OpenHeadless(window.Options{Width: 1100, Height: 760, Frameless: true, Content: el.Root(app)})
	defer w.Close()
	for _, name := range []string{"新建会话", "搜索会话", "发送"} {
		if _, err := w.Find(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Click("收起侧栏"); err != nil {
		t.Fatal(err)
	}
	if !app.Collapsed() {
		t.Fatal("sidebar did not collapse")
	}
	if err := w.Click("展开侧栏"); err != nil {
		t.Fatal(err)
	}
	if app.Collapsed() {
		t.Fatal("sidebar did not expand")
	}
	if err := w.Click("新建会话"); err != nil {
		t.Fatal(err)
	}
	if app.Value() != "new" {
		t.Fatal("new conversation not selected")
	}
	if err := w.Click("在这里输入消息"); err != nil {
		t.Fatal(err)
	}
	if err := w.Type("A retained message"); err != nil {
		t.Fatal(err)
	}
	if err := w.Click("发送"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Find("A retained message"); err != nil {
		t.Fatal(err)
	}
}

func TestCollapseModeConfiguration(t *testing.T) {
	app := newApp()
	w := window.OpenHeadless(window.Options{Width: 1100, Height: 760, Content: el.Root(app)})
	defer w.Close()
	if err := w.Click("收起侧栏"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Find("整理项目笔记"); err == nil {
		t.Fatal("hidden sidebar remains visible")
	}
	for _, mode := range []string{"保留图标栏", "完全隐藏"} {
		if err := w.Click("折叠方式"); err != nil {
			t.Fatal(err)
		}
		if err := w.Click(mode); err != nil {
			t.Fatal(err)
		}
		if !app.Collapsed() {
			t.Fatal("changing mode expanded sidebar")
		}
		_, err := w.Find("整理项目笔记")
		if (err == nil) != (mode == "保留图标栏") {
			t.Fatalf("mode %s: navigation lookup %v", mode, err)
		}
	}
	if err := w.Click("展开侧栏"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Find("整理项目笔记"); err != nil {
		t.Fatal(err)
	}
}
