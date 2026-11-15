package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"strings"
)

func init() {
	registerSection("diff", "data", func() core.Widget {
		enableHighlight()
		context := strings.Repeat("// unchanged context\n", 16)
		initial := kit.DiffValue{Before: "package main\n\nfunc greeting() string {\n\treturn \"Hello\"\n}\n\n" + context + "// old note\n", After: "package main\n\nfunc greeting() string {\n\treturn \"Hello, Keel\"\n}\n\n" + context + "// reviewed note\n"}
		patch := kit.Diff(initial.Before, initial.After).Patch("main.go", "main.go")
		d := kit.Diff(initial.Before, initial.After).Language("go").Name(demoText("Review changes", "审阅更改"))
		lineStatus := demoText("Click a line number to locate its source.", "点击行号查看对应源位置。")
		d.OnLineNumberPress(func(line kit.DiffLine) {
			side := demoText("After", "更改后")
			if line.Side == kit.DiffSideBefore {
				side = demoText("Before", "更改前")
			}
			lineStatus = fmt.Sprintf("%s · %d", side, line.Line+1)
		})
		d.HunkActions(func(h kit.DiffHunk) []kit.DiffAction {
			return []kit.DiffAction{
				{Label: demoText("Inspect region", "查看区域"), OnSelect: func(current kit.DiffHunk) {
					lineStatus = fmt.Sprintf("%s %d · -%d,%d +%d,%d", demoText("Region", "区域"), current.Index+1, current.BeforeStart+1, current.BeforeEnd-current.BeforeStart, current.AfterStart+1, current.AfterEnd-current.AfterStart)
				}},
				{Label: demoText("Restore region", "还原区域"), Disabled: h.PatchOnly, OnSelect: func(current kit.DiffHunk) { d.RevertAt(current.Index) }},
			}
		})
		modes := kit.Select(demoText("Presentation", "展示方式"), "Auto", "Unified", "Split").OnChange(func(s string) {
			switch s {
			case "Unified":
				d.Mode(kit.DiffModeUnified)
			case "Split":
				d.Mode(kit.DiffModeSplit)
			default:
				d.Mode(kit.DiffModeAuto)
			}
		})
		modes.SetValue("Auto")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().ScrollY().P(16).Gap(12).Child(modes.Render(cx), el.Text(lineStatus), kit.Button(demoText("Inspect patch", "查看补丁"), func() { _ = d.SetPatch(patch) }).Render(cx), kit.Button(demoText("Reset comparison", "重置比较"), func() { d.SetValue(initial) }).Render(cx), d.Render(cx))
		}))
	})
}
