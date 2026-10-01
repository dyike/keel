# kit 组件规范

新组件放在 `ui/kit`，用 `ui/el` 组织元素、布局和交互。`ui/widget` 只修 bug。M0 建立规范；组件从 M1 开始，焦点与按键接口须经 review 后才能继续扩展。迁移顺序见[设计决策](decisions.md#新组件基于-eluiwidget-冻结)。

## 模块与 API

kit 只直接依赖 Keel 的 `core`、`theme`、`el`；不引用 `widget`、`layout`、`window`。el 缺少的基础能力先在 el 中实现，不在各组件里复制 Gio 输入路由、定时或浮层机制。依赖测试按传递依赖登记 `internal/loop` 和 `internal/editorstyle`，它们不是 kit 的直接依赖。

构造函数 `Xxx(...)` 返回 `*XxxView`。组件以 `Render(*el.Context) el.Element` 接入 el，实例保留业务状态，Render 根据当前状态生成元素树。动态列表使用稳定 ID，不用数组位置代表可移动项目。

有值的交互组件提供 `Value()`、`SetValue(...)`、链式 `OnChange(...)`、`SetDisabled(bool)`。程序赋值不触发回调，只有用户操作触发。按钮等动作组件使用点击回调和 `SetDisabled`，纯展示组件不强加值与回调接口。返回切片等可变值时不能暴露内部存储。

## 状态与布局

回调在帧锁内改状态；后台任务通过 `core.Update` 返回 UI。组件不能根据 `gtx.Enabled()` 为 false 重置跨帧状态：它也可能表示 el 测量或离屏布局没有输入源。关闭菜单、停止拖动等重置放在 `SetDisabled(true)` 中；事件、外观与语义遵守父级禁用状态。

装饰的显示、隐藏和内容变化不应改变宿主尺寸或基线，例如 Badge 计数变化不能让按钮跳动。普通内容变化可以重新排版。窄容器中遵守约束；文本检查中文、拉丁字母、数字、混排、1×/2×缩放，输入光标和选择区使用统一字形度量，不按某个截图硬补偏移。

颜色在 Render 时读取 theme 语义色，不在构造函数中保存主题快照。自定义固定色属于显式覆盖；主题切换不会替应用推断其含义。缓存必须包含主题版本，或使用支持主题失效的 `cx.Cache`。M0 不提供局部主题作用域。

交互组件支持鼠标和键盘；禁用时不能激活或获得焦点，Agent 快照报告 disabled。焦点、按键、禁用、定时、浮层的具体新 API 在对应阶段 review，不在本规范预先定型。

## 文件模板与验收

一个组件对应 `ui/kit/<name>.go`、`<name>_test.go`、`docs/kit/<name>.md`、`examples/components/<name>.go`。在 kit README 和文档索引登记。示例注册独立 `-section <name>`，展示常用状态、边界和浅深色。

组件文档包含用途、最小用法、公开 API、键盘操作（适用时）、语义、边界和验证入口；测试验证用户可观察行为，不复制实现算法。

提交前逐项检查：

- `uitest` 驱动布局与交互；纯展示组件检查尺寸、约束、状态和颜色。
- `ui/window` Agent 快照覆盖名称、角色、值和状态；新增 role 同步 `automation.go` 与 `docs/automation.md`。
- 有交互时覆盖键盘、禁用、恢复、程序赋值不触发回调；嵌入 el 后连续多帧不丢状态。
- 示例可运行，浅深色和文字布局经过截图检查。
- 公开 API、组件文档、README、示例同时更新。
- `go build ./... && go vet ./ui/... && go test ./... -count=1` 全部通过，包括 `cmd/keel-mcp` 端到端测试。

每个组件完成后单独提交，再开始下一个。迁移不删除仍有调用者的旧组件；所有组件和调用者迁完才删除 `ui/widget`。

## 已实现组件

- [Alert](kit/alert.md)：行内状态提示。
- [Empty](kit/empty.md)：空状态说明。
- [Avatar](kit/avatar.md)：图片与姓名回退头像。
- [Tag](kit/tag.md)：不可移除标签。
- [DescriptionList](kit/description_list.md)：字段说明列表。
- [GroupBox](kit/group_box.md)：带标题的视图分组。
- [StatusBar](kit/status_bar.md)：状态与详情栏。
- [Marker](kit/marker.md)：纯图形标记。

Icon：矢量图标，默认颜色随主题切换。

主题文本使用场景：选中底色 Highlight 上使用 PrimaryText；Primary 保留为按钮背景和描边。Markdown 默认使用 CodeBg/CodeText，旧包级颜色变量的零值表示跟随主题，CodeStyle 为空时按背景自动选择 github/github-dark。

[Spinner](kit/spinner.md)：支持减少动画的不确定进度。

[Skeleton](kit/skeleton.md)：占位、圆形和 Shimmer 扫光。
