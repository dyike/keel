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
}

// Answer holds the answer to one question; only the field for its kind is set.
type Answer struct {
	Text    string   // Single, Text, LongText
	Choices []string // Multiple, in option order
	Rating  int      // Rating; 0 means not rated
}

// Empty reports whether the question was left unanswered.
func (a Answer) Empty() bool {
	return strings.TrimSpace(a.Text) == "" && len(a.Choices) == 0 && a.Rating == 0
}

type questionControl struct {
	radio  *RadioGroupView
	checks []*CheckboxView
	input  *InputView
	rating *RatingView
}

// QuestionnaireView asks questions one page at a time with a progress bar.
// 下一题 checks that a required question is answered; 提交 on the last page
// checks every question, jumps to the first unanswered one, and otherwise
// calls OnSubmit with all answers by question ID.
type QuestionnaireView struct {
	questions []Question
	controls  map[string]*questionControl
	page      int
	disabled  bool
	err       string
	onSubmit  func(map[string]Answer)
}

func Questionnaire(questions ...Question) *QuestionnaireView {
	return &QuestionnaireView{questions: questions, controls: map[string]*questionControl{}}
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
}

func (v *QuestionnaireView) control(q Question) *questionControl {
	c := v.controls[q.ID]
	if c != nil {
		return c
	}
	c = &questionControl{}
	clear := func() { v.err = "" }
	switch q.Kind {
	case QuestionSingle:
		c.radio = RadioGroup("", q.Options...).OnChange(func(string) { clear() })
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
	v.controls[q.ID] = c
	return c
}

func (v *QuestionnaireView) answer(q Question) Answer {
	c := v.control(q)
	switch q.Kind {
	case QuestionSingle:
		return Answer{Text: c.radio.Value()}
	case QuestionMultiple:
		var out []string
		for i, ch := range c.checks {
			if ch.Value() {
				out = append(out, q.Options[i])
			}
		}
		return Answer{Choices: out}
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
	for _, q := range v.questions {
		a, ok := answers[q.ID]
		if !ok {
			continue
		}
		c := v.control(q)
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

func (v *QuestionnaireView) missing(i int) bool {
	q := v.questions[i]
	return q.Required && v.answer(q).Empty()
}

func (v *QuestionnaireView) next() {
	if v.disabled {
		return
	}
	if v.missing(v.page) {
		v.err = locale.Current().Required
		return
	}
	v.SetPage(v.page + 1)
}

func (v *QuestionnaireView) submit() {
	if v.disabled {
		return
	}
	for i := range v.questions {
		if v.missing(i) {
			v.SetPage(i)
			v.err = locale.Current().Required
			return
		}
	}
	if v.onSubmit != nil {
		v.onSubmit(v.Value())
	}
}

func (v *QuestionnaireView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	n := len(v.questions)
	if n == 0 {
		return el.Div().Hidden(true)
	}
	v.page = min(v.page, n-1)
	q := v.questions[v.page]
	progress := Progress(text.Progress(v.page+1, n))
	progress.SetValue(float32(v.page+1) / float32(n))
	title := el.Div().Row().Gap(4).Child(el.Text(q.Title).TextSize(17).Bold())
	if q.Required {
		title.Child(el.Text("*").TextSize(17).TextColor(theme.DangerText))
	}
	card := el.Div().Role("group").Name(q.Title).Gap(12).Items(el.Stretch).Child(title)
	if q.Description != "" {
		card.Child(el.Text(q.Description).TextSize(13).TextColor(theme.Muted))
	}
	c := v.control(q)
	switch {
	case c.radio != nil:
		card.Child(c.radio.Render(cx))
	case c.checks != nil:
		list := el.Div().Gap(8).Items(el.Start)
		for _, ch := range c.checks {
			list.Child(ch.Render(cx))
		}
		card.Child(list)
	case c.rating != nil:
		card.Child(c.rating.Render(cx))
	default:
		card.Child(c.input.Render(cx))
	}
	if v.err != "" {
		card.Child(el.Text(v.err).TextSize(12).TextColor(theme.DangerText))
	}
	prev := Button(text.Previous, func() { v.SetPage(v.page - 1) }).Variant(ButtonSecondary)
	prev.SetDisabled(v.page == 0)
	forward := Button(text.Next, v.next)
	if v.page == n-1 {
		forward = Button(text.Submit, v.submit)
	}
	return el.Div().Disabled(v.disabled).Role("form").Name(text.Progress(v.page+1, n)).Gap(20).Items(el.Stretch).Child(
		progress.Render(cx),
		card,
		el.Div().Row().Gap(8).Justify(el.End).Child(prev.Render(cx), forward.Render(cx)),
	)
}
