package kit

import (
	"strconv"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func countRole(h *uitest.Harness, prefix string) int {
	n := 0
	var walk func(input.SemanticNode)
	walk = func(s input.SemanticNode) {
		if d := s.Desc.Description; len(d) >= len(prefix) && d[:len(prefix)] == prefix {
			n++
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	for _, s := range h.Router.AppendSemantics(nil) {
		walk(s)
	}
	return n
}

func TestVirtualListBuildsOnlyVisibleRows(t *testing.T) {
	built := 0
	vl := VirtualList(100000, 20, func(cx *el.Context, i int) el.Element {
		built++
		return el.Text("第 " + strconv.Itoa(i) + " 行").Name("row " + strconv.Itoa(i))
	}).Height(200)
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().Child(vl.Render(cx)) })))
	h.Frame()
	built = 0
	h.Frame()
	if built == 0 || built > 40 {
		t.Fatalf("built %d rows for a 10-row viewport", built)
	}
	if !shown(h, "row 0") || shown(h, "row 50") {
		t.Fatal("wrong rows shown")
	}
	for range 10 {
		h.Scroll(100, 100, 2000)
	}
	h.Frame()
	if shown(h, "row 0") {
		t.Fatal("did not scroll")
	}
}

func TestListKeyboardSelectionAndActivate(t *testing.T) {
	items := make([]string, 200)
	for i := range items {
		items[i] = "项目 " + strconv.Itoa(i)
	}
	activated := -1
	l := List(items...).Height(160).OnActivate(func(i int) { activated = i })
	h := page(l)
	click(t, h, "项目 1")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NamePageDown, 0)
	h.Frame()
	if l.Value() != 12 {
		t.Fatalf("selected %d", l.Value())
	}
	if !shown(h, "项目 12") {
		t.Fatal("selection not scrolled into view")
	}
	h.Key(key.NameEnd, 0)
	h.Frame()
	if l.Value() != 199 || !shown(h, "项目 199") {
		t.Fatalf("End: %d", l.Value())
	}
	h.Key(key.NameReturn, 0)
	if activated != 199 {
		t.Fatal("Enter did not activate")
	}
}

func TestTreeExpandCollapseAndKeys(t *testing.T) {
	root := &TreeNode{ID: "src", Label: "src", Children: []*TreeNode{
		{ID: "ui", Label: "ui", Children: []*TreeNode{{ID: "kit", Label: "kit"}}},
		{ID: "main", Label: "main.go"},
	}}
	tr := Tree(root, &TreeNode{ID: "go.mod", Label: "go.mod"}).Height(200)
	h := page(tr)
	click(t, h, "src")
	h.Key(key.NameRightArrow, 0) // expand
	h.Key(key.NameRightArrow, 0) // into ui
	h.Key(key.NameRightArrow, 0) // expand ui
	h.Key(key.NameRightArrow, 0) // into kit
	h.Frame()
	if tr.Value() != "kit" {
		t.Fatalf("selected %q", tr.Value())
	}
	if desc(h, "ui") != "treeitem:expanded" {
		t.Fatalf("ui state %q", desc(h, "ui"))
	}
	h.Key(key.NameLeftArrow, 0) // to parent ui
	h.Key(key.NameLeftArrow, 0) // collapse ui
	h.Frame()
	if tr.Value() != "ui" || shown(h, "kit") {
		t.Fatalf("left: %q kit shown=%v", tr.Value(), shown(h, "kit"))
	}
	tr.SetValue("kit") // expands its ancestors
	h.Frame()
	if !shown(h, "kit") {
		t.Fatal("SetValue did not reveal the node")
	}
}

func TestTableSortSelectKeysResize(t *testing.T) {
	rows := [][]string{{"SO-1", "华东", "300"}, {"SO-2", "北京", "1200"}, {"SO-3", "深圳", "45"}}
	var activated []int
	tb := Table(Col("单号"), Col("客户").Flex(1.4), Col("金额").Numeric()).Height(200).OnActivate(func(r int) { activated = append(activated, r) })
	tb.SetRows(rows)
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().P(10).Child(tb.Render(cx)) })))
	if _, ok := semanticNode(h, "table:3 行"); !ok {
		t.Fatal("table semantics")
	}
	click(t, h, "金额")
	h.Frame()
	// numeric ascending: 45, 300, 1200
	click(t, h, "SO-1 | 华东 | 300")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if tb.Value() != 1 || len(activated) != 1 || activated[0] != 1 {
		t.Fatalf("selected %d activated %v", tb.Value(), activated)
	}
	click(t, h, "金额") // descending
	h.Frame()
	h.Key(key.NameHome, 0)
	if tb.Value() != 1 {
		t.Fatalf("Home in descending order should select 1200: %d", tb.Value())
	}
	before := bounds(h, "客户").Dx()
	hb := bounds(h, "客户")
	x, y := float32(hb.Max.X-3), float32(hb.Min.Y+hb.Dy()/2)
	h.Drag(x, y, x+60, y)
	h.Frame()
	if after := bounds(h, "客户").Dx(); after < before+40 {
		t.Fatalf("resize: %d → %d", before, after)
	}
	if tb.sortCol != 2 || !tb.desc {
		t.Fatal("dragging a resize handle changed the sort")
	}
	tb.SetRows(nil)
	h.Frame()
	if !shown(h, "暂无数据") {
		t.Fatal("empty text")
	}
}

func TestPaginationNumbersAndNavigation(t *testing.T) {
	p := Pagination(195, 10)
	for page, want := range map[int]string{1: "1 2 3 4 0 20", 10: "1 0 9 10 11 0 20", 20: "1 0 17 18 19 20"} {
		p.SetValue(page)
		got := ""
		for i, n := range p.numbers() {
			if i > 0 {
				got += " "
			}
			got += strconv.Itoa(n)
		}
		if got != want {
			t.Errorf("page %d: %s want %s", page, got, want)
		}
	}
	p.SetValue(1)
	var seen []int
	p.OnChange(func(n int) { seen = append(seen, n) })
	h := page(p)
	click(t, h, "下一页")
	click(t, h, "20")
	if p.Value() != 20 || len(seen) != 2 {
		t.Fatalf("page %d %v", p.Value(), seen)
	}
	if s, e := p.Bounds(); s != 190 || e != 195 {
		t.Fatalf("bounds %d %d", s, e)
	}
	click(t, h, "下一页") // disabled on the last page
	if p.Value() != 20 {
		t.Fatal("went past the last page")
	}
}

func TestCommandFuzzyKeysAndRun(t *testing.T) {
	ran := ""
	cmd := Command(
		CommandItem{Title: "New window", Shortcut: "mod+shift+n", Action: func() { ran = "window" }},
		CommandItem{Title: "Open settings", Group: "偏好", Action: func() { ran = "settings" }},
		CommandItem{Title: "新建订单", Shortcut: "mod+n", Action: func() { ran = "order" }},
		CommandItem{Title: "打开设置", Action: func() { ran = "设置" }},
	)
	if s, ok := fuzzy("nwo", "New window"); !ok || s >= 2000 {
		t.Fatal("subsequence")
	}
	if a, _ := fuzzy("new", "New window"); a < 2000 {
		t.Fatal("prefix should rank highest")
	}
	if _, ok := fuzzy("xyz", "New window"); ok {
		t.Fatal("non-match")
	}
	h := page(cmd)
	cmd.Toggle()
	h.Frame()
	h.Frame()
	h.Type("设置")
	h.Frame()
	if shown(h, "New window") || !shown(h, "打开设置") {
		t.Fatal("filter")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran != "设置" || cmd.Value() {
		t.Fatalf("ran %q open %v", ran, cmd.Value())
	}
	cmd.Toggle()
	h.Frame()
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if ran != "settings" {
		t.Fatalf("↓ Enter ran %q", ran)
	}
}

// A row selected while the table is on another tab scrolls into view when the
// tab shows, without any further input.
func TestTableRevealsSelectionWhenShown(t *testing.T) {
	rows := make([][]string, 100)
	for i := range rows {
		rows[i] = []string{"行 " + strconv.Itoa(i)}
	}
	tb := Table(Col("名称")).Height(200)
	tb.SetRows(rows)
	showing := false
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		if !showing {
			return el.Text("别的页面")
		}
		return el.Div().Child(tb.Render(cx))
	})))
	tb.SetValue(99)
	showing = true
	for range 3 { // the harness renders only when asked; three frames ≈ one input-free settle
		h.Frame()
	}
	if !shown(h, "行 99") {
		t.Fatal("selection not revealed")
	}
}
