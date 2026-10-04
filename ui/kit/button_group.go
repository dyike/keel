package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ButtonGroupView joins buttons into one control: only the outer corners
// are rounded, outlined neighbours share a border, and filled ones are split
// by a hairline. Each button keeps its own click, icon, loading and selected
// state. For a set of exclusive options, use ToggleGroup instead.
type ButtonGroupView struct {
	buttons  []*ButtonView
	name     string
	vertical bool
	disabled bool
}

func ButtonGroup(buttons ...*ButtonView) *ButtonGroupView {
	return &ButtonGroupView{buttons: slices.Clone(buttons)}
}

// Add appends buttons.
func (g *ButtonGroupView) Add(buttons ...*ButtonView) *ButtonGroupView {
	g.buttons = append(g.buttons, buttons...)
	return g
}

// Name is the group's accessible name, such as "Text alignment".
func (g *ButtonGroupView) Name(s string) *ButtonGroupView { g.name = s; return g }

// Vertical stacks the buttons, rounding the top and bottom instead.
func (g *ButtonGroupView) Vertical(on bool) *ButtonGroupView { g.vertical = on; return g }

// SetDisabled disables every button in the group.
func (g *ButtonGroupView) SetDisabled(on bool) { g.disabled = on }

// Buttons returns the group's buttons, to change one after building.
func (g *ButtonGroupView) Buttons() []*ButtonView { return slices.Clone(g.buttons) }

func (g *ButtonGroupView) Render(cx *el.Context) el.Element {
	box := el.Div().Role("group").Name(g.name).Disabled(g.disabled).Items(el.Stretch)
	if !g.vertical {
		box.Row()
	}
	r := float32(theme.RadiusMd)
	last := len(g.buttons) - 1
	for i, b := range g.buttons {
		var c [4]float32 // top left, top right, bottom right, bottom left
		switch {
		case last == 0:
			c = [4]float32{r, r, r, r}
		case i == 0 && g.vertical:
			c = [4]float32{r, r, 0, 0}
		case i == 0:
			c = [4]float32{r, 0, 0, r}
		case i == last && g.vertical:
			c = [4]float32{0, 0, r, r}
		case i == last:
			c = [4]float32{0, r, r, 0}
		}
		e := b.renderCorners(cx, c)
		if i > 0 {
			// Outlined neighbours overlap so their borders read as one line;
			// filled ones leave a hairline of background between them.
			step := float32(1)
			if b.outline {
				step = -1
			}
			if g.vertical {
				e.Mt(step)
			} else {
				e.Ml(step)
			}
		}
		box.Child(e)
	}
	return box
}
