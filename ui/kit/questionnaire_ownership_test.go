package kit

import "testing"

func TestQuestionnaireOwnsQuestionsAndAnswers(t *testing.T) {
	options := []string{"A", "B"}
	questions := []Question{{ID: "q", Title: "title", Kind: QuestionMultiple, Options: options}}
	var submitted map[string]Answer
	q := Questionnaire(questions...).OnSubmit(func(a map[string]Answer) { submitted = a })
	options[0] = "changed"
	questions[0].ID = "changed"
	questions[0].Options = nil
	q.SetValue(map[string]Answer{"q": {Choices: []string{"A"}}})
	value := q.Value()
	if len(value["q"].Choices) != 1 || value["q"].Choices[0] != "A" {
		t.Fatal("question data retained caller aliases")
	}
	value["q"].Choices[0] = "B"
	delete(value, "q")
	if q.Value()["q"].Choices[0] != "A" {
		t.Fatal("Value exposed internal answer")
	}
	q.submit()
	submitted["q"].Choices[0] = "B"
	if q.Value()["q"].Choices[0] != "A" {
		t.Fatal("OnSubmit exposed internal answer")
	}
}

func TestQuestionnaireRejectsInvalidIdentityAndKind(t *testing.T) {
	for _, questions := range [][]Question{
		{{ID: ""}},
		{{ID: "same"}, {ID: "same"}},
		{{ID: "invalid", Kind: QuestionKind(255)}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("invalid questions accepted: %+v", questions)
				}
			}()
			Questionnaire(questions...)
		}()
	}
}
