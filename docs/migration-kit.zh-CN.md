# 迁移到当前 kit

[English](migration-kit.md) | 简体中文

仓库已删除 `ui/widget` 和 `ui/layout`，不提供兼容别名。应用结构用 `ui/el`，通用组件用 `ui/kit`；窗口仍由 `ui/window` 管理。下面列出升级时影响调用方的接口与行为变化。

## 代码高亮和网络图片改为按需引入（v0.0.7）

为了减小程序体积，这两个功能不再默认链接，各约 4 MB：

```go
import (
    _ "github.com/dyike/keel/ui/highlight" // CodeEditor、TextView、Markdown 代码块的语法颜色
    _ "github.com/dyike/keel/ui/netimage"  // Image、Avatar、Attachment、Markdown 图片的 http(s) 地址
)
```

升级后如果代码变成纯文本，或网络图片报 `core.ErrNoImageFetcher`，在 main 包里加上对应的一行即可。本地文件和 data URL 的图片不受影响。

## 页面与状态

旧的 Column / Row / Card 改为 `el.Div()`，按需设置 Row、Gap、Padding、Bg 和 Border；kit 视图通过 `Render(cx)` 加入元素树，页面通过 `el.Root(view)` 交给窗口。旧组件到新组件的具体参数见 [组件索引](kit.zh-CN.md)。

组件实例在页面生命周期内复用，尤其是输入框、浮层、Table、Tree、Dock 和虚拟列表。不要每次 Render 都重建有状态组件。`SetValue` 等程序赋值不触发用户回调；回调内直接修改，后台更新放进 `core.Update`。

集合型组件会复制自己拥有的数据切片或布局树。外部修改原切片不会刷新组件，应调用 SetItems、SetRows、SetKeys、SetData 或 SetLayout。View、图片与业务回调仍按引用持有；复制配置不意味着复制业务对象。

## 聊天列表

`MessageScroller` 现在逐行构建，需要稳定 key、估计行高和行构造函数：

```go
scroller := kit.MessageScroller(keys, 100, func(cx *el.Context, index int) el.Element {
    return messages[index].Render(cx)
})
// 新消息或历史消息到达后，先更新 messages，再提交完整且不重复的 key 列表。
scroller.SetKeys(keys)
```

删除旧的整篇列表构造回调和 `HistoryPrepended` 调用。头部插入与高度变化由稳定 key 保持阅读锚点；跟随到底部用 `SetFollow` / `ScrollToEnd`。不要用当前数组下标作为会插入或重排的消息 ID。完整示例在 `examples/chat`。

## 表格、选择与表单

有重排、筛选或动态更新的列表与树应使用稳定标识。Table 列布局、选择和数据交互见 [Table](kit/table.zh-CN.md)，不依赖返回切片的别名修改内部状态。Select / Combobox 区分标签和值，并提供分组、禁用项、多选和异步结果接口。

异步表单与搜索通过对应请求 token 提交结果；不能让较早请求覆盖新请求。组件所属区域禁用也会阻止用户修改，应用不必为每个子控件重复注册禁用回调。调用方仍负责取消外部网络任务。

## Dock 持久化

`DockLayout` 当前版本为 2，增加 LeftTree / RightTree / BottomTree 嵌套树。旧的无版本布局和版本 1 可由 `SetLayout` 迁移；必须检查其 bool 返回值。非法或未知版本不会部分覆盖当前布局。保存 `Layout()` 返回的快照，不缓存内部树指针。窗口内拖放由 Dock 处理；分离到新窗口需要应用设置 `OnDetach` 并负责打开窗口，关闭时调用 `reattach`。见 [跨窗口](kit/dock.zh-CN.md#跨窗口)。

## 窗口与主题

自定义 `core.WindowControls` 实现需要补 `Focused()` 与 `TitleBarArea(...)`。前者表示原生窗口激活状态，后者接受窗口 dp 坐标；每帧清空再由标题栏登记，应用控件不应包含在拖动区域里。

运行时主题用 `theme.Apply`，不要逐个修改全局颜色。减少动画用 `SetReducedMotion` 显式覆盖，或 `FollowSystemMotion` 恢复系统偏好；当前系统桥接支持 macOS。其他平台可由应用指定偏好。

## 验证

迁移后先运行 `go build ./...`、`go vet ./...`、`go test ./... -count=1`。页面涉及订单表单、聊天滚动或 Dock 状态时，保留对应行为测试；视觉修改按 [视觉规范](visual-guidelines.zh-CN.md) 和 [截图矩阵](testing.zh-CN.md#全组件截图矩阵) 复核。
