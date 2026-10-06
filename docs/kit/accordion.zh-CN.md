# Accordion

[English](accordion.md) | 简体中文

可展开的分节面板。默认同时只展开一节，`Multiple()` 后可以展开多节。

```go
faq := kit.Accordion().Add("如何退款？", answer1).Add("多久到账？", answer2).Multiple()
```

- 标题可以获得焦点：↑ ↓ Home End 在标题之间移动，会跳过禁用的节；回车或空格展开、收起。
- `Value()` 返回已展开节的序号；`SetValue(i...)` 不触发回调，单节模式只保留第一个；`SetItemDisabled(i, bool)`。

Agent：容器角色 `group`；每个标题是 `disclosure`，`value` 为 expanded / collapsed；展开的内容单独列出。

验证：`go run ./examples/components -section accordion`，加 `-theme dark` 检查深色。

`Heading(i, view)` 自定义第 i 项的展示标题，原 title 仍作为无障碍名称。`Trigger(i)` 和 `Content(i)` 可分开放进自定义布局，每项各渲染一次；不要同时再渲染完整 Accordion。方向键会跳过未渲染、禁用或位于禁用容器中的标题。

`SetDisabled` 禁用整组，`SetItemDisabled` 同时禁用该项的标题与已展开内容，但保留展开状态。展开/收起有 180ms 动画，连续切换从当前高度反转；减少动画立即切换。收起时输入状态保留，焦点若在正文中会回到标题，收起过程不能继续操作正文。

`Bordered(false)` 隐藏外框与分节分隔线，保留背景、圆角和展开状态；默认有边框。`Size` 支持 `AccordionSizeXSmall`、`AccordionSizeSmall`、`AccordionSizeMedium`（默认）、`AccordionSizeLarge`，统一调整标题、箭头、间距与正文继承字号。默认档保留原有字号继承；自定义标题或正文显式设置的字号优先。
