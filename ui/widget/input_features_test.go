package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"reflect"
	"testing"
)

func TestInputAdornmentsClearFocusAndReadOnly(t *testing.T) {
	var changes []string
	f := Input("搜索").Prefix(Icon(IconSearch)).Suffix(Kbd("mod+k").Plain()).Clearable().OnChange(func(v string) { changes = append(changes, v) })
	h := uitest.New(f)
	// The label delegates focus and is not an extra keyboard stop.
	h.Click(10, 8)
	h.Frame()
	h.Type("你好")
	if f.Value() != "你好" {
		t.Fatal("label did not focus editor")
	}
	clickNamed(t, h, "清空 搜索")
	if f.Value() != "" || !reflect.DeepEqual(changes, []string{"你好", ""}) {
		t.Fatalf("clear callback mismatch: %q %v", f.Value(), changes)
	}
	h.Frame()
	h.Type("再试")
	if f.Value() != "再试" {
		t.Fatal("clear did not return focus")
	}
	f.SetReadOnly(true)
	h.Frame()
	clickNamed(t, h, "清空 搜索")
	h.Type("x")
	if f.Value() != "再试" {
		t.Fatal("read-only field cleared or edited")
	}
	f.SetReadOnly(false)
	f.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "清空 搜索")
	h.Type("y")
	if f.Value() != "再试" {
		t.Fatal("disabled field changed")
	}
}

func TestFieldDisabledFocusAndRecovery(t *testing.T) {
	for _, makeField := range []struct {
		name string
		new  func(string) *Field
	}{{"Input", Input}, {"TextArea", TextArea}} {
		t.Run(makeField.name, func(t *testing.T) {
			changes, submits := 0, 0
			f := makeField.new("名称").OnChange(func(string) { changes++ }).OnSubmit(func(string) { submits++ })
			f.SetValue("原文")
			h := uitest.NewFunc(func(gtx core.C) {
				if f.disabled {
					gtx.Execute(key.FocusCmd{Tag: &f.editor})
				}
				f.Layout(gtx)
			})
			if changes != 0 {
				t.Fatal("setter notified")
			}
			clickNamed(t, h, "名称")
			f.SetDisabled(true)
			h.Frame()
			clickNamed(t, h, "名称")
			h.Type("禁止")
			h.Key(key.NameDeleteBackward, 0)
			h.Key(key.NameReturn, 0)
			if f.Value() != "原文" || changes != 0 || submits != 0 || h.Router.Source().Focused(&f.editor) {
				t.Fatal("disabled field edited, submitted or focused")
			}
			f.SetValue("")
			f.SetDisabled(false)
			h.Frame()
			clickNamed(t, h, "名称")
			h.Type("恢复")
			if f.Value() != "恢复" || changes != 1 {
				t.Fatalf("field did not recover: %q, callbacks %d", f.Value(), changes)
			}
		})
	}
}
