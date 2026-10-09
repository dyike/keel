package kit

import (
	"image/color"
	"slices"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestNiceTicksAndFormat(t *testing.T) {
	for _, tc := range []struct {
		lo, hi float64
		want   []float64
	}{
		{0, 95, []float64{0, 20, 40, 60, 80, 100}}, {0, 1000, []float64{0, 200, 400, 600, 800, 1000}}, {0, 900, []float64{0, 200, 400, 600, 800, 1000}},
		{-3, 7, []float64{-5, 0, 5, 10}}, {5, 5, []float64{2, 4, 6, 8}},
	} {
		if got := niceTicks(tc.lo, tc.hi, 4); !slices.Equal(got, tc.want) {
			t.Errorf("%v..%v: %v want %v", tc.lo, tc.hi, got, tc.want)
		}
	}
	for v, want := range map[float64]string{1234567: "1,234,567", -1500.5: "-1,500.5", 0.126: "0.13", 0: "0"} {
		if got := formatNumber(v); got != want {
			t.Errorf("%v → %q want %q", v, got, want)
		}
	}
}

func TestChartHoverTooltipTableAndStack(t *testing.T) {
	c := BarChart([]string{"一月", "二月", "三月"},
		Series{Name: "线上", Values: []float64{30, 50, 40}},
		Series{Name: "门店", Values: []float64{20, 10, 60}}).Title("收入")
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(400)).Child(c.Render(cx)) })))
	h.Frame()
	if _, ok := semanticNode(h, "figure:3x2"); !ok {
		t.Fatal("figure semantics")
	}
	if lo, hi := c.span(); lo != 0 || hi != 60 {
		t.Fatalf("grouped span %v..%v", lo, hi)
	}
	c.stacked = true
	if lo, hi := c.span(); lo != 0 || hi != 100 {
		t.Fatalf("stacked span %v..%v", lo, hi)
	}
	// Hover the last band: a tooltip lists both series.
	h.Move(330, 150)
	h.Frame()
	if c.hover != 2 || !shown(h, "三月") || !shown(h, "60") {
		t.Fatalf("hover %d", c.hover)
	}
	click(t, h, "查看数据表")
	h.Frame()
	if !shown(h, "二月 | 50 | 10") {
		t.Fatal("data table")
	}
}

func TestPlotZoomPanResetAndPick(t *testing.T) {
	p := Plot(PlotSeries{Name: "点", Points: []PlotPoint{{0, 0}, {5, 5}, {10, 10}}})
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(400)).Child(p.Render(cx)) })))
	h.Frame()
	x0, x1, _, _ := p.View()
	h.Scroll(200, 150, -300) // zoom in around the pointer
	h.Frame()
	if a, b, _, _ := p.View(); b-a >= x1-x0 {
		t.Fatalf("zoom: %v..%v from %v..%v", a, b, x0, x1)
	}
	click(t, h, "复位")
	h.Key(key.NameRightArrow, 0) // focused by the click: pan right
	if a, _, _, _ := p.View(); a <= x0 {
		t.Fatal("arrow pan")
	}
	h.Key("0", 0)
	if a, _, _, _ := p.View(); a != x0 {
		t.Fatal("0 resets")
	}
	// Hover the middle point to pick it.
	h.Frame()
	_, _, y0, y1 := p.View()
	_, _ = y0, y1
	// The middle point is at the center of the fitted view: hover the plot's center.
	n, _ := semanticNode(h, "figure")
	box := n.Desc.Bounds
	px := float32(box.Min.X) + 12 + axisWidth + p.plotW/2
	py := float32(box.Min.Y) + 12 + 24 + 10 + p.height/2
	h.Move(px, py)
	h.Frame()
	if p.pick != [2]int{0, 1} || !shown(h, "x 5   y 5") {
		t.Fatalf("pick %v at %v,%v", p.pick, px, py)
	}
}

func TestColorPickerModelAndInput(t *testing.T) {
	for _, c := range []color.NRGBA{{255, 0, 0, 255}, {37, 99, 235, 255}, {128, 128, 128, 255}, {0, 0, 0, 255}} {
		h, s, v := rgbToHSV(c.R, c.G, c.B)
		if r, g, b := hsvToRGB(h, s, v); r != c.R || g != c.G || b != c.B {
			t.Errorf("round trip %v → %d %d %d", c, r, g, b)
		}
	}
	if c, ok := parseHex("#0f8"); !ok || c != (color.NRGBA{0, 0xff, 0x88, 0xff}) {
		t.Fatalf("short hex %v", c)
	}
	if _, ok := parseHex("#12345"); ok {
		t.Fatal("bad hex parsed")
	}
	var got []color.NRGBA
	p := ColorPicker().Alpha().Swatches(color.NRGBA{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff}).OnChange(func(c color.NRGBA) { got = append(got, c) })
	h := page(p)
	click(t, h, "#16A34AFF")
	if p.Value() != (color.NRGBA{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff}) || len(got) != 1 {
		t.Fatalf("swatch: %v", p.Value())
	}
	// Drag past the bottom-left corner of the square: black, hue kept.
	n, _ := node(h, "饱和度与亮度")
	b := n.Desc.Bounds
	h.Drag(float32(b.Min.X+5), float32(b.Min.Y+5), float32(b.Min.X-20), float32(b.Max.Y+20))
	if c := p.Value(); c.R != 0 || c.G != 0 || c.B != 0 || p.h == 0 {
		t.Fatalf("drag to black: %v (hue %v)", c, p.h)
	}
	clickClass(t, h, "Editor", "HEX")
	h.Key(key.NameEnd, 0)
	for range 9 {
		h.Key(key.NameDeleteBackward, 0)
	}
	h.Type("#00FF0080")
	h.Key(key.NameReturn, 0)
	if c := p.Value(); c != (color.NRGBA{G: 0xff, A: 0x80}) {
		t.Fatalf("hex: %v", c)
	}
}

func TestQuestionnaireRequiredNavigationSubmit(t *testing.T) {
	var answers map[string]Answer
	q := Questionnaire(
		Question{ID: "role", Title: "你的角色", Kind: QuestionSingle, Options: []string{"开发", "设计"}, Required: true},
		Question{ID: "tools", Title: "常用工具", Kind: QuestionMultiple, Options: []string{"Go", "Figma", "Git"}},
		Question{ID: "score", Title: "满意度", Kind: QuestionRating, Required: true},
		Question{ID: "note", Title: "建议", Kind: QuestionLongText},
	).OnSubmit(func(a map[string]Answer) { answers = a })
	h := page(q)
	if !shown(h, "第 1 / 4 题") {
		t.Fatal("progress")
	}
	click(t, h, "下一题")
	if q.Page() != 0 || !shown(h, "这一题必须回答") {
		t.Fatal("required not enforced")
	}
	click(t, h, "开发")
	h.Frame()
	if shown(h, "这一题必须回答") {
		t.Fatal("answering should clear the error")
	}
	click(t, h, "下一题")
	click(t, h, "Go")
	click(t, h, "Git")
	click(t, h, "下一题")
	click(t, h, "下一题") // rating is required
	if q.Page() != 2 {
		t.Fatal("skipped a required rating")
	}
	q.SetValue(map[string]Answer{"score": {Rating: 4}, "note": {Text: "很好"}})
	q.SetPage(3)
	h.Frame()
	click(t, h, "提交")
	if answers["role"].Text != "开发" || !slices.Equal(answers["tools"].Choices, []string{"Go", "Git"}) || answers["score"].Rating != 4 || answers["note"].Text != "很好" {
		t.Fatalf("answers %+v", answers)
	}
}

func TestColorPickerDisabledAndBlurDraft(t *testing.T) {
	calls := 0
	red := color.NRGBA{R: 255, A: 255}
	p := ColorPicker().Alpha().Swatches(red).OnChange(func(color.NRGBA) { calls++ })
	parentDisabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(el.Div().Disabled(parentDisabled).Child(p.Render(cx)), Button("disable-parent", func() { parentDisabled = true }).Render(cx), Button("outside", func() {}).Render(cx))
	}), 400, 1)
	initial := p.Value()
	p.SetDisabled(true)
	h.Frame()
	n, _ := node(h, "色相")
	x, y := center(n.Desc.Bounds)
	h.Drag(x, y, x+50, y)
	h.Key(key.NameRightArrow, 0)
	click(t, h, "#FF0000FF")
	if p.Value() != initial || calls != 0 {
		t.Fatal("disabled picker responded")
	}
	p.SetValue(red)
	h.Frame()
	if p.Value() != red || calls != 0 {
		t.Fatal("disabled programmatic color")
	}
	p.SetDisabled(false)
	h.Frame()
	edit := func() {
		clickClass(t, h, "Editor", "HEX")
		h.Key(key.NameEnd, 0)
		for range 9 {
			h.Key(key.NameDeleteBackward, 0)
		}
		h.Type("#00FF00FF")
		h.Frame()
	}
	edit()
	click(t, h, "disable-parent")
	h.Frame()
	h.Frame()
	if p.Value() != red || calls != 0 {
		t.Fatal("inherited disabled committed draft")
	}
	parentDisabled = false
	h.Frame()
	h.Frame()
	if p.Value() != red || calls != 0 {
		t.Fatal("reenable committed abandoned draft")
	}
	edit()
	click(t, h, "outside")
	h.Frame()
	h.Frame()
	if p.Value() != (color.NRGBA{G: 255, A: 255}) || calls != 1 {
		t.Fatalf("normal blur failed: %v calls %d", p.Value(), calls)
	}
}

func TestQuestionnaireDisabledAnswersNavigationAndSubmit(t *testing.T) {
	calls := 0
	q := Questionnaire(
		Question{ID: "single", Title: "single", Kind: QuestionSingle, Options: []string{"A", "B"}},
		Question{ID: "multi", Title: "multi", Kind: QuestionMultiple, Options: []string{"X", "Y"}},
		Question{ID: "text", Title: "text", Kind: QuestionText},
		Question{ID: "rating", Title: "rating", Kind: QuestionRating},
	).OnSubmit(func(map[string]Answer) { calls++ })
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return q.Render(cx) }), 400, 1)
	q.SetDisabled(true)
	h.Frame()
	n, _ := semanticNode(h, "form")
	if !n.Desc.Disabled {
		t.Fatal("disabled form missing semantics")
	}
	click(t, h, "A")
	click(t, h, "下一题")
	if q.Page() != 0 || q.Value()["single"].Text != "" {
		t.Fatal("disabled first page changed")
	}
	q.SetPage(1)
	h.Frame()
	click(t, h, "X")
	click(t, h, "上一题")
	if q.Page() != 1 || len(q.Value()["multi"].Choices) != 0 {
		t.Fatal("disabled checkbox/navigation")
	}
	q.SetPage(2)
	h.Frame()
	clickClass(t, h, "Editor", "text")
	h.Type("ignored")
	if q.Value()["text"].Text != "" {
		t.Fatal("disabled text edited")
	}
	q.SetPage(3)
	h.Frame()
	click(t, h, "提交")
	if calls != 0 {
		t.Fatal("disabled submit")
	}
	q.SetValue(map[string]Answer{"text": {Text: "program"}, "rating": {Rating: 3}})
	if q.Value()["text"].Text != "program" || q.Value()["rating"].Rating != 3 || calls != 0 {
		t.Fatal("disabled programmatic answers")
	}
	q.SetDisabled(false)
	q.SetValue(map[string]Answer{"single": {Skipped: true}, "multi": {Skipped: true}})
	h.Frame()
	click(t, h, "提交")
	if calls != 1 || q.Page() != 3 {
		t.Fatal("reenabled submit/state")
	}
}
