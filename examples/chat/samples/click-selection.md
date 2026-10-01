## 双击选词、三击选段

Double-click these words: streaming renderer, hello_world, context.Context, and can't. 标点旁边的词不应把整个段落一起选中。

这段中文包含连续汉字、中文标点，以及 mixed English words。双击不同位置，检查中文按字、英文按词的选择边界。

跨样式单词：render**ing**。双击这个词时应选中完整的 rendering。独立的 **bold**、*italic*、`code_word` 也应能选词。

三击这一段应选中整个段落，包括自动折行后的内容。把窗口缩窄到让这一段折成多行，再分别点击首行和末行，确认相邻段落没有被选中。

这是相邻段落，用来检查三击选段的边界。

```go
hello_world := "double click a word"
second_line := "triple click selection boundary"
```

代码块里三击应只选中当前源码行。双击或三击后按住并拖动，检查选区按整词或整段扩展；拖到消息区上下边缘并停住，内容应持续滚动，松开鼠标后停止。

保留选区后使用滚轮，页面应能滚动且选区不变。

每次双击或三击后都按 Cmd/Ctrl+C，粘贴检查选区内容。
