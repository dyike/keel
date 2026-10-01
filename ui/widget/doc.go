// Package widget provides interactive components: text, buttons, links, text
// fields and checkboxes, one component per file.
//
// To add a component: keep its state in a struct, implement core.Widget, read
// colors from theme, and run user callbacks through core.Call.
//
// A false gtx.Enabled() can mean either a disabled parent or no input source
// (for example, el measurement or off-screen layout). Use it only to decide
// appearance, semantics and whether to process events; never use it to mutate
// state retained across frames. Reset state such as open menus, saved focus or
// active drags in SetDisabled(true), as SliderView.SetDisabled does.
package widget

import "github.com/dyike/keel/ui/core"

type (
	C = core.C
	D = core.D
)
