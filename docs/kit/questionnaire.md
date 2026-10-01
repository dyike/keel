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
- 最后一页的"提交"会检查所有题目：有必填题没答时，跳到第一道没答的题；全部答完才调用 `OnSubmit`。
- 答案用 `Answer` 表示，按题目类型填写 `Text`、`Choices`（按选项顺序）或 `Rating` 中的一个；`Answer.Empty()` 判断是否作答。
- `Value()` 返回目前所有的答案；`SetValue` 回填答案，比如恢复草稿；`Page()` / `SetPage(i)` 读取或切换当前页。
- "上一题""下一题""提交"、进度文字、必填提示都来自 locale。

Agent：容器角色 `form`，名字是进度文字；当前题目是以题目标题命名的 `group`，控件以题目标题为名字。

验证：`go run ./examples/components -section questionnaire`，加 `-theme dark` 检查深色。
