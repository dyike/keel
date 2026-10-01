// Package locale holds the text Keel itself shows or reports to agents: button
// captions such as 确定 and 复制, accessible names such as 关闭, placeholders
// and counts. Apply switches all of it at runtime and redraws every window.
//
// Only framework text lives here. An app translates its own labels however it
// likes; Lang tells it which language is active, and Revision changes on every
// Apply so render caches can include it.
package locale

import (
	"strconv"
	"time"

	"github.com/dyike/keel/ui/internal/loop"
)

// Strings is every piece of framework text. Start from Chinese or English
// when customizing: Apply replaces all of it.
type Strings struct {
	Lang string // BCP 47 tag of this text, e.g. zh-CN, en

	OK, Cancel, Close, Remove, Toggle, Clear string
	Copy, Copied                             string
	Loading, Menu, MoreOptions               string
	SelectHint, NoData                       string
	Image, ImageLoading, ImageFailed         string
	PlainText, WrapLines, NoWrapLines        string
	Search, NoMatches                        string
	Increase, Decrease                       string
	PrevMonth, NextMonth                     string
	PrevPage, NextPage                       string
	Commands, SearchCommands                 string
	Latest, Uploading                        string
	More, Resize                             string
	CollapseSidebar, ExpandSidebar           string
	PrevSlide, NextSlide                     string
	DockLeft, DockRight, DockBottom          string
	SearchSettings                           string
	// Total formats an item count for a pager, e.g. "共 36 条".
	Total func(n int) string

	// Weekdays are short day names starting with Sunday; FirstWeekday is the
	// column a calendar starts with (time.Monday for Chinese).
	Weekdays     [7]string
	FirstWeekday time.Weekday
	// Month titles a calendar page, e.g. "2026年10月" or "October 2026".
	Month func(year int, month time.Month) string
	// Date formats a day for a date field, e.g. "2026-10-01".
	Date func(t time.Time) string

	// Rows formats a row count, e.g. "36 行" or "36 rows".
	Rows func(n int) string
}

// Name joins an action and its target into an accessible name, such as
// "关闭 保存成功" or "Close Saved".
func (s Strings) Name(action, target string) string {
	if target == "" {
		return action
	}
	return action + " " + target
}

// Chinese returns the default text, simplified Chinese.
func Chinese() Strings {
	return Strings{
		Lang: "zh-CN",
		OK:   "确定", Cancel: "取消", Close: "关闭", Remove: "移除", Toggle: "切换", Clear: "清空",
		Copy: "复制", Copied: "已复制",
		Loading: "加载中", Menu: "菜单", MoreOptions: "更多选项",
		SelectHint: "请选择", NoData: "暂无数据",
		Image: "图片", ImageLoading: "图片加载中", ImageFailed: "图片加载失败",
		PlainText: "纯文本", WrapLines: "自动换行", NoWrapLines: "取消自动换行",
		Search: "搜索", NoMatches: "无匹配项", Increase: "增加", Decrease: "减少",
		PrevMonth: "上个月", NextMonth: "下个月", PrevPage: "上一页", NextPage: "下一页",
		Commands: "命令面板", SearchCommands: "搜索命令…",
		Latest: "回到最新", Uploading: "上传中",
		More: "更多", Resize: "调整大小", CollapseSidebar: "收起侧栏", ExpandSidebar: "展开侧栏",
		PrevSlide: "上一张", NextSlide: "下一张",
		DockLeft: "停靠到左侧", DockRight: "停靠到右侧", DockBottom: "停靠到底部", SearchSettings: "搜索设置",
		Total:        func(n int) string { return "共 " + strconv.Itoa(n) + " 条" },
		Weekdays:     [7]string{"日", "一", "二", "三", "四", "五", "六"},
		FirstWeekday: time.Monday,
		Month:        func(y int, m time.Month) string { return strconv.Itoa(y) + "年" + strconv.Itoa(int(m)) + "月" },
		Date:         func(t time.Time) string { return t.Format("2006-01-02") },
		Rows:         func(n int) string { return strconv.Itoa(n) + " 行" },
	}
}

// English returns English text.
func English() Strings {
	return Strings{
		Lang: "en",
		OK:   "OK", Cancel: "Cancel", Close: "Close", Remove: "Remove", Toggle: "Toggle", Clear: "Clear",
		Copy: "Copy", Copied: "Copied",
		Loading: "Loading", Menu: "Menu", MoreOptions: "More options",
		SelectHint: "Select…", NoData: "No data",
		Image: "Image", ImageLoading: "Loading image", ImageFailed: "Image failed to load",
		PlainText: "Plain text", WrapLines: "Wrap lines", NoWrapLines: "Don't wrap lines",
		Search: "Search", NoMatches: "No matches", Increase: "Increase", Decrease: "Decrease",
		PrevMonth: "Previous month", NextMonth: "Next month", PrevPage: "Previous page", NextPage: "Next page",
		Commands: "Command palette", SearchCommands: "Type a command…",
		Latest: "Jump to latest", Uploading: "Uploading",
		More: "More", Resize: "Resize", CollapseSidebar: "Collapse sidebar", ExpandSidebar: "Expand sidebar",
		PrevSlide: "Previous slide", NextSlide: "Next slide",
		DockLeft: "Dock left", DockRight: "Dock right", DockBottom: "Dock bottom", SearchSettings: "Search settings",
		Total: func(n int) string {
			if n == 1 {
				return "1 item"
			}
			return strconv.Itoa(n) + " items"
		},
		Weekdays:     [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"},
		FirstWeekday: time.Sunday,
		Month:        func(y int, m time.Month) string { return m.String() + " " + strconv.Itoa(y) },
		Date:         func(t time.Time) string { return t.Format("Jan 2, 2006") },
		Rows: func(n int) string {
			if n == 1 {
				return "1 row"
			}
			return strconv.Itoa(n) + " rows"
		},
	}
}

var (
	current  = Chinese()
	revision uint64
)

// Current returns the active text. Read it under the UI lock, in Render or a
// callback, or before opening the first window.
func Current() Strings { return current }

// Revision changes whenever Apply replaces the text.
func Revision() uint64 { return revision }

// Apply replaces the framework text and redraws every window. Call it before
// opening windows or under the UI lock; from other goroutines use
// core.Update(func() { locale.Apply(s) }). A nil Rows keeps the current one.
func Apply(s Strings) {
	if s.Rows == nil {
		s.Rows = current.Rows
	}
	if s.Month == nil {
		s.Month = current.Month
	}
	if s.Total == nil {
		s.Total = current.Total
	}
	if s.Date == nil {
		s.Date = current.Date
	}
	current = s
	revision++
	loop.InvalidateAll()
}
