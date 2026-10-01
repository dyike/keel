package locale

import (
	"reflect"
	"testing"

	"github.com/dyike/keel/ui/internal/loop"
)

// Every preset fills every field, so a missing translation cannot show blank.
func TestPresetsAreComplete(t *testing.T) {
	for _, s := range []Strings{Chinese(), English()} {
		v := reflect.ValueOf(s)
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if f.Kind() == reflect.String && f.String() == "" || f.Kind() == reflect.Func && f.IsNil() {
				t.Errorf("%s: %s is empty", s.Lang, v.Type().Field(i).Name)
			}
		}
	}
	if got := English().Rows(1) + "/" + English().Rows(2) + "/" + Chinese().Rows(36); got != "1 row/2 rows/36 行" {
		t.Fatal(got)
	}
	if Chinese().Name("关闭", "保存成功") != "关闭 保存成功" || English().Name("Close", "") != "Close" {
		t.Fatal("Name")
	}
}

func TestApplyRedrawsAndBumpsRevision(t *testing.T) {
	loop.Lock()
	defer loop.Unlock()
	defer Apply(Chinese())
	redraws := 0
	tag := new(int)
	loop.Register(tag, func() { redraws++ })
	defer loop.Unregister(tag)
	before := Revision()
	en := English()
	en.Rows = nil
	Apply(en)
	if Current().Close != "Close" || Revision() != before+1 || redraws != 1 {
		t.Fatal("Apply did not switch, bump revision and redraw")
	}
	if Current().Rows(3) != "3 行" {
		t.Fatal("nil Rows should keep the current one")
	}
}
