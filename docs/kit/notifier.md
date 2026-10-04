# Notifier

按位置分组显示窗口内通知，默认在右上角。

```go
n := kit.Notifier()
// Render 中：作为根视图最外层元素的直接子元素
root.Child(n.Render(cx))
// 回调中：
n.Notify(kit.Notice{Title: "保存成功", Body: "订单已更新", Tone: kit.ToneSuccess})
// 后台 goroutine 中：
core.Update(func() { n.Notify(kit.Notice{Title: "同步完成"}) })
```

- 超时：`Timeout` 为 0 时用 `kit.NotificationTimeout`（5 秒）；为负数时不自动消失。
- 悬停：指针停在通知上时，这条通知不会消失；移开后继续剩余时间，键盘焦点进入通知或通知被禁用/遮挡时也暂停。
- 数量：每个位置最多同时显示 `kit.MaxNotifications`（5）条，超出的排队，前面的消失后依次显示。
- `Notify` 返回 id，`Dismiss(id)` 移除指定通知（不论在显示还是排队），`Len()` 返回管理中的通知总数；纯应用内模式为显示加排队，系统模式还包含仅系统通知及应用内已超时但等待撤回的记录。
- 每条通知都有关闭按钮。通知不抢焦点，也不处理 Esc，下面的对话框仍然可以用 Esc 关闭。
- `Notify`、`Update`、`Dismiss` 必须在 UI 帧锁内调用，也就是在回调里，或者用 `core.Update` 包起来。需要 `el.Root`。

Agent：每条通知的角色是 `status`，名字是标题，`value` 是 Tone 名称（neutral/info/success/warning/danger）；关闭按钮名为"关闭 标题"。

验证：`go run ./examples/components -section notifier`。

`Update(id, Notice)` 原位替换通知、保留 ID 和队列位置，重新开始该条通知的超时；不存在的 ID 返回 false。后台任务用 `core.Update` 包住调用。队列中的通知从实际显示时才开始计时；窄窗口会限制通知栈宽高，长通知栈可以滚动。

`Notifier.Placement(position)` 设置容器默认位置，现有未指定位置的通知也会跟随移动。`Notice.Placement` 为单条通知覆盖位置；默认零值 `NoticeDefault` 跟随容器。容器传 NoticeDefault 恢复右上角，非法默认值忽略，单条非法值按跟随默认处理。

支持 NoticeTopLeft、NoticeTopCenter、NoticeTopRight、NoticeLeftCenter、NoticeRightCenter、NoticeBottomLeft、NoticeBottomCenter、NoticeBottomRight。每个位置独立排队，组内保持发送顺序从上往下排列；更新 Notice 可将其移入其他位置。移动默认位置不重置超时，Update 仍按原约定重启该条超时。

```go
n.Placement(kit.NoticeBottomRight)
n.Notify(kit.Notice{Title: "下载完成", Placement: kit.NoticeBottomLeft})
```

默认位置属于 Notifier 实例，不写入全局主题。多个位置的通知栈分别限制在窗口范围内，不自动避让其他栈；窄窗口同时启用多个位置时可能重叠。


`Notice.Content` 接受任意 `el.View`，非 nil 时替换 Body，Title 仍显示并作为通知的可访问名称；nil 恢复普通正文。可放富文本、图片或交互控件。`Notice.Action` 是正文下方的独立 View 槽，可放一个完整配置的 Button，或组合多个操作。点击操作不会自动关闭通知，需要时在回调里调用 `Dismiss(id)`。

```go
var id int
id = n.Notify(kit.Notice{
    Title: "连接中断", Timeout: -1,
    Content: kit.Label("请检查网络后重试。"),
    Action: kit.Button("重试", func() {
        retry()
        n.Dismiss(id)
    }).Variant(kit.ButtonPrimary),
})
```

正文和操作区保留独立稳定身份；Update 可替换或清空这两个槽。复用有状态 View 时应保持实例稳定，不要把同一个交互实例同时放进多条通知。自定义内容应适应通知可用宽度。焦点进入任一子控件都会暂停倒计时，所属容器禁用会同时禁用内容和操作；排队通知在实际显示后才渲染内容。

不同位置属于不同浮层；移动通知到另一个位置时不保留子控件键盘焦点。需要连续编辑的通知应保持位置固定。


`Notice.OnClick` 在通知背景、标题或普通正文被点击时执行，不自动关闭通知。设置后增加一个以 Title 命名的 button 操作区域，支持 Tab 聚焦及 Enter/Space 激活，外层仍保留 status 语义。富内容和操作按钮独立处理各自点击，关闭按钮不会执行 OnClick。移除回调后恢复普通通知。

`Notice.OnClose` 在应用内通知关闭后同步执行一次，覆盖关闭按钮、超时、显式 Dismiss 和排队中取消（仅系统模式除外）；重复或未知 ID 不触发。回调可安全再次 Dismiss 同一 ID、更新其他通知或新增通知。Update 替换回调但不触发关闭；最终关闭使用最新回调。卸载 Notifier 不等同于 Dismiss，不触发 OnClose。回调运行在 UI 帧锁内，后台工作应异步执行，回写时用 core.Update。


`NotifyKey(key, Notice)` 用业务字符串标识通知，作用域限当前 Notifier。重复发送同一非空 key 会原位替换，返回原 ID、保留队列位置并重启超时，不触发旧 OnClose。Notice 的正文、操作、回调和位置都会被新值替换。空 key 等同普通 Notify，每次新增。

```go
n.NotifyKey("download/report", kit.Notice{Title: "正在下载", Timeout: -1})
n.NotifyKey("download/report", kit.Notice{Title: "下载完成", Tone: kit.ToneSuccess})
n.DismissKey("download/report")
```

`Update(id, Notice)` 保留业务 key；`DismissKey(key)` 删除显示中或排队中的对应通知并返回是否找到，空 key 返回 false。删除后重新发送同一 key 会分配新 ID，旧超时不会影响新通知。不同业务应自行设置 key 前缀，避免同一容器内冲突；不使用 Rust 类型作为标识。

`Clear()` 先移除调用时的全部通知，再按原队列顺序执行各自的 OnClose，返回移除数量。回调中新增的通知保留，除非后续回调显式删除它；重复清空不会重复通知已移除项。以上方法同样要求在 UI 帧锁内调用。系统通知的撤回行为见下文。


系统投递通过 `NoticeSystemBackend` 接入，kit 不直接依赖原生模块。接口的 `Post(id,title,body,done)` 和 `Remove(id,done)` 在工作 goroutine 执行，必须每次完成后调用 done（失败也要调用）。Notifier 串行等待每次完成，再执行下一条请求；后端不得直接修改 UI。原生适配器和权限申请见 `examples/notification`。

```go
n := kit.Notifier().SystemBackend(backend, func(r kit.NoticeSystemResult) {
    // 已在 UI 帧锁内，可以展示 r.Err；Removing 区分投递和撤回。
})
n.NotifyKey("download", kit.Notice{
    Title: "下载完成", Body: "report.pdf 已保存。",
    Delivery: kit.NoticeInAppAndSystem,
})
```

- `NoticeInApp` 仅应用内，`NoticeSystemOnly` 仅系统，`NoticeInAppAndSystem` 两者同时。`NoticeDeliveryDefault` 使用容器默认；`Notifier.Delivery(mode)` 只影响后续 Notify/Update，不迁移已有通知，初始默认仅应用内。
- 系统正文只取 Title/Body；富内容、按钮和 Tone 不发送。标题正文都空时跳过系统请求。系统模式不自动申请权限；应用应先完成平台授权。
- 两者同时投递时，应用内超时只隐藏卡片并触发一次 OnClose，保留系统通知和业务 key；重新发送同 key 会更新同一系统 ID 并重新显示卡片。Dismiss/DismissKey/Clear 才请求撤回，隐藏卡片不会重复触发 OnClose。仅系统通知不触发 OnClose，不占应用内五条可见队列。
- 系统 ID 使用随机容器前缀与通知序号，避免不同容器/进程互相覆盖；不跨启动恢复，跨启动撤回需应用自行管理原生接口。
- 已有通知保留创建时的后端；更换 SystemBackend 只影响新通知，结果回调在请求排队时捕获。请在首次 Notify 前配置后端。切换已有通知到仅应用内，或更新为空系统正文，会撤回先前系统通知。
- 缺后端返回 `ErrNoticeSystemUnavailable`，不会静默视为成功。`SystemError()` 是最近已完成请求的错误，成功会清空；结果回调在后续 UI 帧执行。投递失败时，两者模式仍显示应用内卡片；仅系统模式不自动改成应用内。
- 关闭窗口/卸载 Notifier 不会自动撤回系统通知；需要时显式 Clear。后端不调用 done 会阻塞该容器后续系统请求。

支持点击响应的后端可实现 NoticeSystemInteractiveBackend，见下文。当前测试验证投递状态机和请求顺序，不等于 macOS/Linux 通知中心的真实展示验收。


`NoticeSystemInteractiveBackend` 在基本后端上增加 `PostInteractive(id,title,body,activated,done)`。后端收到系统打开动作后调用 activated，kit 自动切回 UI 帧处理；即使 Notice.OnClick 为空也会处理打开动作。普通后端继续通过 Post 投递，无法提供系统点击响应。

`OnSystemActivate(fn)` 设置应用的窗口唤起回调，例如 `func(){ if !w.Closed() { w.Raise() } }`。响应先从管理列表取出通知、安排系统撤回，再依次调用窗口唤起、尚未执行过的应用内 OnClose、最新 Notice.OnClick。重复响应、已删除通知和已切换为纯应用内的通知被忽略。应用内已超时仍可响应系统打开动作，但不重复调用 OnClose；仅系统模式不调用 OnClose。

示例在 macOS/Linux/Windows 选择支持点击的原生适配器；Linux 服务需要 actions 能力。示例在 OnSystemActivation 里记下令牌，在 OnSystemActivate 里调用 `Window.Activate(token)`：X11 把令牌作为启动 ID 写到窗口上，再带着其中的时间戳请求激活，窗口管理器据此放行而不是只闪任务栏；Wayland 把令牌交给 xdg-activation 激活窗口的 surface；合成器不支持时，以及在其他平台上，退回 Raise。窗口实际能否置前由操作系统决定，原生点击和窗口唤起仍待真实系统验收。

`NoticeSystemActivationBackend` 增加 `PostActivated(id,title,body,activated,done)`，其中 activated 接收 `NoticeActivation{Token: ...}`。后端同时实现两种交互接口时优先使用带激活数据的接口；只实现 PostInteractive 的后端继续工作，激活数据为空。

`OnSystemActivation(func(NoticeActivation))` 在 UI 帧内接收数据，顺序为：移除通知并安排撤回 → OnSystemActivation → OnSystemActivate → 尚需执行的 OnClose → OnClick。重复事件或通知已删除时不再交付令牌。应用在这个回调中把令牌交给自己的窗口后端；组件不会修改环境变量或推断令牌格式。
