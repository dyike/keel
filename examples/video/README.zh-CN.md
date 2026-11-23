# Keel Motion

[English](README.md) | 简体中文

Keel Motion 用 Go 定义动态图形。选好模板后，可以播放、拖动时间线或逐帧查看。点 **Export MP4** 导出 1280 × 720、30 fps 的视频。Orbit、Signal、Bloom 是 4 秒的无声动效。Keel Film 是 27 秒的中文宣传片，工作台提供画面预览，导出的 MP4 包含普通话旁白和原创配乐。

## 运行网站

在仓库根目录执行以下命令。第一条将 Keel 界面编译成 WebAssembly，第二条启动本机预览和渲染服务。

```sh
go run ./cmd/gogio -target js -tags osusergo -name "Keel Motion" -o work/video/web ./examples/video
go run ./examples/video -web work/video/web
```

浏览器打开 [Keel Motion](http://127.0.0.1:4173)。播放和拖动时间线在浏览器内完成；导出时，本机服务用离屏 GPU 逐帧绘制，再调用 FFmpeg 编码并将文件发给浏览器。导出需要机器上可用的 GPU 和带 libx264 的 `ffmpeg`，后者须在 `PATH` 中。只托管静态文件也可以预览。

## 桌面与命令行

```sh
go run ./examples/video
go run ./examples/video -scene Bloom -palette Tangerine -render bloom.mp4
go run ./examples/video -scene "Keel Film" -render keel-film.mp4
```

桌面工作台将视频保存到当前工作目录，文件名带有模板名和随机后缀。命令行直接输出到指定的新文件。配色可选 Cobalt、Tangerine、Iris，省略时使用模板的默认配色。

## 编写动效

在 `scene.go` 中编写绘图函数并加入 `templates`。每个函数接收 Gio 绘图指令列表、配色和当前相位。一个周期的相位是 `2π × 帧序号 / 总帧数`，首帧序号为 0。画面以中心为原点，高度为 360 个绘图单位。预览和导出按目标像素大小缩放同一张画面。暂停后把时间线拖回同一帧，就能检查同一个姿态。
