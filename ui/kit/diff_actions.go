package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// DiffAction is an application operation on a specific region. OnSelect receives
// a fresh snapshot of that region. It does not implicitly accept or revert it.
type DiffAction struct {
	Label    string
	Disabled bool
	OnSelect func(DiffHunk)
}

// HunkActions supplies actions embedded immediately before each changed region.
// Rows scroll with the source; narrow rows scroll horizontally. The factory runs
// when the comparison/presentation is rebuilt, not for each frame. It must not
// mutate this view. Pass nil to remove the rows.
func (v *DiffView) HunkActions(factory func(DiffHunk) []DiffAction) *DiffView {
	v.hunkActions = factory
	v.rebuildActions()
	return v
}
func (v *DiffView) rebuildActions() {
	v.actionGeneration++
	generation := v.actionGeneration
	v.old.setBlocks(nil)
	v.new.setBlocks(nil)
	v.unified.setBlocks(nil)
	if v.hunkActions == nil {
		return
	}
	var oldBlocks, newBlocks, unifiedBlocks []codeBlock
	for index, c := range v.changes {
		hunk, _ := v.Hunk(index)
		actions := append([]DiffAction(nil), v.hunkActions(hunk)...)
		if len(actions) == 0 {
			continue
		}
		makeRoot := func() *el.RootWidget {
			return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
				row := el.Div().Row().Gap(theme.SpaceSm).Items(el.Center)
				for _, action := range actions {
					button := Button(action.Label, func() {
						if v.disabled || action.Disabled || generation != v.actionGeneration || action.OnSelect == nil {
							return
						}
						if current, ok := v.Hunk(index); ok {
							action.OnSelect(current)
						}
					}).Variant(ButtonSecondary).Size(44)
					button.SetDisabled(action.Disabled || action.OnSelect == nil || v.disabled)
					row.Child(button.Render(cx))
				}
				return el.Div().P(4).ScrollX().Child(row)
			}))
		}
		oldBlocks = append(oldBlocks, codeBlock{line: c.oldStart})
		newBlocks = append(newBlocks, codeBlock{line: c.newStart, widget: makeRoot()})
		unifiedBlocks = append(unifiedBlocks, codeBlock{line: v.positions[index], widget: makeRoot()})
	}
	v.old.setBlocks(oldBlocks)
	v.new.setBlocks(newBlocks)
	v.unified.setBlocks(unifiedBlocks)
}
