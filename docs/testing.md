# 测试

界面测试大部分不需要开窗口：`ui/internal/uitest` 测试工具通过 Gio 真实的输入路由送入点击、按键、文字，组件走的代码和真实窗口完全一样。`go test -race ./...` 在本机跑完大约 10 秒。

## 交互测试

`ui/` 下的模块都可以引用它（它在 `internal` 里，外部项目引用不到）：

```go
func TestInputSubmit(t *testing.T) {
    var got string
    f := Input("").OnSubmit(func(s string) { got = s })   // 在 package widget 里
    h := uitest.New(f)   // 布局一帧
    h.Click(20, 10)      // 点击获得焦点
    h.Type("你好")        // 输入文字
    h.Key("⏎", 0)        // 回车
    if got != "你好" {
        t.Fatalf("OnSubmit saw %q", got)
    }
}
```

| 函数 | 作用 |
| --- | --- |
| `uitest.New(w)` | 把组件放在左上角 (0,0) 布局一帧 |
| `uitest.NewFunc(fn)` | 用任意帧函数，比如 `window` 包的测试用 `w.layout` 带上根视图和快捷键 |
| `h.Click(x, y)` | 在 (x, y) 按下并松开左键，然后画一帧 |
| `h.Key(name, mods)` | 按下并松开一个键，然后画一帧。`name` 是 Gio 的键名：字母大写 `"A"`，回车 `"⏎"`，也可以用 `key.NameReturn` 等常量 |
| `h.Type(s)` | 向有焦点的输入框插入文字，然后画一帧 |
| `h.Frame()` | 执行 `core.Update` 队列并画一帧 |

坐标规则：测试视口是 400×300，1dp = 1 像素。`uitest.New` 不带根视图，组件从 (0,0) 开始；用窗口布局测试时，内容从 (24,24) 开始。

每个 `h.Click` 等操作后面自带一帧，但回调修改了状态、要看重绘结果时，多调一次 `h.Frame()`。

## 模块边界测试

`internal/deps` 检查每个模块只引用了允许的包（见[架构 · 模块](architecture.md#模块)）：

- 下层模块引用了上层（比如 `layout` 引用 `widget`），同层互相引用，或者 `native/*` 引用了 `ui`、Gio，测试失败；
- 新增了模块目录却没在允许表里登记，测试失败。

它随 `go test ./...` 一起运行。

## 什么该测

- 回调触发次数：点一次触发一次，禁用时不触发。
- 回调参数：`OnChange` 收到的是新值。
- 程序调用和用户操作的区别：`SetValue` 不触发 `OnChange`，打字会触发。
- 解析类函数（快捷键字符串）：合法输入和每一种非法输入。
- `native` 包：参数校验路径（不需要权限，也不会真的动鼠标）。

## 截图对比

改了主题、间距、字体后，渲染前后截图，确认只有预期的地方变了：

```sh
go run ./examples/hello -screenshot /tmp/before.png
# 修改代码
go run ./examples/hello -screenshot /tmp/after.png
cmp /tmp/before.png /tmp/after.png && echo 完全一致
```

纯重构应该输出"完全一致"。截图由 GPU 离屏渲染，同一台机器上结果稳定；不同机器、不同系统字体下像素可能不同，不要把截图提交成跨机器的基准文件。

## 真实窗口测试

有些 bug 只在真实窗口里出现，比如多窗口之间的死锁。`ui/window/desktop_test.go` 在子进程里运行 `ui/window/testdata/raise`：打开两个窗口，在界面代码里（持有帧锁）调用 `Raise`、`Close`，15 秒内没走完就判定为死锁。

```sh
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window
```

默认跳过，因为它会在屏幕上弹出窗口，而且需要图形界面环境。改动 `ui/window`、`ui/internal/loop` 里和窗口、锁相关的代码后必须跑一次。新增窗口方法时，把它加进 `testdata/raise` 的步骤里。

## 需要手动验证的部分

无界面测试覆盖不到这些，改动相关代码后手动跑一遍：

| 场景 | 怎么验证 |
| --- | --- |
| 真实窗口的打开、关闭、置前 | `go run ./examples/multiwindow`，点"打开设置窗口"两次，应该只有一个设置窗口 |
| 最后一个窗口关闭后退出 | 关掉所有窗口，终端里进程应结束 |
| 窗口快捷键 | 在主窗口按 ⌘+, |
| 全局快捷键 | `go run ./examples/hotkey`，切到别的应用按 ⌘⇧K，计数增加 |
| 后台 `core.Update` | hotkey 示例里的时钟每秒走 |
| 权限、截图、合成输入 | 需要授权，按 [原生能力](native.md) 的说明手动测 |
| 中文输入法 | 在输入框里用拼音输入，候选框位置正确，上屏后内容正确 |
