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
			kit.Question{ID: "role", Title: "你在团队里的角色？", Kind: kit.QuestionSingle, Options: []string{"开发", "设计", "产品"}, FreeformLabel: "其他角色", Required: true},
			kit.Question{ID: "tools", Title: "平时用哪些工具？", Description: "可以多选；Figma 暂不可选，数字快捷键按可选项编号", Kind: kit.QuestionMultiple, Options: []string{"Go", "Figma", "Git", "Keel"}, FreeformLabel: "其他工具"},
			kit.Question{ID: "score", Title: "整体满意度", Kind: kit.QuestionRating, Required: true},
			kit.Question{ID: "name", Title: "怎么称呼你？", Kind: kit.QuestionText},
			kit.Question{ID: "note", Title: "还有什么建议？", Kind: kit.QuestionLongText},
		)
		q.Shortcuts(kit.QuestionnaireShortcutsNumbers)
		q.SetChoiceDisabled("tools", "Figma", true)
		q.OnSubmit(func(a map[string]kit.Answer) {
			result = fmt.Sprintf("已提交：%s · %s · %d 星", a["role"].Text, strings.Join(a["tools"].Choices, "/"), a["score"].Rating)
		})
		density := kit.QuestionnaireSizeMedium
		disabled := false
		custom := true
		q.Layout(func(cx *el.Context, state kit.QuestionnaireContext, parts kit.QuestionnaireParts) el.Element {
			if !custom {
				return nil
			}
			return el.Div().Gap(12).Items(el.Stretch).Child(
				parts.Progress,
				el.Div().Border(1, theme.Border).Rounded(theme.RadiusLg).P(16).Gap(12).Items(el.Stretch).Child(parts.Title, parts.Description, parts.Answer, parts.Freeform, parts.Error),
				el.Div().Row().Gap(8).Child(parts.Previous, el.Div().Grow(), parts.Skip, parts.Forward),
			)
		})
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(16).W(el.Dp(480)).MaxW(el.Full).Child(kit.Button("启用 / 禁用问卷", func() { disabled = !disabled; q.SetDisabled(disabled) }).Variant(kit.ButtonSecondary).Render(cx), kit.Button("切换默认 / 自定义布局", func() { custom = !custom }).Render(cx), kit.Button("切换问卷尺寸", func() { density = (density + 1) % 4; q.Size(density) }).Render(cx), q.Render(cx), el.Text(result).TextColor(theme.Muted))
		}))
	})
}
