package kit

import (
	"slices"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// QuestionKind is the kind of answer a Question takes.
type QuestionKind uint8

const (
	QuestionSingle   QuestionKind = iota // one of Options
	QuestionMultiple                     // any of Options
	QuestionText                         // one line of text
	QuestionLongText                     // several lines of text
	QuestionRating                       // 1..Scale stars
)

// Question is one question of a questionnaire. IDs must be unique.
type Question struct {
	ID, Title, Description string
	Kind                   QuestionKind
	Options                []string // Single and Multiple
	Scale                  int      // Rating: number of stars, 5 by default
	Required               bool
	Disabled               bool
	FreeformLabel          string // Single/Multiple: optional freeform input.
	Validate               func(Answer, map[string]Answer) string
}

// Answer holds a typed answer, optional active freeform text, or an intentional skip.
type Answer struct {
	Text     string   // Single, Text, LongText
	Choices  []string // Multiple, in option order
	Rating   int      // Rating; 0 means not rated
	Freeform string   // Additional active text for choice questions.
	Skipped  bool
}

// Empty reports whether the question was left unanswered.
func (a Answer) Empty() bool {
	return strings.TrimSpace(a.Freeform) == "" && strings.TrimSpace(a.Text) == "" && len(a.Choices) == 0 && a.Rating == 0
}

type questionControl struct {
	radio          *RadioGroupView
	checks         []*CheckboxView
	input          *InputView
	rating         *RatingView
	freeform       *InputView
	freeformActive bool
}

// QuestionnaireView asks questions one page at a time with a progress bar.
// 下一题 checks that a required question is answered; 提交 on the last page
// checks every question, jumps to the first unanswered one, and otherwise
// calls OnSubmit with all answers by question ID.
type QuestionnaireView struct {
	questions     []Question
	controls      map[string]*questionControl
	page          int
	disabled      bool
	err           string
	onSubmit      func(map[string]Answer)
	external      map[string]string
	skipped       map[string]bool
	completed     bool
	onChange      func(string, Answer)
	onComplete    func(map[string]Answer)
	shortcuts     QuestionnaireShortcuts
	focusPending  bool
	heldShortcuts map[string]bool
}

func Questionnaire(questions ...Question) *QuestionnaireView {
	owned := slices.Clone(questions)
	ids := make(map[string]bool, len(questions))
	for i, q := range owned {
		if q.ID == "" || ids[q.ID] {
			panic("kit.Questionnaire: empty or duplicate question ID")
		}
		if q.Kind > QuestionRating {
			panic("kit.Questionnaire: invalid question kind")
		}
		ids[q.ID] = true
		owned[i].Options = slices.Clone(q.Options)
		seen := map[string]bool{}
		for _, option := range q.Options {
			if seen[option] {
				panic("kit.Questionnaire: duplicate option")
			}
			seen[option] = true
		}
	}
	v := &QuestionnaireView{questions: owned, controls: map[string]*questionControl{}, external: map[string]string{}, skipped: map[string]bool{}}
	v.SetPage(0)
	v.focusPending = false
	return v
}
func (v *QuestionnaireView) OnSubmit(fn func(answers map[string]Answer)) *QuestionnaireView {
	v.onSubmit = fn
	return v
}

// SetDisabled disables every answer control, navigation, and submission.
// SetPage and SetValue remain available for programmatic updates.
func (v *QuestionnaireView) SetDisabled(on bool) { v.disabled = on }

// Page is the index of the question shown; SetPage moves to another.
func (v *QuestionnaireView) Page() int { return v.page }
func (v *QuestionnaireView) SetPage(i int) {
	v.page, v.err = min(max(i, 0), max(len(v.questions)-1, 0)), ""
	if len(v.questions) > 0 && v.questions[v.page].Disabled {
		found := -1
		for j := v.page; j < len(v.questions); j++ {
			if !v.questions[j].Disabled {
				found = j
				break
			}
		}
		if found < 0 {
			for j := v.page - 1; j >= 0; j-- {
				if !v.questions[j].Disabled {
					found = j
					break
				}
			}
		}
		v.page = found
	}
	v.focusPending = true
}

func (v *QuestionnaireView) control(q Question) *questionControl {
	c := v.controls[q.ID]
	if c != nil {
		return c
	}
	c = &questionControl{}
	clear := func() { v.changed(q.ID) }
	switch q.Kind {
	case QuestionSingle:
		c.radio = RadioGroup("", q.Options...).OnChange(func(string) { c.freeformActive = false; clear() })
		c.radio.setName(q.Title)
	case QuestionMultiple:
		for _, o := range q.Options {
			c.checks = append(c.checks, Checkbox(o, false).OnChange(func(bool) { clear() }))
		}
	case QuestionText:
		c.input = Input("").OnChange(func(string) { clear() })
		c.input.setName(q.Title)
	case QuestionLongText:
		c.input = TextArea("").Rows(4).OnChange(func(string) { clear() })
		c.input.setName(q.Title)
	case QuestionRating:
		scale := q.Scale
		if scale < 1 {
			scale = 5
		}
		c.rating = Rating("", scale).OnChange(func(int) { clear() })
		c.rating.setName(q.Title)
	}
	if q.FreeformLabel != "" && (q.Kind == QuestionSingle || q.Kind == QuestionMultiple) {
		c.freeform = Input(q.FreeformLabel).OnChange(func(string) {
			c.freeformActive = true
			if c.radio != nil {
				c.radio.SetValue("")
			}
			clear()
		})
	}
	v.controls[q.ID] = c
	return c
}

func (v *QuestionnaireView) answer(q Question) Answer {
	if v.skipped[q.ID] {
		return Answer{Skipped: true}
	}
	c := v.control(q)
	switch q.Kind {
	case QuestionSingle:
		if c.freeformActive && c.freeform != nil {
			return Answer{Freeform: c.freeform.Value()}
		}
		return Answer{Text: c.radio.Value()}
	case QuestionMultiple:
		var out []string
		for i, ch := range c.checks {
			if ch.Value() {
				out = append(out, q.Options[i])
			}
		}
		a := Answer{Choices: out}
		if c.freeform != nil {
			a.Freeform = c.freeform.Value()
		}
		return a
	case QuestionRating:
		return Answer{Rating: c.rating.Value()}
	}
	return Answer{Text: c.input.Value()}
}

// Value returns every answer so far, by question ID.
func (v *QuestionnaireView) Value() map[string]Answer {
	out := make(map[string]Answer, len(v.questions))
	for _, q := range v.questions {
		out[q.ID] = v.answer(q)
	}
	return out
}

// SetValue fills in answers without submitting; unknown IDs are ignored.
func (v *QuestionnaireView) SetValue(answers map[string]Answer) {
	v.completed = false
	for _, q := range v.questions {
		a, ok := answers[q.ID]
		if !ok {
			continue
		}
		c := v.control(q)
		v.skipped[q.ID] = a.Skipped && !q.Required
		if c.freeform != nil {
			c.freeform.SetValue(a.Freeform)
			c.freeformActive = strings.TrimSpace(a.Freeform) != ""
		}
		switch q.Kind {
		case QuestionSingle:
			c.radio.SetValue(a.Text)
		case QuestionMultiple:
			for i, ch := range c.checks {
				ch.SetValue(slices.Contains(a.Choices, q.Options[i]))
			}
		case QuestionRating:
			c.rating.SetValue(a.Rating)
		default:
			c.input.SetValue(a.Text)
		}
	}
}

func (v *QuestionnaireView) next() {
	if v.disabled || v.page < 0 || v.page >= len(v.questions) {
		return
	}
	if msg := v.validate(v.page); msg != "" {
		v.err = msg
		v.focusPending = true
		return
	}
	for i := v.page + 1; i < len(v.questions); i++ {
		if !v.questions[i].Disabled {
			v.SetPage(i)
			return
		}
	}
	v.submit()
}
func (v *QuestionnaireView) submit() {
	if v.disabled {
		return
	}
	for i, q := range v.questions {
		if q.Disabled {
			continue
		}
		if msg := v.validate(i); msg != "" {
			v.SetPage(i)
			v.err = msg
			return
		}
	}
	if !v.completed {
		v.completed = true
		if v.onComplete != nil {
			v.onComplete(v.enabledAnswers())
		}
	}
	if v.completed && v.onSubmit != nil {
		v.onSubmit(v.enabledAnswers())
	}
}

func (v *QuestionnaireView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	n := len(v.questions)
	if n == 0 || v.Progress().Total == 0 {
		return el.Div().Hidden(true)
	}
	if v.page < 0 || v.page >= n || v.questions[v.page].Disabled {
		v.SetPage(max(v.page, 0))
	}
	q := v.questions[v.page]
	state := v.Progress()
	progress := Progress(text.Progress(state.Current, state.Total))
	progress.SetValue(float32(state.Current) / float32(state.Total))
	title := el.Div().Row().Gap(theme.SpaceXs).Child(el.Text(q.Title).TextSize(theme.TextLg).Bold())
	if q.Required {
		title.Child(el.Text("*").TextSize(theme.TextLg).TextColor(theme.DangerText))
	}
	card := el.Div().Role("group").Name(q.Title).Gap(theme.SpaceLg).Items(el.Stretch).Child(title)
	if q.Description != "" {
		card.Child(el.Text(q.Description).TextSize(theme.TextMd).TextColor(theme.Muted))
	}
	c := v.control(q)
	switch {
	case c.radio != nil:
		card.Child(c.radio.Render(cx))
	case q.Kind == QuestionMultiple:
		list := el.Div().Gap(theme.SpaceMd).Items(el.Start)
		for _, ch := range c.checks {
			list.Child(ch.Render(cx))
		}
		card.Child(list)
	case c.rating != nil:
		card.Child(c.rating.Render(cx))
	default:
		card.Child(c.input.Render(cx))
	}
	if c.freeform != nil {
		card.Child(c.freeform.Render(cx))
	}
	message := v.err
	if v.external[q.ID] != "" {
		message = v.external[q.ID]
	}
	if message != "" {
		card.Child(el.Text(message).TextSize(theme.TextSm).TextColor(theme.DangerText))
	}
	if v.focusPending {
		target := ""
		switch {
		case c.freeformActive && c.freeform != nil:
			target = c.freeform.FocusID()
		case c.input != nil:
			target = c.input.FocusID()
		case c.radio != nil:
			target = c.radio.FocusID()
		case c.rating != nil:
			target = c.rating.FocusID()
		case len(c.checks) > 0:
			target = c.checks[0].FocusID()
		}
		if target != "" && cx.Enabled(target) {
			cx.Focus(target)
			v.focusPending = false
		} else {
			cx.AfterEnabled(autoID("questionnaire", v), v, 0, func() {})
		}
	}
	prev := Button(text.Previous, v.previous).Variant(ButtonSecondary)
	prev.SetDisabled(state.Current <= 1)
	forward := Button(text.Next, v.next)
	if state.Current == state.Total {
		forward = Button(text.Submit, v.submit)
	}
	actions := el.Div().Row().Gap(theme.SpaceMd).Justify(el.End).Child(prev.Render(cx))
	if !q.Required {
		actions.Child(Button(text.Skip, v.Skip).Variant(ButtonGhost).Render(cx))
	}
	actions.Child(forward.Render(cx))
	return el.Div().ID(autoID("questionnaire", v)).Disabled(v.disabled).Role("form").Name(text.Progress(state.Current, state.Total)).Focusable(true).OnKey(func(e el.KeyEvent) bool { return v.handleKey(cx, e) }).Gap(20).Items(el.Stretch).Child(progress.Render(cx), card, actions)
}
