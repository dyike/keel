package main

import (
	"fmt"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("questionnaire", "inputs", func() core.Widget {
		result := ""
		q := kit.Questionnaire(
			kit.Question{ID: "role", Title: "你在团队里的角色？", Kind: kit.QuestionSingle, Options: []string{"开发", "设计", "产品", "其他 Other"}, Required: true},
			kit.Question{ID: "tools", Title: "平时用哪些工具？", Description: "可以多选", Kind: kit.QuestionMultiple, Options: []string{"Go", "Figma", "Git", "Keel"}},
			kit.Question{ID: "score", Title: "整体满意度", Kind: kit.QuestionRating, Required: true},
			kit.Question{ID: "name", Title: "怎么称呼你？", Kind: kit.QuestionText},
			kit.Question{ID: "note", Title: "还有什么建议？", Kind: kit.QuestionLongText},
		)
		q.OnSubmit(func(a map[string]kit.Answer) {
			result = fmt.Sprintf("已提交：%s · %s · %d 星", a["role"].Text, strings.Join(a["tools"].Choices, "/"), a["score"].Rating)
		})
		disabled := false
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(480)).MaxW(el.Full).Child(kit.Button("启用 / 禁用问卷", func() { disabled = !disabled; q.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx), q.Render(cx), el.Text(result).TextColor(theme.Muted))
		}))
	})
}
