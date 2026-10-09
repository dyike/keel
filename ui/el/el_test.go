package el

import (
	"image"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"

	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

// viewFunc adapts a function to View.
type viewFunc func(cx *Context) Element

func (f viewFunc) Render(cx *Context) Element { return f(cx) }

// render lays out root once at 400×300 (1dp = 1px) and returns it.
func render(t *testing.T, root Element) *uitest.Harness {
	t.Helper()
	return uitest.New(Root(viewFunc(func(*Context) Element { return root })))
}

func rect(e Element) image.Rectangle {
	n := e.node()
	return image.Rectangle{Min: n.pos, Max: n.pos.Add(n.size)}
}

func TestRowFixedAndGrow(t *testing.T) {
	a, b, c := Div().W(Dp(50)).H(Dp(10)), Div().Grow().H(Dp(10)), Div().W(Dp(50)).H(Dp(10))
	render(t, Div().Child(Div().Row().W(Dp(300)).Child(a, b, c)))
	if got := [3]int{rect(a).Min.X, rect(b).Min.X, rect(c).Min.X}; got != [3]int{0, 50, 250} {
		t.Fatalf("x positions %v", got)
	}
	if w := b.n.size.X; w != 200 {
		t.Fatalf("grow width %d, want 200", w)
	}
}

func TestPaddingGapJustifyItems(t *testing.T) {
	a, b := Div().Size(Dp(20)), Div().Size(Dp(20))
	box := Div().Row().W(Dp(200)).H(Dp(100)).P(10).Gap(10).Center().Child(a, b)
	render(t, Div().Items(Start).Child(box))
	// content box 180×80; children 20+10+20 = 50 wide, centered: 10 + (180-50)/2 = 75
	if rect(a).Min != image.Pt(75, 40) || rect(b).Min != image.Pt(105, 40) {
		t.Fatalf("a %v b %v", rect(a), rect(b))
	}
}

func TestColumnStretchesWidthMinusPadding(t *testing.T) {
	child := Div().H(Dp(10)).Mx(5)
	col := Div().P(20).Child(child)
	render(t, col)
	if r := rect(child); r.Min.X != 25 || r.Dx() != 400-40-10 {
		t.Fatalf("child %v", r)
	}
	if col.n.size != image.Pt(400, 300) {
		t.Fatalf("root fills the window: %v", col.n.size)
	}
}

func TestTextWrapsInNarrowBox(t *testing.T) {
	one := Text("短")
	long := Text(strings.Repeat("很长的一段文字 ", 10))
	render(t, Div().Items(Start).Child(one, Div().W(Dp(100)).Child(long)))
	if long.n.size.Y < 3*one.n.size.Y {
		t.Fatalf("wrapped text is %d tall, one line is %d", long.n.size.Y, one.n.size.Y)
	}
	if long.n.size.X > 100 {
		t.Fatalf("text %d wider than its box", long.n.size.X)
	}
}

func TestRowShrinksToFit(t *testing.T) {
	a, b := Div().W(Dp(80)).H(Dp(10)), Div().W(Dp(80)).H(Dp(10))
	keep := Div().W(Dp(80)).H(Dp(10)).NoShrink()
	render(t, Div().Items(Start).Child(Div().Row().W(Dp(100)).Child(a, b), Div().Row().W(Dp(100)).Child(keep, Div().W(Dp(80)).H(Dp(10)))))
	if a.n.size.X != 50 || b.n.size.X != 50 {
		t.Fatalf("shrunk to %d and %d, want 50 each", a.n.size.X, b.n.size.X)
	}
	if keep.n.size.X != 80 {
		t.Fatalf("NoShrink child is %d wide", keep.n.size.X)
	}
}

func TestAbsolute(t *testing.T) {
	badge := Div().Absolute().Top(4).Right(4).Size(Dp(10))
	render(t, Div().Items(Start).Child(Div().Size(Dp(100)).Child(badge)))
	if r := rect(badge); r.Min != image.Pt(86, 4) {
		t.Fatalf("badge at %v", r)
	}
}

func TestAbsoluteStretchesBetweenEdges(t *testing.T) {
	line := Div().Absolute().Left(0).Right(0).Bottom(0).H(Dp(2))
	render(t, Div().Items(Start).Child(Div().W(Dp(80)).H(Dp(40)).Child(line)))
	if r := rect(line); r != image.Rect(0, 38, 80, 40) {
		t.Fatalf("underline at %v", r)
	}
}

func TestScrollY(t *testing.T) {
	var rows []Element
	for i := range 10 {
		rows = append(rows, Div().H(Dp(30)).Child(Text("row "+strconv.Itoa(i))))
	}
	list := Div().H(Dp(100)).ScrollY().Children(rows)
	h := uitest.New(Root(viewFunc(func(*Context) Element { return Div().Items(Start).Child(Div().W(Dp(200)).Child(list)) })))
	if list.n.contentH != 300 || list.n.size.Y != 100 {
		t.Fatalf("content %d, height %d", list.n.contentH, list.n.size.Y)
	}
	h.Scroll(50, 50, 90)
	h.Frame()
	st := h // silence unused in case of failure paths
	_ = st
	if y := labelY(h, "row 3"); y > 5 || y < -5 {
		t.Fatalf("after scrolling 90dp, row 3 is at y=%d, want ~0", y)
	}
}

type counter struct{ n int }

func (c *counter) Render(cx *Context) Element {
	return Div().Items(Start).Gap(8).Child(
		Text("count "+strconv.Itoa(c.n)),
		Div().ID("inc").P(8).Bg(rgb(0x2563eb)).OnClick(func() { c.n++ }).Child(Text("+1")),
	)
}

func TestClickShowsInTheSameFrame(t *testing.T) {
	c := &counter{}
	h := uitest.New(Root(c))
	bounds := nodeBounds(h, "+1")
	h.Click(center(bounds))
	if c.n != 1 {
		t.Fatalf("clicked %d times", c.n)
	}
	// Handlers run before Render, so the frame after the click already shows it.
	if _, ok := labelBounds(h, "count 1"); !ok {
		t.Fatal("the frame after the click does not show the new count")
	}
}

func TestInputBind(t *testing.T) {
	var name string
	var submitted string
	h := uitest.New(Root(viewFunc(func(*Context) Element {
		return Div().P(10).Child(Input().ID("name").Placeholder("名字").Bind(&name).OnSubmit(func(s string) { submitted = s }))
	})))
	b := nodeBounds(h, "名字")
	h.Click(center(b))
	h.Type("小明")
	if name != "小明" {
		t.Fatalf("bound string %q", name)
	}
	h.Key(key.NameReturn, 0)
	if submitted != "小明" {
		t.Fatalf("OnSubmit saw %q", submitted)
	}
	name = "程序改的" // a program change shows up in the box
	h.Frame()
	h.Frame()
	if v := description(h, "名字"); v != "程序改的" {
		t.Fatalf("box shows %q after a program change", v)
	}
}

func TestStateFollowsIDs(t *testing.T) {
	order := []string{"a", "b"}
	values := map[string]*string{"a": new(string), "b": new(string)}
	h := uitest.New(Root(viewFunc(func(*Context) Element {
		return Div().Children(Map(order, func(_ int, id string) Element {
			return Input().ID(id).Placeholder(id).Bind(values[id])
		}))
	})))
	h.Click(center(nodeBounds(h, "a")))
	h.Type("typed into a")
	order = []string{"b", "a"} // reorder: a's box keeps its text
	h.Frame()
	if v := description(h, "a"); v != "typed into a" {
		t.Fatalf("after reordering, box a shows %q", v)
	}
}

func TestStateIsDroppedWithTheElement(t *testing.T) {
	show := true
	r := Root(viewFunc(func(*Context) Element {
		return Div().When(show, func(d *DivEl) { d.Child(Div().ID("x").Size(Dp(10)).OnClick(func() {})) })
	}))
	h := uitest.New(r)
	n := len(r.store.states)
	show = false
	h.Frame()
	if len(r.store.states) >= n {
		t.Fatalf("states: %d before, %d after removing the element", n, len(r.store.states))
	}
}

func TestSemantics(t *testing.T) {
	h := render(t, Div().Child(
		Text("标题"),
		Div().OnClick(func() {}).Child(Text("保存")),
		Div().Role("tab").Selected(true).OnClick(func() {}).Child(Text("订单")),
	))
	got := map[string]string{}
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label != "" {
			got[n.Desc.Label] = n.Desc.Class.String() + " " + n.Desc.Description
		}
	})
	if !strings.HasPrefix(got["标题"], "Unknown") {
		t.Errorf("text: %q", got["标题"])
	}
	// A clickable div is a button, named by the text inside it.
	var buttons []string
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Class.String() == "Button" {
			buttons = append(buttons, n.Desc.Description+"|"+childLabel(n))
		}
	})
	slices.Sort(buttons)
	if want := []string{"tab|订单", "|保存"}; !slices.Equal(buttons, []string{"tab|订单", "|保存"}) && !slices.Equal(buttons, []string{"|保存", "tab|订单"}) {
		t.Fatalf("buttons %v, want %v", buttons, want)
	}
}

// helpers over the semantic tree

func walk(n input.SemanticNode, fn func(input.SemanticNode)) {
	fn(n)
	for _, c := range n.Children {
		walk(c, fn)
	}
}

func childLabel(n input.SemanticNode) string {
	for _, c := range n.Children {
		if c.Desc.Label != "" {
			return c.Desc.Label
		}
		if l := childLabel(c); l != "" {
			return l
		}
	}
	return ""
}

func labelBounds(h *uitest.Harness, label string) (image.Rectangle, bool) {
	var r image.Rectangle
	found := false
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if !found && n.Desc.Label == label {
			r, found = n.Desc.Bounds, true
		}
	})
	return r, found
}

func nodeBounds(h *uitest.Harness, label string) image.Rectangle {
	r, _ := labelBounds(h, label)
	return r
}

func labelY(h *uitest.Harness, label string) int {
	r, ok := labelBounds(h, label)
	if !ok {
		return -1 << 30
	}
	return r.Min.Y
}

func description(h *uitest.Harness, label string) string {
	var d string
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == label && n.Desc.Class.String() == "Editor" && d == "" {
			d = n.Desc.Description
		}
	})
	return d
}

func center(r image.Rectangle) (float32, float32) {
	return float32(r.Min.X+r.Max.X) / 2, float32(r.Min.Y+r.Max.Y) / 2
}

func TestStickToBottom(t *testing.T) {
	lines := 5
	root := Root(viewFunc(func(*Context) Element {
		list := Div().ID("log").H(Dp(100)).ScrollY().StickToBottom()
		for i := range lines {
			list.Child(Div().H(Dp(30)).Child(Text("line " + strconv.Itoa(i))))
		}
		return Div().Items(Start).Child(Div().W(Dp(200)).Child(list))
	}))
	h := uitest.New(root)
	bottom := func() string { // the last line fully in view
		last := ""
		for i := range lines {
			if y := labelY(h, "line "+strconv.Itoa(i)); y >= 0 && y+30 <= 110 {
				last = "line " + strconv.Itoa(i)
			}
		}
		return last
	}
	lines = 10
	h.Frame()
	if b := bottom(); b != "line 9" {
		t.Fatalf("growing content: bottom shows %q, want line 9", b)
	}
	h.Scroll(50, 50, -120) // the user scrolls up to read
	lines = 12
	h.Frame()
	if b := bottom(); b == "line 11" {
		t.Fatal("followed new content although the user scrolled up")
	}
	h.Scroll(50, 50, 1000) // back to the end: following resumes
	lines = 14
	h.Frame()
	h.Frame()
	if b := bottom(); b != "line 13" {
		t.Fatalf("after returning to the end, bottom shows %q, want line 13", b)
	}
}

func TestScrollToEndOn(t *testing.T) {
	version := 0
	root := Root(viewFunc(func(*Context) Element {
		list := Div().ID("log").H(Dp(100)).ScrollY().StickToBottom().ScrollToEndOn(version)
		for i := range 10 + version {
			list.Child(Div().H(Dp(30)).Child(Text("line " + strconv.Itoa(i))))
		}
		return Div().Items(Start).Child(Div().W(Dp(200)).Child(list))
	}))
	h := uitest.New(root)
	h.Scroll(50, 50, -1000) // read from the top
	version = 1             // e.g. the user sent a message
	h.Frame()
	if y := labelY(h, "line 10"); y < 0 || y > 100 {
		t.Fatalf("did not jump to the new last line (at y=%d)", y)
	}
}

func TestThemeSwitchRebuildsCachedElements(t *testing.T) {
	original := theme.Current()
	defer theme.Apply(original)
	builds := 0
	var cached Element
	root := Root(viewFunc(func(cx *Context) Element {
		cached = cx.Cache("card", func() Element {
			builds++
			return Div().Bg(theme.Surface).Child(Text("主题内容"))
		})
		return Div().Child(cached)
	}))
	h := uitest.New(root)
	h.Frame()
	if builds != 1 {
		t.Fatal("cache missed without theme change")
	}
	for i, palette := range []theme.Palette{theme.Dark(), theme.Light()} {
		theme.Apply(palette)
		h.Frame()
		if builds != i+2 || *cached.node().style.bg != palette.Surface {
			t.Fatal("cached element kept old theme")
		}
		h.Frame()
		if builds != i+2 {
			t.Fatal("cache did not stabilize after theme change")
		}
	}
}
