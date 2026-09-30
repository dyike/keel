// Package widget provides interactive components: text, buttons, links, text
// fields and checkboxes, one component per file.
//
// To add a component: keep its state in a struct, implement core.Widget, read
// colors from theme, and run user callbacks through core.Call.
package widget

import "github.com/dyike/keel/ui/core"

type (
	C = core.C
	D = core.D
)
