# 扩展指南

新增东西之前，先判断它属于哪个现有模块。绝大多数情况是给现有模块加一个文件；新建模块的标准见[架构 · 什么时候新建模块](architecture.md#什么时候新建模块)。

| 你要加的东西 | 放在哪里 | 例子 |
| --- | --- | --- |
| 界面组件：有状态、能交互 | `ui/widget/` 新文件 | `select.go`、`list.go`、`switch.go` |
| 布局容器：只负责摆放子组件 | `ui/layout/` 新文件 | `grid.go`、`center.go` |
| 颜色、字号等可调参数 | `ui/theme/theme.go` | 被两个以上组件用到的数值 |
| 与窗口本身有关 | `ui/window/` | 窗口位置、置顶 |
| 系统能力，不开窗口也有用 | `native/` 下新模块 | `native/clipboard`、`native/notify`、`native/tray` |
| 只有一个页面用到的布局片段 | 不进库，业务代码里用 `core.Func` | 用到第二次再考虑挪进来 |

## 新增组件

以仓库里的 `widget.Link`（可点击的文字）为例，完整代码在 `ui/widget/link.go`：

```go
package widget

import (
    "image"

    "gioui.org/io/pointer"
    "gioui.org/widget"
    "gioui.org/widget/material"

    "github.com/dyike/keel/ui/core"
    "github.com/dyike/keel/ui/theme"
)

// LinkText is clickable text in the primary color.
type LinkText struct {
    text    string
    onClick func()
    click   widget.Clickable // ① 交互状态存在结构体里，跨帧保留
}

// Link creates clickable text.
func Link(text string, onClick func()) *LinkText { return &LinkText{text: text, onClick: onClick} }

func (l *LinkText) SetText(s string) { l.text = s }

func (l *LinkText) Layout(gtx C) D {
    gtx.Constraints.Min = image.Point{} // ② 不被 Column 拉满
    for l.click.Clicked(gtx) {          // ③ 先处理事件
        core.Call(gtx, l.onClick)       // ④ 用户回调一律走 core.Call
    }
    return l.click.Layout(gtx, func(gtx C) D {
        pointer.CursorPointer.Add(gtx.Ops)
        lb := material.Label(theme.Material, theme.BodySize, l.text)
        lb.Color = theme.Primary        // ⑤ 颜色、字号从 theme 取
        return lb.Layout(gtx)
    })
}
```

`C`、`D` 是 `core.C`、`core.D` 的别名，定义在 `ui/widget/doc.go`。

必须遵守的六条：

1. **状态放在结构体里。** `widget.Clickable`、`widget.Editor`、`widget.Bool` 这些 Gio 状态对象必须跨帧存活。在 `Layout` 里新建它们，点击永远不会生效。
2. **决定是否被 Column 拉满。** `Column` 会把子组件的最小宽度设成满宽。按钮这类"自身多宽就多宽"的组件，要在 `Layout` 开头把 `gtx.Constraints.Min` 清零。
3. **先处理事件，再绘制。** 同一帧里，事件要在 `Layout` 画自己之前处理完，状态变化这一帧就能画出来。
4. **用户回调一律走 `core.Call(gtx, fn)`。** 它负责让所有窗口重绘。直接调 `fn()`，修改了别处组件时，别处要等到下次有输入才会刷新。
5. **颜色、字号从 `theme` 取，不写死。** 用户改了 `theme.Primary`，你的组件要跟着变。在 `Layout` 里读取，不要在构造函数里读好存下来。
6. **声明语义信息，让 Agent 看得见。** Agent 测试靠 Gio 的语义树知道"页面上有什么"（见 [Agent 端到端测试](automation.md#原理)）。用 `ui/widget/semantics.go` 的 `area` 包住组件，声明角色、名字和状态：

   ```go
   return area(gtx, st.Layout, semantic.Button, semantic.LabelOp(b.text), semantic.EnabledOp(!b.disabled))
   ```

   Gio 自带控件产生的语义节点不一定可靠：禁用的按钮会丢节点，输入框没有名字和内容，文字节点的高度会撑满整个可滚动区域。所以 Keel 的组件都自己声明。角色目前只有 Gio 定义的几种（`Button`、`CheckBox`、`Editor` 等）；新角色（比如链接）用 `DescriptionOp` 标注，再在 `ui/window/automation.go` 的 `snapshot` 里识别。

API 风格与现有组件保持一致：

- 构造函数返回指针，名字就是组件名：`widget.Link(...)`。
- 创建时的配置用链式方法，返回自身：`.Hint(s)`、`.Secondary()`、`.OnChange(fn)`。
- 运行时修改用 `SetXxx`，读取用 `Xxx()` 或 `Value()`。
- 程序调用 `SetXxx` 时不触发 `OnXxx` 回调，只有用户操作才触发。`Field.SetValue` 就是这样处理的，否则"监听变化再回写"会死循环。
- 类型名不能和构造函数同名（Go 的限制），现有的做法是 `Btn`、`Field`、`Check`、`LinkText`。
- 所有组件在 `widget` 一个包里，名字不能冲突。起名前在 `ui/widget/` 里搜一下。

最后加交互测试，放在 `ui/widget/link_test.go` 这样的同名测试文件里：

```go
func TestLinkClick(t *testing.T) {
    n := 0
    h := uitest.New(Link("文档", func() { n++ }))
    h.Click(5, 5)
    if n != 1 {
        t.Fatalf("clicked %d times", n)
    }
}
```

再在 `ui/window/automation_test.go` 里确认 Agent 能看到它、状态正确。然后在 [ui/widget/README.md](../ui/widget/README.md) 的文件表里加一行，更新 [组件与布局](widgets.md)，有必要的话在某个示例里用上它。

## 新增容器

容器只摆放子组件，不调用户回调。最简单的容器可以直接用 `core.Func` 写，放在 `ui/layout/` 下：

```go
// Pad adds equal padding around w.
func Pad(dp unit.Dp, w core.Widget) core.Widget {
    return core.Func(func(gtx C) D { return layout.UniformInset(dp).Layout(gtx, w.Layout) })
}
```

复杂一点的容器参考 `Column` 的实现：用结构体保存子组件列表，在 `Layout` 里转成 Gio 的 `layout.Flex`。需要圆角背景时调用 `Frame`。注意 `ui/layout` 包内部引用的 `layout.` 是 Gio 的 `gioui.org/layout`。

## 新增原生能力

以"读取剪贴板文本"为例走一遍。这个例子在写文档时编译并测试通过，没有合入仓库。

**第一步：写 C 函数。** 追加到 `native/internal/sys/sys_darwin.m`：

```objc
int keel_clipboard_text(char **out){
 @autoreleasepool {
 NSString *s=[[NSPasteboard generalPasteboard] stringForType:NSPasteboardTypeString];
 if(!s){*out=NULL;return 0;}
 *out=strdup(s.UTF8String);return *out?0:100;
 }
}
```

返回值约定：`0` 成功，`1` 无权限，`2` 不支持，`3` 参数错，`4` 线程不对，`5` 超时，`6` 冲突，其他值是失败。`sys_darwin.go` 里的 `status()` 把它们转成 `native.Err*`。需要新的错误类别时，两边一起加。

在 `sys_darwin.h` 里声明：

```c
int keel_clipboard_text(char **out);
```

**第二步：包成 Go 函数。** `sys_darwin.go`：

```go
func ClipboardText() (string, error) {
    var p *C.char
    if err := status(C.keel_clipboard_text(&p)); err != nil {
        return "", err
    }
    if p == nil {
        return "", nil
    }
    defer C.free(unsafe.Pointer(p))
    return C.GoString(p), nil
}
```

C 分配的内存由 Go 侧 `C.free` 释放。不要把 Go 指针交给 C 长期保存。

**第三步：给其他平台加桩。** `sys_other.go`：

```go
func ClipboardText() (string, error) { return "", native.ErrUnsupported }
```

漏了这一步，Linux、Windows 就编译不过。用 `CGO_ENABLED=0 GOOS=linux go vet ./native/...` 检查。

**第四步：公开包。** 新建 `native/clipboard/clipboard.go`，参数校验放在这一层，`sys` 层只做翻译：

```go
// Package clipboard reads the system clipboard.
package clipboard

import "github.com/dyike/keel/native/internal/sys"

// Text returns the clipboard's plain text, or "" when it holds none.
func Text() (string, error) { return sys.ClipboardText() }
```

**第五步：登记模块边界。** 在 `internal/deps/deps_test.go` 的 `allowed` 表里加一行：

```go
"native/clipboard": {"native", "native/internal/sys"},
```

不登记，`TestEveryModuleIsListed` 会失败。这一步强制你想清楚：新模块依赖谁，是否真的独立。

**第六步：写文档。** 在 `native/clipboard/README.md` 写清楚它做什么、依赖什么、怎么单独使用（照抄其他模块的 README 格式）；在 [native/README.md](../native/README.md) 的表格和 [原生能力](native.md) 里各加一节，写清需要什么权限、会不会阻塞、在哪个线程能调用。

涉及主线程的系统 API（AppKit 的大部分 UI 类）要注意：Gio 的事件循环占着主线程，C 代码里用 `dispatch_sync(dispatch_get_main_queue(), ...)` 切过去（参考现有的 `onMain`）。但如果调用方本身就在主线程上，`dispatch_sync` 会死锁，所以先判断 `[NSThread isMainThread]`。

## 新增示例

`examples/<名字>/main.go`，一个示例演示一件事。要能出截图的，参考 `examples/hello` 支持 `-screenshot` 参数。

## 提交前检查

```sh
gofmt -l .                          # 应无输出
go vet ./...
CGO_ENABLED=0 GOOS=linux go vet ./native/... ./cmd/...   # 桩函数齐全
go test -race ./...                 # 包括模块边界检查
go run ./examples/hello -screenshot /tmp/after.png   # 改了样式时对比截图
```
