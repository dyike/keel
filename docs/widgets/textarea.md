# 多行输入自动高度

```go
notes := widget.TextArea("备注").AutoHeight(1,4)
```
根据实际排版内容（包括软换行）增长或缩小，超过最大可见行数后内部滚动，不截断文本。最小行数至少 1，最大行数至少等于最小行数；单行 Input 忽略该设置。保留占位文字、只读、禁用与静默 SetValue。

验证入口：`go run ./examples/components -section textarea`。
