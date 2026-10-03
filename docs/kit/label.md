# Label

`kit.Label(text)` 显示可换行标签。`Secondary(text)` 在同一文本流后追加次级色文案，空字符串移除；`SetText` 更新主文案。

```go
kit.Label("公司名称").Secondary("（可选）").Highlights("公司")
kit.Label("账户余额").Masked(true)
kit.Label("Hello World").HighlightPrefix("Hello").Style(func(t *el.TextEl) {
    t.TextSize(20).Bold().LineHeight(1.5)
})
```

`Highlights` 区分大小写，标记全部不重叠的精确匹配；`HighlightPrefix` 只匹配开头，两者后调用生效，空字符串清除。`HighlightColor` 覆盖主题 PrimaryText。次级文案使用当前主题 Muted，不参与搜索。

`Masked(true)` 把主文案每个 Unicode rune 替换成一个圆点，同时关闭主文案高亮；渲染和 Agent 语义均只包含圆点。次级文案仍可见。组合字符和 emoji 序列可能对应多个圆点，不按字形计数。

`Style` 接收每帧新建的 TextEl，可配置字体、字重、行高、对齐、宽度、MaxLines 和 FocusOnPress 字段聚焦关联。不要保留该元素引用，或通过自定义 Name 重新暴露遮罩原文。nil 恢复默认样式。

底层 `el.Text(...).Ranges(...)` 用半开 rune 区间着色，不拆开整段排版。匹配涉及连字/组合字符时，整个字形簇着色；彩色位图字形保持原色，截断省略号使用基础颜色。重叠区间后者优先，Shimmer 优先于 Ranges。
