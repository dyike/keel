# 扩展指南

新增东西之前，先判断它属于哪个现有模块。绝大多数情况是给现有模块加一个文件；新建模块的标准见[架构 · 什么时候新建模块](architecture.md#什么时候新建模块)。

| 你要加的东西 | 放在哪里 | 例子 |
| --- | --- | --- |
| 应用里的界面、业务组件 | 不进库：用 `ui/el` 写成函数或视图，见[元素与视图](el.md#做成可复用的组件) | 订单卡片、工具栏 |
| 通用的元素能力 | `ui/el` | 新的样式方法、布局特性、元素类型 |
| 通用组件 | `ui/kit/` 新文件 | `select.go`、`chart.go`、`dock.go` |
| 颜色、字号等可调参数 | `ui/theme/theme.go` | 被两个以上组件用到的数值 |
| 与窗口本身有关 | `ui/window/` | 窗口位置、置顶 |
| 系统能力，不开窗口也有用 | `native/` 下新模块 | `native/clipboard`、`native/notify`、`native/tray` |
| 只有一个页面用到的界面片段 | 不进库，业务代码里写成返回 `el.Element` 的函数 | 用到第二次再考虑挪进来 |

## 新增组件

通用组件放在 `ui/kit`，一个组件一个文件，用 `ui/el` 写成视图。完整规范和验收清单在 [kit 组件规范](kit.md)，这里只走一遍骨架。以一个计数器为例（kit 已有 `NumberInput`，这里只为说明结构）：

```go
// CounterView shows a number between − and + buttons.
type CounterView struct {
    value    int
    disabled bool
    onChange func(int)
}

func Counter(value int) *CounterView { return &CounterView{value: value} }

func (v *CounterView) Value() int         { return v.value }
func (v *CounterView) SetValue(n int)     { v.value = n } // 程序赋值不触发回调
func (v *CounterView) SetDisabled(d bool) { v.disabled = d }
func (v *CounterView) OnChange(fn func(int)) *CounterView {
    v.onChange = fn
    return v
}

func (v *CounterView) set(n int) {
    v.value = n
    if v.onChange != nil {
        v.onChange(n) // 只有用户操作触发
    }
}

func (v *CounterView) Render(cx *el.Context) el.Element {
    minus := Button("−", func() { v.set(v.value - 1) }).Variant(ButtonSecondary)
    plus := Button("+", func() { v.set(v.value + 1) }).Variant(ButtonSecondary)
    minus.SetDisabled(v.disabled)
    plus.SetDisabled(v.disabled)
    return el.Div().Row().Gap(8).Items(el.Center).Child(
        minus.Render(cx),
        el.Text(strconv.Itoa(v.value)).TextColor(theme.Text), // 颜色在 Render 时读取
        plus.Render(cx),
    )
}
```

要点：

1. **构造函数 `Xxx(...)` 返回 `*XxxView`。** 业务状态存在结构体里；悬停、按下、焦点这类交互状态由 el 按元素位置保存，不用声明。
2. **有值的组件提供 `Value`、`SetValue`、`OnChange`、`SetDisabled`。** 程序赋值不触发回调。
3. **颜色和框架文字在 Render 时从 `theme`、`locale` 读取**，不在构造时保存，也不写死中文。
4. **el 缺的能力先加到 el**（焦点、定时、浮层、拖动），不在组件里直接写 Gio 输入路由。

然后补齐配套文件，`ui/kit/conventions_test.go` 会检查缺了哪个：

- `ui/kit/counter_test.go`：用 `page(v)`、`click(t, h, "名字")` 等辅助函数（在 `kit_test.go`）走真实输入路由；
- `ui/window/kit_*_test.go`：Agent 快照里角色、名字、值、状态正确；需要单独列出子元素的容器角色加入 `containerRoles`，并补进 [Agent 端到端测试](automation.md#元素)的表；
- `docs/kit/counter.md`，并在 [kit.md](kit.md) 的索引和 `ui/kit/README.md` 的表里登记；
- `examples/components/counter.go`，注册 `-section counter`。

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
