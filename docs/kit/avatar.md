# Avatar

固定尺寸的圆形头像，默认 40dp；图片缺失时显示姓名首字母。

```go
person := kit.Avatar("Ada Lovelace").Size(40).Image(decodedImage)
return person.Render(cx)
```

构造函数返回 `*AvatarView`。`Size(dp)` 限制在 16–256dp，非正数不改变尺寸；`SetName` 更新姓名。`Image(image.Image)` 使用已解码图片，按中心裁剪填满圆形；nil 或空图片恢复首字母，不改变布局。图片由应用在后台加载，经 `core.Update` 赋值，本组件不发网络请求。

单词名显示首个 Unicode 字符，多词名显示首尾词的首字符，英文字母转大写；空姓名显示 `?`。这是字符回退，不负责复杂 emoji 字素分组。Agent 复用 `image` 角色，名称为完整姓名（空名用 Avatar），value 为 loaded 或 initials。无点击和键盘交互。

验证：`go run ./examples/components -section avatar -theme dark`，支持 light。测试检查中英文、数字、空名和图片/回退切换的尺寸稳定性。
