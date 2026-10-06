# Questionnaire

[English](questionnaire.md) | 简体中文

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

`Question.DefaultAnswer` 定义构造时的默认答案，Choices 会复制。`Reset()` 恢复这份默认答案、清除内部错误和完成状态，回到第一道启用题；保留外部错误与当前禁用条件。后续 SetValue 只改变当前答案，不移动 Reset 基线。

## 选项快捷键

`Shortcuts(kit.QuestionnaireShortcutsLetters)` 开启 A–Z，`QuestionnaireShortcutsNumbers` 开启 1–9，默认 Off。快捷键按当前题未禁用选项的顺序映射，单选选中，多选切换；只在问卷获得焦点且焦点不在文本输入中时处理，不处理带修饰键的按键。按住按键不重复切换，释放后可再次触发。默认键盘导航支持左右键前后翻题、上下键在多选控件/自由输入之间移动焦点、Enter 确认已作答或已跳过的当前题，以及 Cmd/Ctrl+Enter 从文本输入确认。文本输入保留普通方向键和 Enter；单选组、评分控件保留自身方向键行为；操作按钮保留 Enter 激活。空答案不会因 Enter 隐式提交。`KeyboardNavigation(false)` 关闭这些导航绑定，独立于选项快捷键配置。选项文字应由应用提示对应快捷键。

自动测试覆盖条件导航、禁用题排除、跳过、重新提交、完成事件、外部错误/Reset、自定义校验、自由输入保留及选项快捷键；尚未完成真机视觉/输入法验收。


`SetChoiceDisabled(questionID, option, on)` 动态禁用固定选项，未知题/选项返回 false。禁用项保留内部选择草稿，但退出有效答案、校验快照和快捷键编号；重新启用恢复有效答案，条件变化会清除完成状态。全题必填规则仍生效：若禁用了全部选项且没有可用自由输入，应用应同时禁用该题以避免无法作答。程序 SetValue 可以恢复禁用项的草稿，但 Value/提交只包含可用项。

新增自动测试覆盖默认答案所有权/Reset、禁用项有效答案与快捷键重编号、文本 Cmd/Ctrl+Enter、普通箭头保护、多选焦点移动和关闭导航；未做真机输入法及视觉验收。

## 自定义组合布局

`Layout(func(cx, state, parts) el.Element)` 可重新排列当前题目的组成部分。`QuestionnaireParts` 提供 Progress、Title、Description、Answer、Freeform、Error、Previous、Skip、Forward；不适用的部分为 nil。返回 nil 或 `Layout(nil)` 恢复默认排列。

这些元素使用同一份答案控件与原有校验/跳过/提交命令。外层 form 始终保留禁用继承、键盘导航和焦点管理。每个元素只能放入本帧树中一次，不能跨帧保存；通常应保留 Answer 和导航按钮。省略部件只隐藏对应界面，不会关闭快捷键命令。自定义布局改变控件的树路径时，输入焦点和选区可能重置。

`QuestionnaireContext` 提供当前 Question、Answer、Progress、从零开始的 Page、错误文字及问卷自身 Disabled；题目选项、默认答案与当前答案的切片均为副本。不要在布局回调中修改问卷状态，修改应放在事件回调中。组件库提供默认布局与自定义卡片的切换。自动测试覆盖组合后的必填校验、输入、禁用提交、上下文副本及默认布局恢复；原生输入法未验收。

## 统一尺寸

`Size(QuestionnaireSizeXSmall/Small/Medium/Large)` 调整标题、说明、错误字号、选项标记和文字、评分星形、输入框、部件间距与导航按钮。默认 Medium 沿用原布局；大小切换保留答案和控件实例。自定义 Layout 可从 `QuestionnaireContext.Size` 读取档位，提供的内置部件已应用尺寸，应用新增的容器与控件自行配置。组件库提供循环切换按钮；自动测试覆盖双倍率五种题型、四档导航按钮、答案保留、省略部件及恢复输入。原生视觉未验收。
