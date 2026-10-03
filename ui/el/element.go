package el

import (
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"github.com/dyike/keel/ui/theme"

	"gioui.org/io/pointer"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
)

// Element is a node of the tree a View renders each frame. Build elements
// with Div, Text, Input and Widget; they are cheap, and a new tree is built on
// every frame.
type Element interface{ node() *Node }

// Node holds what an element was built with and, after layout, where it is.
type Node struct {
	style                       Style
	hover, active               func(*Style)
	focus                       func(*Style)
	focusable                   bool
	focusSet                    bool
	tabConfigured, tabSkip      bool
	tabIndex                    int
	disabled, effectiveDisabled bool
	disabledRoot                bool // disabled here, not only by an ancestor
	disabledStyle               func(*Style)
	onKey                       func(KeyEvent) bool
	id                          string
	keyContext                  string
	keyHint                     *actionKeyHint
	children                    []Element

	text       string
	shimmer    *textShimmer
	textRanges []TextRange
	isText     bool
	input      *inputSpec
	widget     core.Widget
	decorate   func(core.C, func())

	onClick       func()
	onContextMenu func()
	contextButton pointer.Buttons
	focusOnPress  string         // ID to focus on a press no child takes
	palette       *theme.Palette // Themed: colors for this subtree
	onDoubleClick func()
	onDrag        func(DragEvent)
	onScroll      *scrollHandler
	role, name    string
	value         string
	selected      *bool

	// Layout results, in pixels. pos is relative to the parent's border box.
	pos, size     image.Point
	forceW        int // set by the parent: stretch or flex size; -1 none
	forceH        int
	flowPositions []image.Point // wrapped/grid child positions within the content box
	contentW      int           // scroll containers: width of the content
	contentH      int           // scroll containers: height of the content
	key           stateKey
	textStyle     textStyle // resolved, inherited

	// Set on subtrees returned by Context.Cache: they are reused across
	// frames, so a layout with the same inputs can reuse the last result.
	cached bool
	memo   layoutMemo
	reused bool // this frame's layout came from memo: skip placing children
}

type layoutMemo struct {
	ok     bool
	in     [4]int
	parent textStyle
	size   image.Point
}

func (n *Node) node() *Node { return n }

func (n *Node) isFocusable() bool {
	return !n.effectiveDisabled && (n.focusable || !n.focusSet && n.onClick != nil)
}
func (n *Node) interactive() bool {
	return n.focusable || n.onClick != nil || n.onContextMenu != nil || n.onDoubleClick != nil || n.onDrag != nil || n.hover != nil || n.active != nil || n.style.cursor != pointer.CursorDefault
}

// Styled carries the builder methods every element shares. T is the element
// type, so each method returns it and calls chain: Div().P(8).Bg(c).Child(...).
type Styled[T any] struct {
	n    *Node
	self *T
}

func (s *Styled[T]) node() *Node { return s.n }

// Structure and identity.

// ID names the element among its siblings. State that outlives a frame, such
// as hover, scroll position and text box content, is kept per ID; give items
// of lists that change an ID so their state follows them.
func (s *Styled[T]) ID(id string) *T { s.n.id = id; return s.self }

// Child appends children; nil elements are skipped.
func (s *Styled[T]) Child(children ...Element) *T {
	for _, c := range children {
		if c != nil {
			s.n.children = append(s.n.children, c)
		}
	}
	return s.self
}

// Decorate wraps painting with custom operations, e.g. a shared input area
// around a subtree. gtx uses the element's coordinate system and exact size.
// Call draw once to paint the element and its children. This does not run
// during measurement, and must not change the element tree or its layout.
func (s *Styled[T]) Decorate(fn func(gtx core.C, draw func())) *T {
	s.n.decorate = fn
	return s.self
}

// VisitWidgets visits widgets in tree order with their content bounds relative
// to root, after layout. It includes widgets outside the viewport, skips hidden
// elements, and reports layout coordinates before any ScrollY translations.
func VisitWidgets(root Element, metric unit.Metric, visit func(core.Widget, image.Rectangle)) {
	var walk func(*Node, image.Point)
	walk = func(n *Node, origin image.Point) {
		if n.style.hidden {
			return
		}
		if n.widget != nil {
			bw := metric.Dp(unit.Dp(n.style.borderWidth))
			pad := n.style.pad
			min := image.Pt(bw+metric.Dp(unit.Dp(pad.Left)), bw+metric.Dp(unit.Dp(pad.Top)))
			max := n.size.Sub(image.Pt(bw+metric.Dp(unit.Dp(pad.Right)), bw+metric.Dp(unit.Dp(pad.Bottom))))
			visit(n.widget, image.Rectangle{Min: min, Max: max}.Add(origin))
		}
		for _, c := range n.children {
			cn := c.node()
			walk(cn, origin.Add(cn.pos))
		}
	}
	walk(root.node(), image.Point{})
}

// Children is Child for a slice, e.g. from Map.
func (s *Styled[T]) Children(children []Element) *T { return s.Child(children...) }

// When applies fn to the element if cond holds, without breaking the chain.
func (s *Styled[T]) When(cond bool, fn func(*T)) *T {
	if cond {
		fn(s.self)
	}
	return s.self
}

// Layout: direction, alignment, gap.

// Row lays children out left to right. The default is top to bottom.
func (s *Styled[T]) Row() *T {
	s.n.style.row, s.n.style.wrap, s.n.style.grid = true, false, 0
	return s.self
}

// Col lays children out top to bottom (the default).
func (s *Styled[T]) Col() *T {
	s.n.style.row, s.n.style.wrap, s.n.style.grid = false, false, 0
	return s.self
}

// Wrap lays children left to right and starts a new line when width runs out.
// Gap applies between items and lines; Grow distributes space within each line.
func (s *Styled[T]) Wrap() *T {
	s.n.style.row, s.n.style.wrap, s.n.style.grid = true, true, 0
	s.n.style.wrapFit = false
	return s.self
}

// WrapFit wraps like Wrap but hugs each line's content when width is automatic.
// An explicit or stretched width retains normal per-line alignment and Grow.
func (s *Styled[T]) WrapFit() *T {
	s.Wrap()
	s.n.style.wrapFit = true
	return s.self
}

// Grid lays children in equal-width columns (at least one), in row order.
// Columns honor fixed/minimum child widths; rows size to their tallest child.
// Gap applies between columns and rows. Children can span tracks with ColSpan.
func (s *Styled[T]) Grid(columns int) *T {
	s.n.style.row, s.n.style.wrap, s.n.style.grid = false, false, max(columns, 1)
	return s.self
}

// ColSpan sets how many tracks a child occupies in a Grid. Values are clamped
// to 1..the parent's column count; a cell that does not fit starts a new row.
func (s *Styled[T]) ColSpan(columns int) *T { s.n.style.colSpan = max(1, columns); return s.self }

// Gap puts dp between children.
func (s *Styled[T]) Gap(dp float32) *T { s.n.style.gap = dp; return s.self }

// Justify places children along the main axis: Start, Center, End, SpaceBetween, SpaceAround.
func (s *Styled[T]) Justify(a Align) *T { s.n.style.justify = a; return s.self }

// Items aligns children across the main axis: Start, Center, End or Stretch
// (default for columns: children fill the width).
func (s *Styled[T]) Items(a Align) *T { s.n.style.align, s.n.style.alignSet = a, true; return s.self }

// Center centers children on both axes.
func (s *Styled[T]) Center() *T {
	s.n.style.justify, s.n.style.align, s.n.style.alignSet = Center, Center, true
	return s.self
}

// Grow lets the element take free space along its parent's main axis.
func (s *Styled[T]) Grow() *T { s.n.style.grow = 1; return s.self }

// Flex grows like Grow, taking free space in proportion to w: a child with
// Flex(2) gets twice the share of one with Flex(1) or Grow.
func (s *Styled[T]) Flex(w float32) *T { s.n.style.grow = max(w, 0); return s.self }

// NoShrink keeps the element from shrinking below its content size.
func (s *Styled[T]) NoShrink() *T { s.n.style.shrink = -1; return s.self }

// Sizes.

func (s *Styled[T]) W(l Length) *T    { s.n.style.w = l; return s.self }
func (s *Styled[T]) H(l Length) *T    { s.n.style.h = l; return s.self }
func (s *Styled[T]) Size(l Length) *T { s.n.style.w, s.n.style.h = l, l; return s.self }
func (s *Styled[T]) MinW(l Length) *T { s.n.style.minW = l; return s.self }
func (s *Styled[T]) MinH(l Length) *T { s.n.style.minH = l; return s.self }
func (s *Styled[T]) MaxW(l Length) *T { s.n.style.maxW = l; return s.self }
func (s *Styled[T]) MaxH(l Length) *T { s.n.style.maxH = l; return s.self }

// AspectRatio derives an automatic height from a resolved width (width/height).
// Explicit heights take precedence. Zero clears the ratio; invalid values are ignored.
func (s *Styled[T]) AspectRatio(ratio float32) *T {
	if ratio >= 0 && !math.IsInf(float64(ratio), 0) {
		s.n.style.aspectRatio = ratio
	}
	return s.self
}

// WFull and HFull fill the parent's content box.
func (s *Styled[T]) WFull() *T { return s.W(Full) }
func (s *Styled[T]) HFull() *T { return s.H(Full) }

// Spacing, in dp.

func (s *Styled[T]) P(v float32) *T  { s.n.style.pad = Edges{v, v, v, v}; return s.self }
func (s *Styled[T]) Px(v float32) *T { s.n.style.pad.Left, s.n.style.pad.Right = v, v; return s.self }
func (s *Styled[T]) Py(v float32) *T { s.n.style.pad.Top, s.n.style.pad.Bottom = v, v; return s.self }
func (s *Styled[T]) Pt(v float32) *T { s.n.style.pad.Top = v; return s.self }
func (s *Styled[T]) Pb(v float32) *T { s.n.style.pad.Bottom = v; return s.self }
func (s *Styled[T]) Pl(v float32) *T { s.n.style.pad.Left = v; return s.self }
func (s *Styled[T]) Pr(v float32) *T { s.n.style.pad.Right = v; return s.self }
func (s *Styled[T]) M(v float32) *T  { s.n.style.margin = Edges{v, v, v, v}; return s.self }
func (s *Styled[T]) Mx(v float32) *T {
	s.n.style.margin.Left, s.n.style.margin.Right = v, v
	return s.self
}
func (s *Styled[T]) My(v float32) *T {
	s.n.style.margin.Top, s.n.style.margin.Bottom = v, v
	return s.self
}
func (s *Styled[T]) Mt(v float32) *T { s.n.style.margin.Top = v; return s.self }
func (s *Styled[T]) Mb(v float32) *T { s.n.style.margin.Bottom = v; return s.self }

// Scrolling and positioning.

// ScrollY clips the children and scrolls them vertically. The element needs a
// definite height: set H, or let it Grow in a column.
func (s *Styled[T]) ScrollY() *T { s.n.style.scrollY = true; return s.self }

// ScrollX clips and scrolls children horizontally. Set W or constrain the
// width through the parent. ScrollX and ScrollY can be combined.
func (s *Styled[T]) ScrollX() *T { s.n.style.scrollX = true; return s.self }

// StickToBottom keeps a ScrollY container scrolled to the end while content
// grows, as long as the user has not scrolled away from the end: a chat that
// follows a streaming answer but lets the user read back.
func (s *Styled[T]) StickToBottom() *T { s.n.style.stickBottom = true; return s.self }

// ScrollToEndOn scrolls a ScrollY container to the end whenever version
// changes, and resumes StickToBottom: pass the number of messages so that
// sending one jumps to it even after the user scrolled up to read.
func (s *Styled[T]) ScrollToEndOn(version int) *T { s.n.style.endVersion = version; return s.self }

// KeepBottomOn keeps the distance from the bottom of a ScrollY element when
// version changes, so content inserted above (older chat history) does not
// move what is on screen. Bump version in the same callback that inserts.
func (s *Styled[T]) KeepBottomOn(version int) *T { s.n.style.keepVersion = version; return s.self }

// Absolute takes the element out of flow and places it by Top/Right/Bottom/Left
// within its parent's padding box.
func (s *Styled[T]) Absolute() *T         { s.n.style.absolute = true; return s.self }
func (s *Styled[T]) Top(dp float32) *T    { s.n.style.top = &dp; return s.self }
func (s *Styled[T]) Right(dp float32) *T  { s.n.style.right = &dp; return s.self }
func (s *Styled[T]) Bottom(dp float32) *T { s.n.style.bottom = &dp; return s.self }
func (s *Styled[T]) Left(dp float32) *T   { s.n.style.left = &dp; return s.self }

// IsHidden reports the element's declared visibility, before ancestor inheritance.
// Composite views can use it to keep adjoining controls hidden with their body.
func (s *Styled[T]) IsHidden() bool { return s.n.style.hidden }

// Hidden removes the element from layout and paint.
func (s *Styled[T]) Hidden(h bool) *T { s.n.style.hidden = h; return s.self }

// Visuals.

func (s *Styled[T]) Bg(c color.NRGBA) *T { s.n.style.Bg(c); return s.self }

// BgGradient fills the background with a linear gradient; see Style.BgGradient.
func (s *Styled[T]) BgGradient(g theme.Gradient) *T { s.n.style.BgGradient(g); return s.self }

// Border draws a line of width dp inside the element's edge.
func (s *Styled[T]) Border(dp float32, c color.NRGBA) *T {
	s.n.style.borderWidth, s.n.style.borderColor = dp, c
	return s.self
}

// BorderDashed draws 4dp dashes with 3dp gaps; false restores a solid border.
func (s *Styled[T]) BorderDashed(on bool) *T { s.n.style.BorderDashed(on); return s.self }

// Rounded rounds the corners by dp; backgrounds, borders and hit areas follow.
func (s *Styled[T]) Rounded(dp float32) *T { s.n.style.radius = dp; return s.self }

// CursorPointer shows a hand over the element.
func (s *Styled[T]) CursorPointer() *T { s.n.style.cursor = pointer.CursorPointer; return s.self }

// Cursor sets the pointer shape over the element, e.g. pointer.CursorColResize
// on a splitter.
func (s *Styled[T]) Cursor(c pointer.Cursor) *T { s.n.style.cursor = c; return s.self }

// Text style, inherited by descendant text like CSS.

func (s *Styled[T]) TextColor(c color.NRGBA) *T { s.n.style.text.color = &c; return s.self }
func (s *Styled[T]) TextSize(sp float32) *T     { s.n.style.text.size = unit.Sp(sp); return s.self }
func (s *Styled[T]) Bold() *T                   { w := font.Bold; s.n.style.text.weight = &w; return s.self }

// TextAlign sets alignment within wrapped text. Descendants inherit it.
// Start, Center and End are accepted; other values leave the style unchanged.
func (s *Styled[T]) TextAlign(a Align) *T {
	if a == Start || a == Center || a == End {
		s.n.style.text.align = &a
	}
	return s.self
}

// Medium sets a weight between regular and Bold, for emphasis that should
// not shout: selected tabs, table headers.
func (s *Styled[T]) Medium() *T { w := font.Medium; s.n.style.text.weight = &w; return s.self }

// Mono sets theme.MonoFace: code and numbers that must line up.
func (s *Styled[T]) Mono() *T { on := true; s.n.style.text.mono = &on; return s.self }

// LineHeight sets the distance between lines as a multiple of the text
// size, e.g. 1.5 for long paragraphs; descendants inherit it.
func (s *Styled[T]) LineHeight(scale float32) *T { s.n.style.text.lineHeight = scale; return s.self }

// Shadow lifts the element with a soft shadow below it, in theme.Shadow's
// color: theme.ElevationMd for popovers and menus, ElevationLg for dialogs.
// The shadow is drawn outside the element and does not change its size.
func (s *Styled[T]) Shadow(e theme.Elevation) *T { s.n.style.shadow = &e; return s.self }

// Opacity draws the element and its descendants at this alpha, 0..1.
func (s *Styled[T]) Opacity(a float32) *T {
	if a != a {
		return s.self
	}
	s.n.style.opacity, s.n.style.opacitySet = min(max(a, 0), 1), true
	return s.self
}

// MaxLines truncates text to n lines with an ellipsis.
func (s *Styled[T]) MaxLines(n int) *T { s.n.style.text.lines = n; return s.self }

// State variants: visual changes while hovered or pressed.
func (s *Styled[T]) Hover(fn func(*Style)) *T  { s.n.hover = fn; return s.self }
func (s *Styled[T]) Active(fn func(*Style)) *T { s.n.active = fn; return s.self }

// Interaction.

// OnClick runs fn on a primary click. It runs before the next render, so the
// frame that follows already shows its effects.
func (s *Styled[T]) OnClick(fn func()) *T { s.n.onClick = fn; return s.self }

// OnDoubleClick runs fn on a double click (OnClick also sees both clicks).
func (s *Styled[T]) OnDoubleClick(fn func()) *T { s.n.onDoubleClick = fn; return s.self }

// Agent semantics: roles are inferred (OnClick → button, Text → text, Input
// → textbox); these override or add to that.

// Role sets the role agents see, e.g. "tab", "row", "option", "dialog".
func (s *Styled[T]) Role(role string) *T { s.n.role = role; return s.self }

// Name sets the name agents see; by default it is the text inside.
func (s *Styled[T]) Name(name string) *T { s.n.name = name; return s.self }

// Value reports a value with the role, e.g. "40%" for a progressbar.
func (s *Styled[T]) Value(v string) *T { s.n.value = v; return s.self }

// Selected reports a selected or checked state to agents.
func (s *Styled[T]) Selected(v bool) *T { s.n.selected = &v; return s.self }

func newNode() *Node { return &Node{forceW: -1, forceH: -1} }

// DivEl is a box: the only element with children.
type DivEl struct{ Styled[DivEl] }

// Div creates an empty box. Children stack top to bottom unless Row is set.
func Div() *DivEl {
	d := &DivEl{}
	d.n, d.self = newNode(), d
	return d
}

// TextEl is a run of text. It wraps to the available width.
type TextEl struct{ Styled[TextEl] }

// Text creates a text element; style it like any element (TextColor, TextSize,
// Bold) or let it inherit from its parent.
func Text(s string) *TextEl {
	t := &TextEl{}
	t.n, t.self = newNode(), t
	t.n.text, t.n.isText = s, true
	return t
}

// WidgetEl wraps any core.Widget, such as Gio code wrapped in core.Func.
type WidgetEl struct{ Styled[WidgetEl] }

// Widget embeds w. It is laid out with the element's box as its constraints.
func Widget(w core.Widget) *WidgetEl {
	e := &WidgetEl{}
	e.n, e.self = newNode(), e
	e.n.widget = w
	return e
}

// Map builds one element per item, for Children.
func Map[T any](items []T, fn func(int, T) Element) []Element {
	out := make([]Element, 0, len(items))
	for i, it := range items {
		out = append(out, fn(i, it))
	}
	return out
}

// PinLeft keeps an element at an offset from its nearest ScrollX viewport's
// left edge. It retains layout space and paints above unpinned siblings.
func (s *Styled[T]) PinLeft(dp float32) *T {
	s.n.style.pinX = -1
	s.n.style.pinOffset = max(0, dp)
	return s.self
}

// PinRight is PinLeft relative to the right edge.
func (s *Styled[T]) PinRight(dp float32) *T {
	s.n.style.pinX = 1
	s.n.style.pinOffset = max(0, dp)
	return s.self
}

// OnContextMenu runs fn on a secondary pointer press. It does not consume
// primary clicks; add an OnKey handler for a keyboard context-menu action.
func (s *Styled[T]) OnContextMenu(fn func()) *T { return s.OnMousePress(pointer.ButtonSecondary, fn) }

// OnMousePress observes a primary, secondary or tertiary press without consuming
// descendant events or adding a Tab stop. Chords are ignored. It shares a handler
// with OnContextMenu; the last call wins. Zero clears it; invalid buttons are ignored.
func (s *Styled[T]) OnMousePress(button pointer.Buttons, fn func()) *T {
	switch button {
	case 0:
		s.n.onContextMenu = nil
		s.n.contextButton = 0
	case pointer.ButtonPrimary, pointer.ButtonSecondary, pointer.ButtonTertiary:
		s.n.onContextMenu = fn
		s.n.contextButton = button
	}
	return s.self
}

// Reveal exposes a fraction of this element's natural height, clipping both
// painting and input. Children retain their full layout, so text does not
// reflow vertically during an expand/collapse animation. NaN becomes zero.
func (s *Styled[T]) Reveal(fraction float32) *T {
	if fraction != fraction {
		fraction = 0
	}
	s.n.style.reveal = max(0, min(1, fraction))
	s.n.style.revealSet = true
	return s.self
}

// Themed renders a view with another palette, for a part of the window in
// different colors: a dark sidebar in a light window, a preview of a theme.
// The palette applies while the view renders and while it paints, so kit
// components and custom drawing inside follow it; the rest of the window
// keeps the global theme. The wrapper stretches its child; set its
// background to fill the area.
func (cx *Context) Themed(p theme.Palette, v View) *DivEl {
	restore := theme.Scope(p)
	child := v.Render(cx)
	restore()
	d := Div().Items(Stretch).Child(child)
	d.n.palette = &p
	return d
}

// ScrollOffset controls absolute x/y offsets in dp for ScrollX/ScrollY. Offsets
// are clamped at paint time, including disabled frames. Omit it to allow user
// scrolling; when supplied, native scroll gestures and scrollbars are omitted.
func (s *Styled[T]) ScrollOffset(x, y float32) *T {
	if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) || math.IsNaN(float64(y)) || math.IsInf(float64(y), 0) {
		return s.self
	}
	s.n.style.controlledScroll = &[2]float32{x, y}
	return s.self
}

// ContentBottom selects an in-flow descendant whose border-box bottom aligns
// with siblings in a row using Items(ContentBottom). Nil, hidden, or missing
// descendants fall back to this element's bottom. This is geometric alignment,
// not a font baseline. Pass an element from the current render tree.
func (s *Styled[T]) ContentBottom(target Element) *T {
	s.n.style.contentBottom = nil
	if target != nil {
		s.n.style.contentBottom = target.node()
	}
	return s.self
}
