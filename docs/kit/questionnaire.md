# Questionnaire

English | [简体中文](questionnaire.zh-CN.md)

One page questionnaire.

```go
q := kit.Questionnaire(
    kit.Question{ID: "role", Title: "What is your role?", Kind: kit.QuestionSingle, Options: roles, Required: true},
    kit.Question{ID: "tools", Title: "Tools you use", Kind: kit.QuestionMultiple, Options: tools},
    kit.Question{ID: "score", Title: "Satisfaction", Kind: kit.QuestionRating, Scale: 5, Required: true},
    kit.Question{ID: "note", Title: "Suggestions", Kind: kit.QuestionLongText},
).OnSubmit(func(a map[string]kit.Answer) { save(a) })
```

- Question types: `QuestionSingle` (single choice), `QuestionMultiple` (multiple choices), `QuestionText` (single line of text), `QuestionLongText` (multiple lines of text), `QuestionRating` (star rating, default 5 stars).
- At the top is a progress bar that displays "Question 3 / 10". When clicking "Next Question", if the required questions are not answered, an error will be displayed and the page cannot be turned; the error will disappear automatically after being answered.
- "Submit" on the last page will verify all enabled questions; if there is no answer, no active skipping, or a verification error, it will jump to the first wrong question, and call `OnSubmit` after passing it.
- The answer is represented by `Answer`. According to the question type, fill in `Text`, `Choices` (in order of options), `Rating`. You can also add `Freeform` or `Skipped`; `Answer.Empty()` determines whether to answer.
- `Value()` returns all current answers; `SetValue` backfills answers, such as restoring a draft; `Page()` / `SetPage(i)` reads or switches the current page.
- "Previous question", "next question", "submit", progress text, and required prompts all come from the locale.

Agent: container role `form`, the name is the progress text; the current topic is `group` named after the topic title, and the control is named after the topic title.

Verify: `go run ./examples/components -section questionnaire`, add `-theme dark` to check the dark theme.

`SetDisabled(true)` disables the current answer control, page forward and backward, and submission, and does not trigger verification or `OnSubmit`; the program can still call `SetPage` and `SetValue`. The current page and answers will be retained after re-enabling.

The questions and options are copied during construction, and modifying the original slices later will not change the questionnaire. Question ID must be non-empty and unique; unsupported question types will panic during construction. `Value` and `OnSubmit` provide independent answer snapshots, and modifying the returned map or multi-select answer slice will not affect the internal state.

## Conditional questions, skip and completion status

`Question.Disabled` sets the initial disabling condition, `SetQuestionDisabled(id, on)` is dynamically updated, and unknown ID returns false. Disable the question to exit navigation, progress, verification and submission, but keep the draft, which can still be read by `Value`. When the current question is disabled, first find the available questions at the back, and then the previous questions; when all are disabled, the Page is -1 and the component is hidden. The application can update the conditions in `OnAnswerChange(func(id string, answer kit.Answer))` based on the answers to the previous questions.

Optional questions now need to be answered or explicitly skipped, and "not answered yet" cannot be implicitly regarded as completed. `Skip()` / "Skip" button clears the answers to the current optional questions, saves `Answer.Skipped=true` and moves forward; skip the last question and try to submit. Required questions cannot be skipped. Returning to a skipped question and answering it again will clear the skipped status. `SetValue` supports restoring the Skipped status of optional questions.

`Progress()` returns a snapshot of the currently enabled question set: Current starts from 1, Total, Answered, Skipped, and Unanswered are the total number of questions and the number of three answering states respectively. Completed means that a valid submission has been completed. `OnComplete` is only executed when entering the completion state and precedes OnSubmit in the first valid submission; OnSubmit is still called for each valid submission. Modifying answers, conditions, or external errors invalidates the completion status. The commit snapshot contains only enabled questions; Value still contains all questions.

## Free input and verification

`FreeformLabel` can be set for single-choice and multiple-choice questions to display a free input box with a name. `Answer.Freeform` Saves valid free text; blank text does not count as an answer. Single-choice questions switch between options and free input. When the fixed option is selected, the free input draft is retained but is not submitted as a valid answer; the free input takes effect when re-entered. Multiple choice questions allow Choices and Freeform to be answered simultaneously.

`Question.Validate func(Answer, map[string]Answer) string` performs synchronous custom verification, receives the current question answer and the independent answer snapshot of the enabled question, and returns an empty string to indicate passing. The next question verifies the current question, submits all enabled questions for verification and locates the first wrong question; the required/unanswered check is automatically skipped and does not run the custom verification before the custom verification. The callback should only check the snapshot and not modify the questionnaire.

`SetExternalError(id, message)` saves server or application errors and clears empty strings; `ExternalError(id)` queries. External errors are displayed first and blocked from passing, and neither input nor Reset are automatically cleared. The application is responsible for clearing a new answer after accepting it and not automatically initiating network verification.

`Question.DefaultAnswer` Defines the default answer on construction, Choices will be copied. `Reset()` Restores this default answer, clears internal errors and completion status, and returns to the first enabled question; retains external errors and current disabling conditions. Subsequent SetValue only changes the current answer and does not move the Reset baseline.

## Options shortcut keys

`Shortcuts(kit.QuestionnaireShortcutsLetters)` turns on A–Z, `QuestionnaireShortcutsNumbers` turns on 1–9, default is Off. The shortcut keys are mapped in the order of the current question's undisabled options, single-select selection, and multi-selection switching; they are only processed when the questionnaire gets focus and the focus is not in text input, and keys with modifier keys are not processed. Pressing and holding the button will not switch repeatedly, and it can be triggered again after releasing it. Default keyboard navigation supports left and right keys to move forward and backward through questions, up and down keys to move focus between multi-select controls/free input, Enter to confirm the current question that has been answered or skipped, and Cmd/Ctrl+Enter to confirm from text input. Text input retains the normal direction keys and Enter; the radio group and rating controls retain their own direction key behaviors; the operation buttons retain Enter activation. Empty answers are not implicitly committed by Enter. `KeyboardNavigation(false)` turns off these navigation bindings, independent of the option shortcut key configuration. Option text should be prompted by the application to correspond to the shortcut key.

Automatic test coverage condition navigation, disabled question exclusion, skip, resubmit, completion event, external error/Reset, custom verification, free input retention and option shortcut keys; real device vision/input method acceptance has not yet been completed.


`SetChoiceDisabled(questionID, option, on)` Dynamically disables fixed options and returns false for unknown questions/options. Disabled items retain internal selection drafts, but exit valid answers, verification snapshots, and shortcut key numbers; re-enabling restores valid answers, and condition changes clear the completion status. The all-question-required rule still applies: if all options are disabled and no free input is available, the application should also disable the question to avoid being unable to answer it. The program SetValue can restore a draft of disabled items, but Value/Submit contains only available items.

The new automatic test covers default answer ownership/Reset, disabled item valid answers and shortcut key renumbering, text Cmd/Ctrl+Enter, ordinary arrow protection, multi-select focus movement and closed navigation; real device input method and visual acceptance are not done.

## Custom combination layout

`Layout(func(cx, state, parts) el.Element)` rearranges the components of the current question. `QuestionnaireParts` Provides Progress, Title, Description, Answer, Freeform, Error, Previous, Skip, Forward; nil if not applicable. Return nil or `Layout(nil)` to restore the default arrangement.

These elements use the same answer control as the original check/skip/submit commands. The outer form always remains disabled with inheritance, keyboard navigation, and focus management disabled. Each element can only be placed into this frame's tree once and cannot be saved across frames; Answer and navigation buttons should generally be retained. Omitting a component only hides the corresponding interface and does not close the shortcut key command. When a custom layout changes the tree path of a control, the input focus and selection may be reset.

`QuestionnaireContext` provides the current Question, Answer, Progress, Page starting from scratch, error text and the questionnaire itself Disabled; question options, default answers and slices of the current answer are all copies. Do not modify the questionnaire status in the layout callback. Modifications should be placed in the event callback. The component library provides switching between default layout and custom cards. The automatic test covers required verification, input, disabled submission, context copy and default layout restoration after combination; the native input method is not accepted.

## Uniform size

`Size(QuestionnaireSizeXSmall/Small/Medium/Large)` Adjust titles, descriptions, error font sizes, option tags and text, rating stars, input boxes, widget spacing, and navigation buttons. By default, Medium retains the original layout; size switching retains answers and control instances. The custom Layout can read the gear position from `QuestionnaireContext.Size`, the built-in components provided have applied dimensions, and can be configured by adding new containers and controls. The component library provides cycle switching buttons; the automatic test covers five question types with double rates, four-level navigation buttons, answer retention, omitted parts and recovery input. Native vision is not accepted.
