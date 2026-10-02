package kit

import (
	"math"
	"testing"
)

func TestAttachmentCancelRetryOpenAndDisabled(t *testing.T) {
	opened, canceled, retried, removed := 0, 0, 0, 0
	a := Attachment("file.pdf", -1).OnOpen(func() { opened++ }).OnCancel(func() { canceled++ }).OnRetry(func() { retried++ }).OnRemove(func() { removed++ })
	a.SetProgress(.4)
	h := renderView(a, 400, 1)
	click(t, h, "file.pdf")
	if opened != 0 {
		t.Fatal("opened incomplete upload")
	}
	click(t, h, "取消 file.pdf")
	if canceled != 1 || !shown(h, "已取消") {
		t.Fatal("cancel failed")
	}
	click(t, h, "重试 file.pdf")
	if retried != 1 || !shown(h, "上传中 0%") {
		t.Fatal("retry failed")
	}
	a.SetError("offline")
	h.Frame()
	click(t, h, "重试 file.pdf")
	if retried != 2 || a.err != "" {
		t.Fatal("error retry failed")
	}
	a.SetProgress(float32(math.NaN()))
	a.SetProgress(float32(math.Inf(1)))
	if a.progress != 0 {
		t.Fatal("invalid progress accepted")
	}
	a.SetProgress(-1)
	h.Frame()
	click(t, h, "file.pdf")
	if opened != 1 {
		t.Fatal("completed attachment cannot open")
	}
	click(t, h, "移除 file.pdf")
	if removed != 1 || opened != 1 {
		t.Fatal("remove bubbled to open")
	}
	a.SetDisabled(true)
	h.Frame()
	click(t, h, "file.pdf")
	click(t, h, "移除 file.pdf")
	if opened != 1 || removed != 1 {
		t.Fatal("disabled actions fired")
	}
	if FileSize(-1) != "0 B" {
		t.Fatal("negative size")
	}
}
