package window

import (
	"github.com/dyike/keel/ui/widget"
	"testing"
)

func TestAutomationCheckboxMixedState(t *testing.T) {
	check := widget.Checkbox("全部项目", false)
	check.SetIndeterminate(true)
	check.SetDisabled(true)
	w := openTest(t, Options{Content: check})
	e := element(t, w, "全部项目")
	if e.Role != "checkbox" || e.Value != "mixed" || !e.Disabled {
		t.Fatalf("missing mixed/disabled state: %+v", e)
	}
}
