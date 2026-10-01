# Accordion

可展开的分节面板。默认同时只展开一节，`Multiple()` 后可以展开多节。

```go
faq := kit.Accordion().Add("如何退款？", answer1).Add("多久到账？", answer2).Multiple()
```

- 标题可以获得焦点：↑ ↓ Home End 在标题之间移动，会跳过禁用的节；回车或空格展开、收起。
- `Value()` 返回已展开节的序号；`SetValue(i...)` 不触发回调，单节模式只保留第一个；`SetItemDisabled(i, bool)`。

Agent：容器角色 `group`；每个标题是 `disclosure`，`value` 为 expanded / collapsed；展开的内容单独列出。

验证：`go run ./examples/components -section accordion`，加 `-theme dark` 检查深色。
