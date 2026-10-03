# StatusMarker

`kit.StatusMarker("已同步")` 用于消息状态、时间线边界和系统提示。默认铺满父容器宽度，内容靠左，文字使用主题 Muted。现有 `kit.Marker(shape)` 继续绘制图表几何标记。

```go
kit.StatusMarker("今天").Variant(kit.StatusMarkerSeparator)
kit.StatusMarker("3 条未读").Variant(kit.StatusMarkerBorder).
    Content(kit.Button("查看", openMessages))
kit.StatusMarker("正在生成…").Loading(true).
    LoadingStyle(kit.StatusMarkerLoadingStyleShimmer).ID("generation").Role("status")
```

Variant 支持 Plain、Separator、Border；分隔线默认居中，Alignment(el.Start/Center/End) 显式覆盖，ResetAlignment 恢复默认。Separator 靠左时只留右侧线，靠右时只留左侧线；Border 在整行底部画线。

Icon 接收任意 View，默认 16dp 槽；Content 在文字后添加富内容，SetText("") 可只用富内容。Children 替换直接添加到行尾的子项，输入切片会复制。按钮等子控件保留自己的点击、焦点和禁用继承，状态行自身不成为可点击控件，也不维护未读计数或通知生命周期。

Loading 默认添加 spinner；已有 Icon 时保留图标。Shimmer 模式对文字扫光，纯富内容按两秒周期做透明度变化；图标和线保持静止。ShimmerStyle 回调配置持久的 ShimmerText（Duration、Spread、Highlight、Reverse、Once 等），文字与 Enabled 由状态行管理。切换 Loading/LoadingStyle 重启文字扫光；减少动画时静态显示。富内容透明度固定在 0.7–1，文字扫光配置不改变它。

PartStyle 分别配置 Root、Row、Icon、Content、Separator；回调每帧收到新元素，ID 在样式之后恢复，避免丢失子控件状态。Separator 样式也作用于 Border 底线。默认不添加状态角色，需要时用 ID/Role；Agent 可以读取文字，系统读屏仍受 Keel/Gio 平台支持限制。

布局使用 Keel 的行布局和主题刻度，不复制上游尺寸。Alignment 同时设置内容组位置和文字行内对齐；子元素可以用 TextAlign 独立覆盖。

验证：`go run ./examples/components -section status_marker`。
