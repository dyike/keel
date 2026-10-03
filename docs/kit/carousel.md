# Carousel

轮播：一次显示一张，带上一张、下一张按钮和指示点。

```go
car := kit.Carousel(slide1, slide2, slide3).Height(200).Autoplay(4 * time.Second)
```

- 点击按钮或指示点切换；轮播获得焦点后，横向用 ← →、竖向用 ↑ ↓ 切换，默认首尾循环；Home/End 选择第一张/最后一张。
- `Autoplay(d)` 每隔 d 切到下一张。指针悬停在轮播上时暂停；开启减少动画时不自动播放。
- `SetDisabled(true)` 禁止按钮、指示点和按键切换，同时停止自动播放并移除焦点；父容器禁用、内容离开视口或被模态层遮挡时也暂停自动播放，恢复后重新计时。
- `Value()` / `SetValue(i)`（不触发回调），`OnChange(fn)`。
- "上一张""下一张"的名称来自 locale。

Agent：容器角色 `group`，名字是"当前/总数"（如 2/3）；按钮名为"上一张""下一张"，指示点的名字是序号。

验证：`go run ./examples/components -section carousel`，加 `-theme dark` 检查深色。

`Vertical(true)` 把导航放到内容右侧：上箭头、纵向指示点和下箭头。`Vertical(false)` 恢复内容下方的水平导航。Height 仍控制内容高度（默认 200dp），导航至少保留 72dp；非正或非有限高度忽略。指示点区域可竖向滚动，以容纳较多项目。切换方向保留当前索引及已有键盘焦点，不触发 OnChange；自动播放、悬停暂停、禁用和减少动画规则保持不变。

本组件当前仍是单项切换，不提供拖动/触控板吸附轨道或同屏多项。前后控件和分页项可独立组合，见下文。

`Loop(false)` 关闭循环：到第一项时禁用上一项，到最后一项时禁用下一项，键盘越界保持原选择。`Loop(true)` 恢复默认循环；切换模式不改变索引，也不触发 OnChange。空列表和单项列表的前后按钮始终禁用，空列表的 Agent 计数为 0/0。

`Previous()` / `Next()` 与内置按钮共用导航逻辑，只在索引实际变化时调用 OnChange，并遵守自身禁用状态。`CanPrevious()` / `CanNext()` 返回结合列表数量、循环模式、边界和自身禁用状态的可用性。它们不读取父容器继承的禁用状态；外部控件应放在相同的禁用容器中，或由应用另行禁用。`SetValue` 仍用于无回调的程序设置，即使禁用也可调用。

非循环自动播放到最后一项后停止；程序或用户返回前面的项后，从下一帧重新计时。其他悬停、减少动画和可见性暂停规则不变。

```go
car.Loop(false)
previous := kit.Button("上一项", car.Previous)
previous.SetDisabled(!car.CanPrevious()) // 渲染前按当前状态更新
```

## 独立组合

保留返回的控件实例，在每帧渲染它们：

```go
content := car.Content()
previous := car.PreviousControl(nil)
next := car.NextControl(kit.Button("下一项", nil).Variant(kit.ButtonSecondary).Size(40))
first := car.PaginationItem(0, nil)
second := car.PaginationItem(1, nil)
// 在自己的布局中分别调用 content/previous/next/first/second.Render(cx)。
```

`Content()` 只显示内容区，保留方向键、Home/End、焦点和自动播放。同一个 CarouselView 只挂载一次 Content 或完整 Carousel；独立控件可创建多份。传入 nil 时，前后控件按当前轴显示箭头，分页项显示从 1 开始的数字并突出当前页；分页索引从 0 开始，无效索引禁用。

传入 Button 时继承尺寸、文字、图标、富内容、名称、外观、加载和自身禁用配置。组件接管 ID 和点击回调，不执行原按钮回调，也不修改原实例。每帧读取配置，因此加载/禁用和外观可动态更新；自定义分页外观自行配置，选中状态仍以 toggle/selected 语义公开。点击当前页不重复回调，程序 SetValue 不触发回调。

轮播自身禁用会同步到所有关联控件；父容器禁用只沿实际布局树继承，组合时应把内容和控件放在同一禁用容器内。自动播放暂停检测范围仍是内容区，外部控件的悬停不自动暂停。控制按钮的 Space/Enter 执行操作；方向导航由内容区焦点处理。
