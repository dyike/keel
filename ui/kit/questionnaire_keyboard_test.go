package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"testing"
)

func TestQuestionnaireDisabledChoicesDefaultsAndShortcutOrder(t *testing.T) {
	defaults := []string{"A", "C"}
	q := Questionnaire(Question{ID: "q", Title: "q", Kind: QuestionMultiple, Options: []string{"A", "B", "C"}, DefaultAnswer: Answer{Choices: defaults}}).Shortcuts(QuestionnaireShortcutsNumbers)
	defaults[0] = "B"
	if !q.SetChoiceDisabled("q", "A", true) || q.SetChoiceDisabled("q", "missing", true) {
		t.Fatal("schema lookup")
	}
	if !slices.Equal(q.Value()["q"].Choices, []string{"C"}) {
		t.Fatal("disabled answer")
	}
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return q.Render(c) })
	h.Frame()
	cx.Focus(autoID("questionnaire", q))
	h.Frame()
	h.Key("1", 0)
	if !slices.Equal(q.Value()["q"].Choices, []string{"B", "C"}) {
		t.Fatal("shortcut should number only enabled choices", q.Value())
	}
	q.SetChoiceDisabled("q", "A", false)
	if !slices.Equal(q.Value()["q"].Choices, []string{"A", "B", "C"}) {
		t.Fatal("reenable lost draft")
	}
	q.SetValue(map[string]Answer{"q": {Choices: []string{"B"}}})
	q.Reset()
	if !slices.Equal(q.Value()["q"].Choices, []string{"A", "C"}) {
		t.Fatal("default snapshot", q.Value())
	}
}

func TestQuestionnaireNavigationProtectsEditorsAndActions(t *testing.T) {
	submitted := 0
	q := Questionnaire(Question{ID: "a", Title: "text", Kind: QuestionLongText, DefaultAnswer: Answer{Text: "hello"}}, Question{ID: "b", Title: "choices", Kind: QuestionMultiple, Options: []string{"A", "B"}}).OnSubmit(func(map[string]Answer) { submitted++ })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return q.Render(c) })
	clickClass(t, h, "Editor", "text")
	h.Key(key.NameRightArrow, 0)
	if q.Page() != 0 {
		t.Fatal("text arrow navigated")
	}
	h.Key(key.NameReturn, key.ModShortcut)
	h.Frame()
	if q.Page() != 1 {
		t.Fatal("modified enter did not confirm text", q.Page())
	}
	h.Frame()
	cx.Focus(autoID("questionnaire", q))
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameSpace, 0)
	if !slices.Equal(q.Value()["b"].Choices, []string{"A"}) {
		t.Fatal("down did not focus first checkbox")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameSpace, 0)
	if !slices.Equal(q.Value()["b"].Choices, []string{"A", "B"}) {
		t.Fatal("down did not move checkbox focus")
	}
	cx.Focus(autoID("questionnaire", q))
	h.Frame()
	h.Key(key.NameReturn, 0)
	if submitted != 1 {
		t.Fatal("enter confirm", submitted)
	}
	q.KeyboardNavigation(false)
	h.Frame()
	h.Key(key.NameLeftArrow, 0)
	if q.Page() != 1 {
		t.Fatal("disabled navigation still active")
	}
}
