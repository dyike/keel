# Terminal

[English](terminal.md) | 简体中文

Terminal() 是纯 Go ANSI 显示组件，输入／输出连接由应用拥有。它不会启动 shell、建立网络连接或执行 OSC 剪贴板指令。将 OnData(func(string)) 接到进程、PTY、WebSocket 或 SSH 通道，把收到的字节传给 Write；也提供 WriteString，均返回常规 writer 返回值。后台调用与其他 Keel 视图一样，需要放在 core.Update 或 core.UpdateIf 中。禁用输入时仍可接收输出，输出不会触发 OnData。

Name 设置文本框名称。Height 设置嵌套视口高度，默认 320dp；Fill 使用可用空间，el.Root 可填满窗口。Size 返回字符列／行数；OnResize 在真实可见布局的尺寸变化后通知，不在测量阶段通知，首次可见布局也会报告。最大 512 列、256 行。调整尺寸会保留并限制当前屏幕，但不重排旧文本。Scrollback 设置保留的历史行数，范围 0–10000，默认 1000。Reset 清空屏幕、历史、解析模式及选区，不发送输入。Text 返回纯文本屏幕／历史，包含屏幕空行。

ANSI 支持 16／256／RGB 颜色、粗体、暗色、斜体、下划线、反色和删除线；回车、换行、退格、制表、延迟折行；光标定位／保存／恢复；擦除屏幕／行／字符；插入／删除字符／行；滚动区域；备用屏幕 1049；光标可见性、应用光标及括号粘贴模式。UTF-8 和转义序列可分块写入。宽字符、组合字符和 ZWJ 使用字素感知字符格。历史及解析缓存均有上限。不支持的控制字符串被消费，不执行。尚不等同于完整 xterm／VT：鼠标报告、终端查询、替代字符集、冒号式 SGR、高级键盘协议及历史重排尚不支持。

输入不本地回显；应用自行写回回显。提交文本、回车、退格、Tab、Escape、Ctrl 字母组合、方向键、Home/End、Delete、PageUp/PageDown 和 F1–F12 转换为终端序列；应用光标模式改变方向键序列。Escape 后按 Tab 可离开焦点，普通 Tab 发送给会话。鼠标点击／触摸轻点聚焦终端。输入法组合文本在本地显示，组合结束后提交一次；事件测试不等于真实平台输入法验收。

鼠标拖动选择。Ctrl+Shift+C（或 Super+C）复制 Selection；Ctrl+C 始终发送字节 3。Ctrl+Shift+V（或 Super+V）读取剪贴板文本，启用括号粘贴模式时添加对应包裹序列。粘贴上限 1MiB。复制采用显示行边界并保留完整宽字素。无关输出保留选区；改写所选文本或淘汰选中历史时清除选区。滚轮／触摸可滚离底部，后续输出保留当前位置并补偿淘汰历史；新的键盘输入恢复跟随输出。SetDisabled 阻止交互并取消临时输入／粘贴，静默输出方法仍可用；遵循祖先禁用状态。

仅绘制可见行。ANSI 背景及字体样式使用已有 Gio／theme 字形渲染器，包括平台支持的彩色 emoji。使用静态光标，空闲终端不安排闪烁帧。无障碍通过已有 textbox 角色提供可见文本及禁用状态。字体覆盖、物理设备输入法、实际 PTY 应用和客户端内存／帧性能仍需单独验收。

组件示例是进程内命令演示（help、colors、progress、echo、clear），不是操作系统 shell。进度命令使用有限次数的定时更新，Ctrl+C 可取消。生产环境的传输、进程生命周期、背压及输出调度由应用负责。

验证：go test ./ui/kit ./ui/window -run TestTerminal；go run ./examples/components -section terminal。
