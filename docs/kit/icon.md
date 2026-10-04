# Icon

`kit.Icon(kit.IconInfo).Size(24).Color(theme.Info)` 创建矢量图标。省略 Color 时每次 Render 读取 theme.Text。Size 接受 float32 dp。零值 `IconNone` 不画任何东西、不占空间，可选图标的字段不设就是没有图标。内置图标来自 Material Design：Check、Done、Close、Plus、Minus、Search、Copy、ChevronLeft、ChevronRight、ChevronDown、Info、Warning、Error、User、Inbox、Star、StarOutline、Calendar、Clock、Settings、Bell、Lock、Folder、File、Archive、Receipt、Home、Trash、Edit。按含义选图标，不要借用形状相近的图标（例如用 Info 表示设置）。解析结果会缓存，每帧调用 `Icon` 不会重复解码。

图标遵守父级尺寸约束，本身不提供语义和键盘操作；由包含它的控件提供名称。运行 `go run ./examples/components -section icon -theme dark` 检查深色，省略 theme 检查浅色。

`Rotate(degrees)` 围绕布局中心顺时针旋转，负数为逆时针，角度按 360° 归一化；非有限值忽略。旋转不改变布局占位，绘制超出原框时由祖先裁剪。支持内置 Icon 和 VectorIcon；`Rotate(0)` 恢复。Size 同样忽略非有限值。组件库加入 0/45/90/180/270° 示例；自动测试验证双倍率旋转、居中及占位。SVG 路径/字节入口见下。

## SVG 文件和字节

`SVGIcon(data)` 与 `SVGIconFile(path)` 返回 `(*IconView, error)`，接收最多 1MiB 的 SVG。文件读取与解析同步执行，应在初始化时创建并保留返回值，避免每帧重复解析。文件入口只读本地文件，不监听变化或下载 URL。

支持形状、路径、组、变换和渐变等 [oksvg 支持的 SVG 子集](https://github.com/srwiley/oksvg)。采用严格错误模式；解析失败、无有效 viewBox/尺寸或识别到不支持元素时返回错误。它不是浏览器 SVG 引擎，不支持脚本、动画、字体文字、滤镜等完整网页能力，适用于应用自带的图标资源。

默认按图形 alpha 轮廓使用 Color（未设置时为 theme.Text）染色，与内置图标一致。`OriginalColors(true)` 保留源色，此时 currentColor 解析为黑色。图标保持长宽比并居中，复用 Size/Rotate。按物理像素栅格化，每个实例只缓存最近尺寸/颜色的图像；栅格最长边限制为 2048 像素，更大显示尺寸会放大缓存图。

自动测试覆盖本地文件、输入上限、无效尺寸和不支持元素，以及双倍率非零 viewBox、渐变、组变换、旋转、原色/主题染色和透明度像素。缓存图转换到 Gio 使用的颜色空间，避免半透明颜色变暗；最长边限制与主题/尺寸切换也已验证。原生窗口视觉未验收。
