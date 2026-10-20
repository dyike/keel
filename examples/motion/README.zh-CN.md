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
