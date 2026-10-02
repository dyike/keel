# Carousel

轮播：一次显示一张，带上一张、下一张按钮和指示点。

```go
car := kit.Carousel(slide1, slide2, slide3).Height(200).Autoplay(4 * time.Second)
```

- 点击按钮或指示点切换；轮播获得焦点后，← → 切换，首尾循环。
- `Autoplay(d)` 每隔 d 切到下一张。指针悬停在轮播上时暂停；开启减少动画时不自动播放。
- `SetDisabled(true)` 禁止按钮、指示点和按键切换，同时停止自动播放并移除焦点；父容器禁用、内容离开视口或被模态层遮挡时也暂停自动播放，恢复后重新计时。
- `Value()` / `SetValue(i)`（不触发回调），`OnChange(fn)`。
- "上一张""下一张"的名称来自 locale。

Agent：容器角色 `group`，名字是"当前/总数"（如 2/3）；按钮名为"上一张""下一张"，指示点的名字是序号。

验证：`go run ./examples/components -section carousel`，加 `-theme dark` 检查深色。
