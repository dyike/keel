package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestQuestionnaireCustomLayoutValidationAndOwnership(t *testing.T) {
	submitted := 0
	q := Questionnaire(Question{ID: "a", Title: "answer", Kind: QuestionText, Required: true}, Question{ID: "b", Title: "choices", Kind: QuestionMultiple, Options: []string{"one", "two"}, DefaultAnswer: Answer{Choices: []string{"one"}}}).OnSubmit(func(map[string]Answer) { submitted++ })
	q.Layout(func(cx *el.Context, s QuestionnaireContext, p QuestionnaireParts) el.Element {
		if len(s.Question.Options) > 0 {
			s.Question.Options[0] = "mutated"
			s.Question.DefaultAnswer.Choices[0] = "mutated"
			s.Answer.Choices[0] = "mutated"
		}
		return el.Div().Gap(8).Child(p.Title, p.Answer, p.Error, el.Div().Row().Child(p.Previous, p.Skip, p.Forward))
	})
	h := render(func(cx *el.Context) el.Element { return q.Render(cx) })
	click(t, h, locale.Current().Next)
	h.Frame()
	if q.Page() != 0 || q.err == "" {
		t.Fatal("custom layout bypassed validation")
	}
	clickClass(t, h, "Editor", "answer")
	h.Type("hello")
	h.Frame()
	q.Size(QuestionnaireSizeLarge)
	h.Frame()
	if q.Value()["a"].Text != "hello" {
		t.Fatal("size lost answer")
	}
	click(t, h, locale.Current().Next)
	h.Frame()
	if q.Page() != 1 || q.questions[1].Options[0] != "one" || q.Value()["b"].Choices[0] != "one" {
		t.Fatal("layout snapshot mutated model")
	}
	q.SetDisabled(true)
	h.Frame()
	click(t, h, locale.Current().Submit)
	h.Frame()
	if submitted != 0 {
		t.Fatal("disabled custom submit")
	}
	q.SetDisabled(false)
	h.Frame()
	click(t, h, locale.Current().Submit)
	h.Frame()
	if submitted != 1 {
		t.Fatal("custom submit unavailable", submitted)
	}
	q.Layout(func(*el.Context, QuestionnaireContext, QuestionnaireParts) el.Element { return nil })
	h.Frame()
	if !shown(h, "choices") || !shown(h, "one") {
		t.Fatal("nil fallback missing")
	}
	q.Reset()
	h.Frame()
	if q.Value()["a"].Text != "" {
		t.Fatal("reset after layout change")
	}
}

func TestQuestionnairePresentationSizesAndOmittedParts(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, kind := range []QuestionKind{QuestionSingle, QuestionMultiple, QuestionText, QuestionLongText, QuestionRating} {
			q := Questionnaire(Question{ID: "a", Title: "first", Kind: kind, Options: []string{"one", "two"}}, Question{ID: "b", Title: "second", Kind: QuestionText})
			h := renderView(q, 480, scale)
			for _, size := range []QuestionnaireSize{QuestionnaireSizeXSmall, QuestionnaireSizeSmall, QuestionnaireSizeMedium, QuestionnaireSizeLarge} {
				q.Size(size)
				h.Frame()
				b := bounds(h, locale.Current().Next)
				if b.Empty() || b.Dy() != int(q.sizeMetrics().button)*scale {
					t.Fatal("navigation size", kind, size, scale, b)
				}
			}
			q.Layout(func(_ *el.Context, _ QuestionnaireContext, p QuestionnaireParts) el.Element {
				return el.Div().Child(p.Title, p.Previous, p.Skip, p.Forward)
			})
			h.Frame()
			click(t, h, locale.Current().Skip)
			h.Frame()
			if q.Page() != 1 || !q.Value()["a"].Skipped {
				t.Fatal("omitted answer prevented skip")
			}
			q.Layout(nil)
			h.Frame()
			clickClass(t, h, "Editor", "second")
			h.Type("usable")
			h.Frame()
			if q.Value()["b"].Text != "usable" {
				t.Fatal("restoring answer lost input")
			}
		}
	}
}
