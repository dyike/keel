# Motion 示例

直接运行本次要求的指针原点与订单动效：

```sh
go run ./examples/motion -section effects -lang zh-CN
```

在蓝色卡片上移动指针：光晕会跟随指针，进入或按下时，波纹从当前指针位置扩散。
点击 **提交订单**，由 Gio 绘制的确定性彩纸会从按钮中心喷出。开启 **减少动态效果**
后，两种效果都不播放。

运行 `go run ./examples/motion` 可以查看完整的九个 Motion 实验。只截取目标实验：

```sh
go run ./examples/motion -section effects -lang zh-CN -screenshot /tmp/keel-motion-effects.zh-CN.png
```

运行 `go run ./examples/motion -section pose -lang zh-CN` 查看只重绘的元素动画。点击“移动 / 反向”或移动中的卡片，点击区域和锚定浮层同步跟随；减少动态效果开启时直接到达目标。

运行 `go run ./examples/motion -section timeline -lang zh-CN` 查看轨道重叠、重播、暂停、倒放和跳转。

时间线页面新增“滚动驱动 / 时间驱动”：滚动下方区域即可直接控制同一组动画轨道。
