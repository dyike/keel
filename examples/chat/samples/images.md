## 图片

下面是本地 PNG，显示红、绿、蓝三个色块。运行示例时从仓库根目录启动，图片路径无需联网。

![红绿蓝色块](examples/chat/assets/colors.png)

图片前后的文字应正常排版。缩窄窗口时，图片应保持比例并适应可用宽度。

点击下面带链接的图片应打开 Go 文档：

[![可点击的色块图片](examples/chat/assets/colors.png)](https://go.dev/doc/)

下面图片没有替代文字，但图片本身仍应显示：

![](examples/chat/assets/colors.png)

下面的图片故意使用不存在的路径，检查加载失败时的占位和替代文字，后面的正文应继续显示：

![图片加载失败时的替代文字](examples/chat/assets/missing.png)

图片样例结束。
