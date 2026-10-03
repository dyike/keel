package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestQuestionnaireConditionsSkipErrorsAndCompletion(t *testing.T) {
	completed, submitted := 0, 0
	q := Questionnaire(Question{ID: "a", Title: "a", Kind: QuestionText, Required: true, Validate: func(a Answer, _ map[string]Answer) string {
		if len(a.Text) < 3 {
			return "too short"
		}
		return ""
	}}, Question{ID: "b", Title: "b", Kind: QuestionText}, Question{ID: "c", Title: "c", Required: true, Disabled: true})
	q.OnComplete(func(map[string]Answer) { completed++ }).OnSubmit(func(a map[string]Answer) {
		submitted++
		if _, exists := a["c"]; exists {
			t.Fatal("disabled submitted")
		}
	})
	q.SetValue(map[string]Answer{"a": {Text: "ok"}})
	q.next()
	if q.Page() != 0 || q.err != "too short" {
		t.Fatal("custom validation")
	}
	q.SetValue(map[string]Answer{"a": {Text: "good"}})
	q.next()
	q.submit()
	if submitted != 0 || q.Page() != 1 {
		t.Fatal("optional unanswered submitted")
	}
	q.Skip()
	if completed != 1 || submitted != 1 || !q.Value()["b"].Skipped {
		t.Fatal("skip/complete")
	}
	q.submit()
	if completed != 1 || submitted != 2 {
		t.Fatal("completion emitted twice")
	}
	q.SetExternalError("a", "server error")
	q.submit()
	if q.Page() != 0 || q.err != "server error" {
		t.Fatal("external error")
	}
	q.Reset()
	if q.ExternalError("a") != "server error" {
		t.Fatal("reset lost external error")
	}
	q.SetQuestionDisabled("a", true)
	p := q.Progress()
	if p.Total != 1 || p.Current != 1 || q.Page() != 1 {
		t.Fatal("disabled navigation", p, q.Page())
	}
	q.SetQuestionDisabled("b", true)
	if q.Progress().Total != 0 || q.Page() != -1 {
		t.Fatal("all disabled")
	}
	q.SetQuestionDisabled("a", false)
	if q.Page() != 0 || q.Progress().Total != 1 {
		t.Fatal("reenable")
	}
}

func TestQuestionnaireFreeformAndShortcuts(t *testing.T) {
	q := Questionnaire(Question{ID: "q", Title: "choice", Options: []string{"A", "B"}, FreeformLabel: "other"}).Shortcuts(QuestionnaireShortcutsNumbers)
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return q.Render(c) })
	clickClass(t, h, "Editor", "other")
	h.Type("custom")
	h.Frame()
	if q.Value()["q"].Freeform != "custom" {
		t.Fatal("freeform", q.Value()["q"])
	}
	click(t, h, "B")
	h.Frame()
	if q.Value()["q"].Text != "B" || q.Value()["q"].Freeform != "" || q.control(q.questions[0]).freeform.Value() != "custom" {
		t.Fatal("choice/freeform draft")
	}
	cx.Focus(autoID("questionnaire", q))
	h.Frame()
	h.Key("1", 0)
	if q.Value()["q"].Text != "A" {
		t.Fatal("choice shortcut")
	}
	clickClass(t, h, "Editor", "other")
	h.Key("2", 0)
	if q.Value()["q"].Text != "A" {
		t.Fatal("shortcut intercepted text input")
	}
	q.SetExternalError("q", "external")
	h.Type("x")
	h.Frame()
	if !shown(h, "external") {
		t.Fatal("edit cleared external error")
	}
	multi := Questionnaire(Question{ID: "m", Kind: QuestionMultiple, Options: []string{"A", "B"}, FreeformLabel: "other"})
	multi.SetValue(map[string]Answer{"m": {Choices: []string{"B"}, Freeform: "extra"}})
	a := multi.Value()["m"]
	if len(a.Choices) != 1 || a.Freeform != "extra" {
		t.Fatal("mixed multiple")
	}
	// Modified key presses remain untouched.
	if q.handleKey(cx, el.KeyEvent{Name: "1", Modifiers: key.ModShortcut, State: el.KeyPress}) {
		t.Fatal("modified key swallowed")
	}
}
