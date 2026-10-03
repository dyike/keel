package kit

import "github.com/dyike/keel/ui/el"

// DockSkin styles a Dock independently of its persisted arrangement. Callbacks
// run on fresh elements each frame. Use them for colors, borders, typography
// and panel padding; preserve element identity, children and event handlers.
// Separators retain their 4dp thickness; do not change their geometry.
type DockSkin struct {
	Panel     func(*el.DivEl)
	Header    func(*el.DivEl)
	Body      func(*el.DivEl)
	Tab       func(tab *el.DivEl, selected bool)
	Separator func(*el.DivEl)
}

// Skin uses a shared live style configuration. Passing nil restores defaults
// without changing panel contents, focus or layout. Mutate it on the UI thread.
func (v *DockView) Skin(skin *DockSkin) *DockView { v.skin = skin; return v }

func (v *DockView) styleSeparator(h *el.DivEl) {
	if v.skin != nil && v.skin.Separator != nil {
		v.skin.Separator(h)
	}
}
