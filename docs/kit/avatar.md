# Avatar

`kit.Avatar("张三").Size(40).Status(kit.Online)` 显示圆形头像，Size 接受 float32 dp，推荐小、中、大尺寸为 24、40、56。Image 接受已解码的 image.Image，nil 恢复姓名回退。后台加载完成后用 core.Update 更新。

中文取第一个字，英文取前两个单词首字母，空姓名显示问号。背景根据姓名哈希从当前主题选取。Status 支持 Online、Busy、Offline，空值隐藏；状态点位于头像内部，不改变布局。

Agent 角色 avatar，名字为人名，value 为状态，无状态为空。纯展示，不处理键盘。验证：`go run ./examples/components -section avatar -theme dark`，省略 theme 查看浅色。
