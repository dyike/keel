package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"time"
)

func init() {
	registerSection("date_picker", "inputs", func() core.Widget {
		due := kit.DatePicker("交货日期").Placeholder("选择日期").Format("2006-01-02").Clearable(true).Size(28)
		trip := kit.DatePicker("出差日期 Range").Range().Months(2).Placeholder("开始 – 结束").Clearable(true).Size(48).Appearance(false)
		appointment := kit.DatePicker("预约时间").TimeSeconds().TimeHour12(false).DefaultTime(9 * time.Hour).Clearable(true)
		today := time.Now()
		appointment.Presets(kit.DatePickerPreset{ID: "tomorrow-morning", Label: "明早 09:30", Start: time.Date(today.Year(), today.Month(), today.Day()+1, 9, 30, 0, 0, today.Location()), IncludeTime: true})
		due.Presets(kit.DatePickerPreset{ID: "today", Label: "今天", Start: today}, kit.DatePickerPreset{ID: "tomorrow", Label: "明天", Start: today.AddDate(0, 0, 1)})
		trip.Presets(kit.DatePickerPreset{ID: "week", Label: "最近七天", Start: today.AddDate(0, 0, -6), End: today})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Items(el.Start).Child(el.Div().Gap(14).W(el.Dp(300)).MaxW(el.Full).Child(due.Render(cx), trip.Render(cx), appointment.Render(cx)))
		}))
	})
}
