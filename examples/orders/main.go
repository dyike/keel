// Orders is a small order desk: search and filter a table, sort it, select
// rows with the mouse or keyboard, delete with confirmation, add orders through
// a form, and see totals per status.
//
// The screen is an el.View: plain state in a struct, rendered as a tree of
// styled elements every frame. Components not yet ported to el (table,
// dropdown, radio, switch, dialog) are embedded with el.Widget.
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type order struct {
	id, customer, status, pay string
	amount                    float64
	urgent                    bool
}

var statuses = []string{"待付款", "已付款", "已发货", "已完成"}

type desk struct {
	orders  []order
	nextID  int
	tab     int
	query   string
	visible []int // table row → index into orders

	filter *kit.SelectView
	table  *kit.TableView
	dlg    *kit.DialogView

	// New order form.
	customer, amount *kit.InputView
	newStatus        *kit.SelectView
	pay              *kit.RadioGroupView
	urgent           *kit.SwitchView
	form             *kit.FormView
}

func newDesk() *desk {
	d := &desk{dlg: kit.Dialog("")}
	customers := []string{"华东物流", "北京百货", "深圳电子", "成都餐饮", "杭州茶业", "上海文具"}
	for i := range 36 {
		d.orders = append(d.orders, order{
			id: fmt.Sprintf("SO-%04d", 1001+i), customer: customers[i%len(customers)],
			status: statuses[(i*7)%len(statuses)], pay: "转账", amount: float64(300 + (i*137)%2900),
		})
	}
	d.nextID = 1001 + len(d.orders)
	d.filter = kit.Select("", append([]string{"全部状态"}, statuses...)...).OnChange(func(string) { d.refresh() })
	d.filter.SetValue("全部状态")
	d.table = kit.Table(kit.Col("单号"), kit.Col("客户").Flex(1.4), kit.Col("状态"), kit.Col("金额").Numeric()).
		Height(300).OnActivate(func(r int) {
		o := d.orders[d.visible[r]]
		d.dlg.Alert(o.id, fmt.Sprintf("%s · %s · ¥%.2f · %s付款", o.customer, o.status, o.amount, o.pay), nil)
	})
	d.customer = kit.Input("").Placeholder("客户名称")
	d.amount = kit.Input("").Placeholder("0.00").Filter("0123456789.")
	d.newStatus = kit.Select("", statuses...)
	d.newStatus.SetValue("待付款")
	d.pay = kit.RadioGroup("", "转账", "支票", "现金").Horizontal()
	d.pay.SetValue("转账")
	d.urgent = kit.Switch("加急处理", false)
	d.form = kit.Form().LabelWidth(64).
		Field("客户", d.customer, func() string { return kit.Required(d.customer.Value(), "请填写客户名称。") }).
		Field("金额", d.amount, func() string {
			if amt, err := strconv.ParseFloat(strings.TrimSpace(d.amount.Value()), 64); err != nil || amt <= 0 {
				return "金额必须是大于 0 的数字。"
			}
			return ""
		}).
		Field("状态", d.newStatus, nil).
		Field("付款方式", d.pay, nil).
		Field("选项", d.urgent, nil)
	d.refresh()
	return d
}

// refresh applies search and filter to the table.
func (d *desk) refresh() {
	q := strings.TrimSpace(d.query)
	want := d.filter.Value()
	d.visible = d.visible[:0]
	var rows [][]string
	for i, o := range d.orders {
		if q != "" && !strings.Contains(o.customer, q) && !strings.Contains(o.id, q) {
			continue
		}
		if want != "" && want != "全部状态" && o.status != want {
			continue
		}
		d.visible = append(d.visible, i)
		rows = append(rows, []string{o.id, o.customer, o.status, strconv.FormatFloat(o.amount, 'f', 2, 64)})
	}
	d.table.SetRows(rows)
}

func (d *desk) selected() (int, bool) {
	if r := d.table.Value(); r >= 0 && r < len(d.visible) {
		return d.visible[r], true
	}
	return 0, false
}

func (d *desk) Render(cx *el.Context) el.Element {
	cx.Shortcut("mod+n", func() { d.tab = 1 })
	pages := []func() el.Element{func() el.Element { return d.listPage(cx) }, func() el.Element { return d.formPage(cx) }, d.statsPage}
	return el.Div().P(24).Gap(16).ScrollY().Child(
		el.Text("订单管理").TextSize(22).Bold(),
		el.Div().Child(
			el.Div().Row().Child(tab(d, 0, "订单列表"), tab(d, 1, "新建订单"), tab(d, 2, "统计")),
			el.Div().H(el.Dp(1)).Bg(theme.Border),
		),
		pages[d.tab](),
		d.dlg.Render(cx),
	)
}

func tab(d *desk, i int, title string) el.Element {
	active := d.tab == i
	return el.Div().ID(title).Px(14).Py(8).Role("tab").Selected(active).CursorPointer().
		TextColor(theme.Muted).
		When(active, func(t *el.DivEl) {
			t.TextColor(theme.Primary).Bold().Child(el.Div().Absolute().Left(0).Right(0).Bottom(0).H(el.Dp(2)).Bg(theme.Primary))
		}).
		OnClick(func() { d.tab = i }).
		Child(el.Text(title))
}

type kind int

const (
	primary kind = iota
	secondary
	danger
)

// button is the app's button: a styled div. Use it like a component.
func button(label string, k kind, onClick func()) el.Element {
	bg, hover, fg := theme.Primary, theme.PrimaryHover, theme.OnColor
	switch k {
	case secondary:
		bg, hover, fg = theme.Subtle, theme.SubtleHover, theme.Text
	case danger:
		bg, hover, fg = theme.Danger, theme.DangerHover, theme.OnColor
	}
	return el.Div().ID(label).Px(16).Py(8).Rounded(6).Bg(bg).TextColor(fg).TextSize(14).
		CursorPointer().Hover(func(s *el.Style) { s.Bg(hover) }).
		OnClick(onClick).Child(el.Text(label))
}

func (d *desk) listPage(cx *el.Context) el.Element {
	shown := fmt.Sprintf("共 %d 条，显示 %d 条", len(d.orders), d.table.Len())
	return el.Div().Gap(12).Child(
		el.Div().Row().Gap(8).Child(
			el.Input().ID("q").Placeholder("搜索客户或单号").Bind(&d.query).OnChange(func(string) { d.refresh() }).Grow(),
			el.Div().W(el.Dp(140)).Child(d.filter.Render(cx)),
		),
		d.table.Render(cx),
		el.Div().Row().Gap(8).Items(el.Center).Child(
			el.Text(shown).TextColor(theme.Muted).TextSize(13).Grow(),
			button("标记已发货", secondary, func() {
				if i, ok := d.selected(); ok {
					d.orders[i].status = "已发货"
					d.refresh()
				}
			}),
			button("删除所选", danger, d.remove),
		),
	)
}

func (d *desk) remove() {
	i, ok := d.selected()
	if !ok {
		d.dlg.Alert("没有选中订单", "先在表格里选中一行。", nil)
		return
	}
	d.dlg.ConfirmDanger("删除订单", fmt.Sprintf("确定删除 %s（%s）？", d.orders[i].id, d.orders[i].customer), "删除", func() {
		d.orders = append(d.orders[:i], d.orders[i+1:]...)
		d.table.SetValue(-1)
		d.refresh()
	})
}

func (d *desk) formPage(cx *el.Context) el.Element {
	return el.Div().Gap(16).Items(el.Start).Child(
		el.Div().W(el.Dp(420)).Child(d.form.Render(cx)),
		el.Div().Pl(76).Child(button("保存订单", primary, func() { d.save(cx) })),
	)
}

func (d *desk) save(cx *el.Context) {
	if !d.form.Validate(cx) {
		return
	}
	amt, _ := strconv.ParseFloat(strings.TrimSpace(d.amount.Value()), 64)
	d.orders = append(d.orders, order{id: fmt.Sprintf("SO-%04d", d.nextID), customer: strings.TrimSpace(d.customer.Value()),
		status: d.newStatus.Value(), pay: d.pay.Value(), amount: amt, urgent: d.urgent.Value()})
	d.nextID++
	d.customer.SetValue("")
	d.amount.SetValue("")
	d.query = ""
	d.filter.SetValue("全部状态")
	d.refresh()
	d.table.SortBy(-1, false)
	d.table.SetValue(d.table.Len() - 1)
	d.tab = 0
}

func (d *desk) statsPage() el.Element {
	total := 0.0
	count := map[string]int{}
	for _, o := range d.orders {
		total += o.amount
		count[o.status]++
	}
	page := el.Div().Gap(14).Child(el.Text(fmt.Sprintf("订单 %d 笔，合计 ¥%.2f", len(d.orders), total)))
	for _, s := range statuses {
		frac := 0.0
		if len(d.orders) > 0 {
			frac = float64(count[s]) / float64(len(d.orders))
		}
		page.Child(progress(fmt.Sprintf("%s %d 笔", s, count[s]), frac))
	}
	return page
}

// progress is a labelled bar; agents see it as a progressbar with its percentage.
func progress(label string, frac float64) el.Element {
	pct := fmt.Sprintf("%d%%", int(frac*100+0.5))
	return el.Div().Gap(6).Role("progressbar").Name(label).Value(pct).Child(
		el.Div().Row().TextColor(theme.Muted).TextSize(13).Child(el.Text(label).Grow(), el.Text(pct)),
		el.Div().H(el.Dp(8)).Rounded(4).Bg(theme.Subtle).Items(el.Start).Child(
			el.Div().H(el.Dp(8)).Rounded(4).Bg(theme.Primary).W(el.Frac(float32(frac))),
		),
	)
}

func main() {
	screenshot := flag.String("screenshot", "", "render a PNG to this path and exit")
	flag.Parse()
	d := newDesk()
	root := el.Root(d)
	if *screenshot != "" {
		if err := window.Screenshot(root, 760, 600, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "订单管理", Width: 760, Height: 600, Content: root})
	window.Main()
}
