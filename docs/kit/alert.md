# Alert

行内显示状态及说明，不抢占焦点。当前为纯展示版本，关闭按钮在 el 焦点、禁用接口 review 后增加。

```go
notice := kit.Alert("保存成功", "订单 SO-123 已保存").Tone(kit.Success)
// 在 Render 中组合：
return notice.Render(cx)
// 单独作为窗口内容：el.Embed(notice)
```

`Alert(title, description)` 返回 `*AlertView`，默认 Info。链式 `Tone` 支持 Neutral、Info、Success、Warning、Danger；`SetTitle`、`SetDescription` 修改文字。空标题或正文不生成对应文本，窄容器自动换行。颜色在 Render 时读取 theme，支持运行时切换；没有点击、键盘或值回调。

Agent 的角色是 `alert`，名字是标题，value 是语义级别，正文作为子文本保留。

验证：`go run ./examples/components -section alert -theme light`，再用 `-theme dark` 检查深色；单元测试覆盖窄容器、混排、1×/2×、连续帧和更新，窗口测试覆盖角色与正文。
