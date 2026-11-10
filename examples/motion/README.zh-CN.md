# Motion 示例

直接运行本次要求的指针原点与订单动效：

```sh
go run ./examples/motion -section effects -lang zh-CN
```

在蓝色卡片上移动指针：光晕会跟随指针，进入或按下时，波纹从当前指针位置扩散。
点击 **提交订单**，由 Gio 绘制的确定性彩纸会从按钮中心喷出。开启 **减少动态效果**
后，两种效果都不播放。


动效演示统一放在这里。运行 `go run ./examples/motion -lang zh-CN` 查看完整实验，或选择分区：

- `-section presets`：Smooth / Snappy / Bouncy 弹簧对照，支持运动中反向。
- `-section pages`：切页淡入与输入保留。
- `-section effects`：指针光晕、波纹和彩纸。

`examples/app-shell` 展示应用布局。
