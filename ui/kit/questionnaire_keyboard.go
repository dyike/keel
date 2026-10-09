package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
)

func (v *QuestionnaireView) choiceDisabled(id, option string) bool {
	return v.disabledChoices[id][option]
}

// SetChoiceDisabled excludes a choice from effective answers and shortcuts.
// The stored selection is retained for re-enabling. Unknown IDs/options return false.
func (v *QuestionnaireView) SetChoiceDisabled(id, option string, on bool) bool {
	i := v.index(id)
	if i < 0 || !slices.Contains(v.questions[i].Options, option) {
		return false
	}
	if v.disabledChoices == nil {
		v.disabledChoices = map[string]map[string]bool{}
	}
	if v.disabledChoices[id] == nil {
		v.disabledChoices[id] = map[string]bool{}
	}
	if v.disabledChoices[id][option] == on {
		return true
	}
	v.disabledChoices[id][option] = on
	c := v.control(v.questions[i])
	if c.radio != nil {
		c.radio.SetOptionDisabled(option, on)
	}
	for j, ch := range c.checks {
		if v.questions[i].Options[j] == option {
			ch.SetDisabled(on)
		}
	}
	v.completed = false
	v.err = ""
	v.heldShortcuts = nil
	if i == v.page {
		v.focusPending = true
	}
	return true
}
func (v *QuestionnaireView) enabledOptions(q Question) []int {
	var result []int
	for i, option := range q.Options {
		if !v.choiceDisabled(q.ID, option) {
			result = append(result, i)
		}
	}
	return result
}

// KeyboardNavigation toggles built-in arrows and confirmation; option shortcut configuration is independent.
func (v *QuestionnaireView) KeyboardNavigation(on bool) *QuestionnaireView {
	v.noNavigation = !on
	return v
}
func (v *QuestionnaireView) focusTargets(q Question) []string {
	c := v.control(q)
	var targets []string
	if c.radio != nil {
		if id := c.radio.FocusID(); id != "" {
			targets = append(targets, id)
		}
	}
	for i, ch := range c.checks {
		if !v.choiceDisabled(q.ID, q.Options[i]) {
			targets = append(targets, ch.FocusID())
		}
	}
	if c.input != nil {
		targets = append(targets, c.input.FocusID())
	}
	if c.rating != nil {
		targets = append(targets, c.rating.FocusID())
	}
	if c.freeform != nil {
		targets = append(targets, c.freeform.FocusID())
	}
	return targets
}
func (v *QuestionnaireView) confirm() {
	if v.page < 0 || v.page >= len(v.questions) || v.answer(v.questions[v.page]).Empty() && !v.skipped[v.questions[v.page].ID] {
		return
	}
	v.next()
}
func (v *QuestionnaireView) registerNavigation(cx *el.Context) {
	id := autoID("questionnaire", v)
	if !v.disabled && !v.noNavigation && cx.Enabled(id) && cx.FocusWithin(id) {
		cx.Shortcut("mod+enter", func() {
			if cx.Enabled(id) && cx.FocusWithin(id) {
				v.confirm()
			}
		})
	}
}
func (v *QuestionnaireView) navigationKey(cx *el.Context, e el.KeyEvent, q Question, c *questionControl, inText bool) bool {
	if v.noNavigation {
		return false
	}
	if key.Name(e.Name) == key.NameReturn && e.Modifiers == key.ModShortcut {
		if e.State == el.KeyPress {
			v.confirm()
		}
		return true
	}
	if inText || e.Modifiers != 0 || cx.FocusWithin(autoID("questionnaire", v)+"/actions") {
		return false
	}
	inRadio := false
	if c.radio != nil {
		for _, option := range q.Options {
			inRadio = inRadio || cx.FocusWithin(c.radio.itemID(option))
		}
	}
	switch key.Name(e.Name) {
	case key.NameReturn:
		if e.State == el.KeyPress {
			v.confirm()
		}
		return !v.answer(q).Empty() || v.skipped[q.ID]
	case key.NameLeftArrow, key.NameRightArrow:
		if inRadio {
			return false
		}
		if e.State == el.KeyPress {
			if key.Name(e.Name) == key.NameLeftArrow {
				v.previous()
			} else {
				v.confirm()
			}
		}
		return true
	case key.NameUpArrow, key.NameDownArrow:
		if inRadio || c.rating != nil && cx.FocusWithin(c.rating.FocusID()) {
			return false
		}
		targets := v.focusTargets(q)
		if len(targets) == 0 {
			return false
		}
		pos := -1
		for i, id := range targets {
			if cx.FocusWithin(id) {
				pos = i
				break
			}
		}
		if key.Name(e.Name) == key.NameDownArrow {
			pos++
		} else if pos >= 0 {
			pos--
		} else {
			pos = 0
		}
		if e.State == el.KeyPress {
			cx.Focus(targets[min(max(pos, 0), len(targets)-1)])
		}
		return true
	}
	return false
}
