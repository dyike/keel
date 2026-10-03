package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"strings"
)

// QuestionnaireShortcuts enables unmodified choice shortcuts outside text inputs.
type QuestionnaireShortcuts uint8

const (
	QuestionnaireShortcutsOff QuestionnaireShortcuts = iota
	QuestionnaireShortcutsLetters
	QuestionnaireShortcutsNumbers
)

// QuestionnaireProgress is an enabled-item snapshot; Current is one-based, or zero.
type QuestionnaireProgress struct {
	Current, Total, Answered, Skipped, Unanswered int
	Completed                                     bool
}

func (v *QuestionnaireView) Progress() QuestionnaireProgress {
	p := QuestionnaireProgress{Completed: v.completed}
	for i, q := range v.questions {
		if q.Disabled {
			continue
		}
		p.Total++
		if i == v.page {
			p.Current = p.Total
		}
		a := v.answer(q)
		if a.Skipped {
			p.Skipped++
		} else if a.Empty() {
			p.Unanswered++
		} else {
			p.Answered++
		}
	}
	return p
}
func (v *QuestionnaireView) Shortcuts(mode QuestionnaireShortcuts) *QuestionnaireView {
	if mode <= QuestionnaireShortcutsNumbers {
		v.shortcuts = mode
	}
	return v
}
func (v *QuestionnaireView) OnAnswerChange(fn func(string, Answer)) *QuestionnaireView {
	v.onChange = fn
	return v
}
func (v *QuestionnaireView) OnComplete(fn func(map[string]Answer)) *QuestionnaireView {
	v.onComplete = fn
	return v
}
func (v *QuestionnaireView) index(id string) int {
	for i, q := range v.questions {
		if q.ID == id {
			return i
		}
	}
	return -1
}

// SetQuestionDisabled removes a question from navigation, validation, progress and submission.
// Its stored answer is retained and remains available in Value.
func (v *QuestionnaireView) SetQuestionDisabled(id string, on bool) bool {
	i := v.index(id)
	if i < 0 {
		return false
	}
	if v.questions[i].Disabled == on {
		return true
	}
	v.questions[i].Disabled = on
	v.completed = false
	v.SetPage(max(v.page, 0))
	return true
}

// SetExternalError stores an owner-managed error. Empty clears it; edits and Reset preserve it.
func (v *QuestionnaireView) SetExternalError(id, message string) bool {
	if v.index(id) < 0 {
		return false
	}
	v.external[id] = message
	v.completed = false
	return true
}
func (v *QuestionnaireView) ExternalError(id string) string { return v.external[id] }
func (v *QuestionnaireView) enabledAnswers() map[string]Answer {
	answers := v.Value()
	for _, q := range v.questions {
		if q.Disabled {
			delete(answers, q.ID)
		}
	}
	return answers
}
func (v *QuestionnaireView) validate(i int) string {
	q := v.questions[i]
	if q.Disabled {
		return ""
	}
	if message := v.external[q.ID]; message != "" {
		return message
	}
	a := v.answer(q)
	if a.Skipped {
		return ""
	}
	if a.Empty() {
		if !q.Required {
			return locale.Current().AnswerOrSkip
		}
		return locale.Current().Required
	}
	if q.Validate != nil {
		return q.Validate(a, v.enabledAnswers())
	}
	return ""
}
func (v *QuestionnaireView) changed(id string) {
	v.err = ""
	v.completed = false
	delete(v.skipped, id)
	if v.onChange != nil {
		i := v.index(id)
		if i >= 0 {
			v.onChange(id, v.answer(v.questions[i]))
		}
	}
}
func (v *QuestionnaireView) previous() {
	if v.disabled {
		return
	}
	for i := v.page - 1; i >= 0; i-- {
		if !v.questions[i].Disabled {
			v.SetPage(i)
			return
		}
	}
}

// Skip clears an optional answer and records an intentional skipped state.
// Skipping the final enabled question attempts submission.
func (v *QuestionnaireView) Skip() {
	if v.disabled || v.page < 0 || v.page >= len(v.questions) {
		return
	}
	page := v.page
	q := v.questions[v.page]
	if q.Required || q.Disabled {
		return
	}
	v.SetValue(map[string]Answer{q.ID: {Skipped: true}})
	v.err = ""
	if v.onChange != nil {
		v.onChange(q.ID, v.answer(q))
	}
	if v.page == page {
		v.next()
	}
}

// Reset restores schema default answers and clears completion, preserving external errors and disabled conditions.
func (v *QuestionnaireView) Reset() {
	values := map[string]Answer{}
	for _, q := range v.questions {
		values[q.ID] = q.DefaultAnswer
	}
	v.SetValue(values)
	v.SetPage(0)
}
func (v *QuestionnaireView) handleKey(cx *el.Context, e el.KeyEvent) bool {
	if v.disabled || v.page < 0 || v.page >= len(v.questions) {
		return false
	}
	q := v.questions[v.page]
	c := v.control(q)
	inText := c.input != nil && cx.FocusWithin(c.input.FocusID()) || c.freeform != nil && cx.FocusWithin(c.freeform.FocusID())
	if v.navigationKey(cx, e, q, c, inText) {
		return true
	}
	if inText {
		v.heldShortcuts = nil
		return false
	}
	if e.Modifiers != 0 || v.shortcuts == QuestionnaireShortcutsOff {
		return false
	}
	name := strings.ToUpper(e.Name)
	index := -1
	if len(name) == 1 {
		if v.shortcuts == QuestionnaireShortcutsLetters && name[0] >= 'A' && name[0] <= 'Z' {
			index = int(name[0] - 'A')
		}
		if v.shortcuts == QuestionnaireShortcutsNumbers && name[0] >= '1' && name[0] <= '9' {
			index = int(name[0] - '1')
		}
	}
	options := v.enabledOptions(q)
	if index < 0 || index >= len(options) || c.radio == nil && c.checks == nil {
		return false
	}
	if e.State != el.KeyPress {
		delete(v.heldShortcuts, name)
		return true
	}
	if v.heldShortcuts[name] {
		return true
	}
	if v.heldShortcuts == nil {
		v.heldShortcuts = map[string]bool{}
	}
	v.heldShortcuts[name] = true
	index = options[index]
	if c.radio != nil {
		c.radio.SetValue(q.Options[index])
		c.freeformActive = false
	} else {
		c.checks[index].SetValue(!c.checks[index].Value())
	}
	v.changed(q.ID)
	return true
}
