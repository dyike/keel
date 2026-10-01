# Attachment

附件卡片：文件名、大小、上传进度或错误，可以打开、移除。

```go
a := kit.Attachment("报价单.pdf", size).OnRemove(remove).OnOpen(open)
a.SetProgress(0.6)          // 上传中；负数表示已完成
a.SetError("超过 10 MB 上限")
```

- 大小用 `kit.FileSize` 格式化为 B / KB / MB / GB。上传中显示进度条和"上传中 60%"，出错时用危险色显示原因。
- "上传中"文字来自 locale。

Agent：角色 `attachment`，名字是文件名；`value` 为空、"上传中 60%"或 `error`；移除按钮名为"移除 文件名"。

验证：`go run ./examples/components -section attachment`，加 `-theme dark` 检查深色。
