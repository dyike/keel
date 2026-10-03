# Questionnaire

一页一题的问卷。

```go
q := kit.Questionnaire(
    kit.Question{ID: "role", Title: "你的角色？", Kind: kit.QuestionSingle, Options: roles, Required: true},
    kit.Question{ID: "tools", Title: "常用工具", Kind: kit.QuestionMultiple, Options: tools},
    kit.Question{ID: "score", Title: "满意度", Kind: kit.QuestionRating, Scale: 5, Required: true},
    kit.Question{ID: "note", Title: "建议", Kind: kit.QuestionLongText},
).OnSubmit(func(a map[string]kit.Answer) { save(a) })
```

- 题型：`QuestionSingle`（单选）、`QuestionMultiple`（多选）、`QuestionText`（单行文本）、`QuestionLongText`（多行文本）、`QuestionRating`（星级评分，默认 5 星）。
- 顶部是进度条，显示"第 3 / 10 题"。点"下一题"时，必填题没有作答会显示错误，不能翻页；作答后错误自动消失。
- 最后一页的"提交"会校验全部启用题；未作答、未主动跳过或有校验错误时跳到首个错误题，通过后调用 `OnSubmit`。
- 答案用 `Answer` 表示，按题目类型填写 `Text`、`Choices`（按选项顺序）、`Rating`，可另带 `Freeform` 或 `Skipped`；`Answer.Empty()` 判断是否作答。
- `Value()` 返回目前所有的答案；`SetValue` 回填答案，比如恢复草稿；`Page()` / `SetPage(i)` 读取或切换当前页。
- "上一题""下一题""提交"、进度文字、必填提示都来自 locale。

Agent：容器角色 `form`，名字是进度文字；当前题目是以题目标题命名的 `group`，控件以题目标题为名字。

验证：`go run ./examples/components -section questionnaire`，加 `-theme dark` 检查深色。

`SetDisabled(true)` 禁用当前答案控件、前后翻页和提交，不触发校验或 `OnSubmit`；程序仍可调用 `SetPage`、`SetValue`。重新启用后保留当前页及答案。

构造时复制题目及选项，之后修改原始切片不会改变问卷。题目 ID 必须非空且唯一；不支持的题型会在构造时 panic。`Value` 与 `OnSubmit` 提供独立的答案快照，修改返回 map 或多选答案切片不会影响内部状态。

## 条件题、跳过与完成状态

`Question.Disabled` 设置初始禁用条件，`SetQuestionDisabled(id, on)` 动态更新，未知 ID 返回 false。禁用题退出导航、进度、校验和提交，但保留草稿，`Value` 仍可读到。禁用当前题时先找后面的可用题，再找前面的题；全部禁用时 Page 为 -1，组件隐藏。应用可在 `OnAnswerChange(func(id string, answer kit.Answer))` 中根据前题答案更新条件。

可选题现在需要作答或显式跳过，不能把“还没答”隐式当作完成。`Skip()` / “跳过”按钮清空当前可选题答案，保存 `Answer.Skipped=true` 并前进；最后一题跳过后尝试提交。必填题不能跳过。回到跳过的题重新作答会清除跳过状态。`SetValue` 支持恢复可选题的 Skipped 状态。

`Progress()` 返回当前启用题集合的快照：Current 从 1 开始，Total、Answered、Skipped、Unanswered 分别是总题数及三种作答状态的数量，Completed 表示已经完成一次有效提交。`OnComplete` 只在进入完成状态时执行，首次有效提交中先于 OnSubmit；每次有效提交仍调用 OnSubmit。修改答案、条件或外部错误会使完成状态失效。提交快照只包含启用题；Value 仍包含所有题。

## 自由输入和校验

单选、多选题可设置 `FreeformLabel`，显示有名称的自由输入框。`Answer.Freeform` 保存有效自由文本；空白文本不算作答。单选题在选项和自由输入之间切换，选择固定选项时保留自由输入草稿，但不将其作为有效答案提交；重新输入时自由输入生效。多选题允许 Choices 与 Freeform 同时作答。

`Question.Validate func(Answer, map[string]Answer) string` 做同步自定义校验，收到当前题答案和启用题的独立答案快照，返回空串表示通过。下一题校验当前题，提交校验所有启用题并定位第一道错误题；必填/未作答检查在自定义校验之前，主动跳过不运行自定义校验。回调应只检查快照，不修改问卷。

`SetExternalError(id, message)` 保存服务器或应用错误，空串清除；`ExternalError(id)` 查询。外部错误优先显示并阻止通过，输入和 Reset 都不会自动清除。应用负责在接受新答案后清除它，不自动发起网络校验。

`Reset()` 清空答案、跳过、内部错误和完成状态，回到第一道启用题；保留外部错误与当前禁用条件。没有单独的 schema 默认答案基线，应用可在 Reset 后用 SetValue 恢复业务默认值。

## 选项快捷键

`Shortcuts(kit.QuestionnaireShortcutsLetters)` 开启 A–Z，`QuestionnaireShortcutsNumbers` 开启 1–9，默认 Off。快捷键按当前题选项顺序映射，单选选中，多选切换；只在问卷获得焦点且焦点不在文本输入中时处理，不处理带修饰键的按键。按住按键不重复切换，释放后可再次触发。导航仍使用可聚焦的前后/跳过/提交按钮，不提供上游完整的方向键和 Cmd/Ctrl+Enter 导航绑定。选项文字应由应用提示对应快捷键。

自动测试覆盖条件导航、禁用题排除、跳过、重新提交、完成事件、外部错误/Reset、自定义校验、自由输入保留及选项快捷键；本批未做真机视觉/输入法验收。
