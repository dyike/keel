package window

import (
	"testing"

	"github.com/dyike/keel/ui/kit"
)

func TestKitVizSnapshot(t *testing.T) {
	line := kit.LineChart([]string{"一月", "二月"}, kit.Series{Name: "销售", Values: []float64{1, 2}}).Title("销售额")
	bars := kit.BarChart([]string{"周一"}, kit.Series{Name: "订单", Values: []float64{3}}).Title("订单")
	plot := kit.Plot(kit.PlotSeries{Name: "点", Points: []kit.PlotPoint{{X: 1, Y: 2}}}).Title("散点")
	o := kitPage(line, bars, plot)
	o.Width, o.Height = 640, 1000
	w := openTest(t, o)
	for _, name := range []string{"销售额", "订单", "散点"} {
		if got := roleOfName(w, name); got != "figure" {
			t.Errorf("%s: %q", name, got)
		}
	}
	w.click(element(t, w, "查看数据表").center()) // equal names: the later one, the bar chart's
	element(t, w, "周一 | 3")
}

func TestKitColorPickerAndQuestionnaireSnapshot(t *testing.T) {
	cp := kit.ColorPicker()
	q := kit.Questionnaire(kit.Question{ID: "a", Title: "你的角色", Kind: kit.QuestionSingle, Options: []string{"开发"}, Required: true})
	o := kitPage(cp, q)
	o.Height = 700
	w := openTest(t, o)
	for name, role := range map[string]string{"饱和度与亮度": "slider", "色相": "slider", "HEX": "textbox", "你的角色": "group", "开发": "radio"} {
		if got := roleOfName(w, name); got != role {
			t.Errorf("%s: %q want %q", name, got, role)
		}
	}
	w.click(element(t, w, "提交").center())
	element(t, w, "这一题必须回答")
}
