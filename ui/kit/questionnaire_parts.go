package kit

import "github.com/dyike/keel/ui/el"

// QuestionnaireParts contains fresh elements backed by the questionnaire's
// existing controls and commands. Optional parts are nil. Insert each element
// at most once and do not retain it after the current Render.
type QuestionnaireParts struct {
	Progress, Title, Description el.Element
	Answer, Freeform, Error      el.Element
	Previous, Skip, Forward      el.Element
}

// QuestionnaireContext is an owned snapshot for custom presentation. Question
// option/default-answer slices and Answer choices are independent of internal state.
type QuestionnaireContext struct {
	Size     QuestionnaireSize
	Question Question
	Answer   Answer
	Progress QuestionnaireProgress
	Page     int
	Error    string
	Disabled bool
}

// Layout replaces the arrangement of the current question's parts. The outer
// form retains disabled inheritance, keyboard navigation and focus management.
// Return nil, or pass nil, to use the default layout. Keep answer and navigation
// parts in the returned tree to retain their built-in interactions. Omitting
// parts intentionally hides those controls; it does not disable their commands.
func (v *QuestionnaireView) Layout(fn func(*el.Context, QuestionnaireContext, QuestionnaireParts) el.Element) *QuestionnaireView {
	v.layout = fn
	return v
}
