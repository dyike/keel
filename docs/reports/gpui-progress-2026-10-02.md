# GPUI Kit 实现进度 · 2026-10-02

更新日期：2026-10-03（原报告 2026-10-02，代码基准 `bc54e85`）。来源：[GPUI Kit 组件目录](https://gpui-kit.com/component/)（页面版本 v0.7.0），按导航中的独立组件链接去重，共 **77 项**。组件分类参考该站，说明和实现判断根据 Keel 当前工作区重写；源站文档采用 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)。这是一份能力对照，不要求复制 Rust API。

“已有”表示 Keel 提供可复用的基础组件，并不表示与 GPUI Kit 功能完全一致；“部分”表示已有实现，但仍缺本表列出的关键能力；“未实现”表示缺少通用实现。最后一列列出建议补齐的能力，不是对源站全部配置项的逐项认证。

## 当前实施清单

原差距实施清单：原 A–F 共 36 项，35 项已完成，F3 的完整原生场景验收仍未完成。后续新增能力单列记录，不混入原清单分母；清单完成率不等于 GPUI Kit 功能对齐率。

- [x] WebAssembly：`28a0e80`，浏览器 hello 示例已验证中文、输入、复选框和按钮；构建需 `-tags osusergo`，见 [Web](../../docs/web.md)。
- [x] 视觉基础：`83c50ed`，圆角/字号/阴影刻度、透明度、字重/等宽/行高与 kit 样式迁移；间距刻度 `08daebf` 已统一。
- [x] Dock 最大化：`bf9f69d`，菜单/双击进入、Esc 恢复、布局保存；中心区文档标签组与拆分、跨窗口分离见下表。
- [x] 多主题基础：`bf9f69d`，主题注册、JSON 配色、运行时切换、Nord/Paper 示例；目录监听、渐变配置 `08daebf` 已完成。
- [x] CodeEditor 基础：`4f22179`，行号、高亮、撤销重做、输入法、自动缩进、诊断/补全/悬停接口；高级能力见下表。
- [x] 编辑器验证与补全修复：`bc54e85`，真机确认补全出现和回车接受；回归测试覆盖 Agent 补全项、回车/点击接受与关闭。20 万行已实际载入并验证滚动、末尾跳转和输入，未采集帧率或输入延迟。
- [x] 后续补齐（2026-10-03）：编辑器多光标/查找替换/折叠/括号 `56fa4f6`；局部主题与内置主题 `cc8e536`；Windows、Linux 原生能力 `643ceb6`（仅交叉编译验证）；无样式基础层 `d4d38b5`；状态栏溢出与动作键位 `3c4e511`；主题渐变/间距/目录监听 `08daebf`；HTML 与扩展 TeX `2bd4c61`；Dock 中心文档与跨窗口（本次提交）。
- [ ] 完整原生验收：仍需复核标题栏、系统偏好、多窗口等；代码编辑器真机可运行不代表 F3 全部通过。
- 系统读屏 / VoiceOver：**暂缓**，未接入，不计为已完成。

## 当前组件对照

保留原 77 项目录口径，仓库位置已更新到当前实现。多个目录项可由同一组件或底层能力承接；“—”表示此表没有登记进一步缺口，不表示已认证源站全部 API。原生验证限制见 F3。

| GPUI Kit 组件 | 状态 | Keel 当前能力 / 位置 | 待补齐 |
| --- | --- | --- | --- |
| [Accordion](https://gpui-kit.com/component/accordion/) | 已有 | [单项/多项、自定义标题、动画、键盘、禁用](../../ui/kit/accordion.go) | — |
| [AlertDialog](https://gpui-kit.com/component/alert-dialog/) | 已有 | [提示/确认/危险对话框、焦点约束与恢复](../../ui/kit/dialog.go) | — |
| [Alert](https://gpui-kit.com/component/alert/) | 已有 | [行内提示、级别、关闭按钮](../../ui/kit/alert.go) | — |
| [Attachment](https://gpui-kit.com/component/attachment/) | 已有 | [附件卡片、进度、取消、重试、错误状态](../../ui/kit/attachment.go) | — |
| [Avatar](https://gpui-kit.com/component/avatar/) | 已有 | [图片/首字母回退、尺寸、状态标记](../../ui/kit/avatar.go) | — |
| [Badge](https://gpui-kit.com/component/badge/) | 已有 | [数字、圆点、图标、尺寸、颜色、角标](../../ui/kit/badge.go) | — |
| [Bubble](https://gpui-kit.com/component/bubble/) | 已有 | [可复用聊天气泡、用户操作栏](../../ui/kit/message.go) | — |
| [Button](https://gpui-kit.com/component/button/) | 已有 | [主/次要/危险、禁用、尺寸、图标、加载](../../ui/kit/button.go) | — |
| [Calendar](https://gpui-kit.com/component/calendar/) | 已有 | [年月切换、多月、范围、禁用日期、键盘](../../ui/kit/calendar.go) | — |
| [Carousel](https://gpui-kit.com/component/carousel/) | 已有 | [轮播、指示器、键盘、禁用与定时暂停](../../ui/kit/carousel.go) | — |
| [Chart](https://gpui-kit.com/component/chart/) | 已有 | [折线/柱状/面积，另有饼图/环图/蜡烛图、图例与数据表](../../ui/kit/chart.go) | — |
| [Checkbox](https://gpui-kit.com/component/checkbox/) | 已有 | [布尔选择、半选、回调、禁用](../../ui/kit/checkbox.go) | — |
| [Clipboard](https://gpui-kit.com/component/clipboard/) | 已有 | [通用复制按钮、提示与连续复制反馈](../../ui/kit/copy_button.go) | — |
| [Collapsible](https://gpui-kit.com/component/collapsible/) | 已有 | [独立 Trigger/Content、动画、焦点恢复](../../ui/kit/collapsible.go) | — |
| [ColorPicker](https://gpui-kit.com/component/color-picker/) | 已有 | [HSV、透明度、HEX、预设、键盘、禁用](../../ui/kit/color_picker.go) | — |
| [Combobox](https://gpui-kit.com/component/combobox/) | 已有 | [过滤、多选标签、异步结果、重试、虚拟化](../../ui/kit/combobox.go) | — |
| [Command](https://gpui-kit.com/component/command/) | 已有 | [模糊过滤、分组、快捷键、异步结果、虚拟化](../../ui/kit/command.go) | — |
| [DataTable](https://gpui-kit.com/component/data-table/) | 已有 | [横向滚动、冻结列、列管理、多选/单元格选择、复制、筛选、分页加载](../../ui/kit/table.go) | — |
| [DatePicker](https://gpui-kit.com/component/date-picker/) | 已有 | [日期输入、日历弹层、范围、取消草稿、键盘](../../ui/kit/date_picker.go) | — |
| [DescriptionList](https://gpui-kit.com/component/description-list/) | 已有 | [标签/值布局、响应式列数](../../ui/kit/description_list.go) | — |
| [Dialog](https://gpui-kit.com/component/dialog/) | 已有 | [可组合内容、嵌套浮层、长内容、焦点约束与恢复](../../ui/kit/dialog.go) | — |
| [Dock](https://gpui-kit.com/component/dock/) | 已有 | [边缘与中心区标签组、嵌套分割、拖放、布局保存、最大化、跨窗口分离](../../ui/kit/dock.go) | — |
| [DropdownButton](https://gpui-kit.com/component/dropdown_button/) | 已有 | [按钮菜单、分体按钮、键盘与焦点恢复](../../ui/kit/dropdown_button.go) | — |
| [Editor](https://gpui-kit.com/component/editor/) | 已有 | [行号、局部重高亮、多光标/矩形选择、查找替换、折叠、语法感知括号配对、诊断/补全/悬停/定义跳转接口](../../ui/kit/code_editor.go) | 未接入 Tree-sitter（用 chroma 词法与局部重高亮代替）；LSP 客户端由应用提供 |
| [Empty](https://gpui-kit.com/component/empty/) | 已有 | [空状态标题、说明与操作](../../ui/kit/empty.go) | — |
| [Focus Trap](https://gpui-kit.com/component/focus-trap/) | 已有 | [弹层焦点循环、关闭后返回焦点](../../ui/el/overlay.go) | — |
| [Form](https://gpui-kit.com/component/form/) | 已有 | [字段组织、校验、错误聚焦、异步提交/取消](../../ui/kit/form.go) | — |
| [GroupBox](https://gpui-kit.com/component/group-box/) | 已有 | [标题、描述与内容分组](../../ui/kit/group_box.go) | — |
| [HoverCard](https://gpui-kit.com/component/hover-card/) | 已有 | [悬停卡片、延迟、定位、跨目标与取消](../../ui/kit/hover_card.go) | — |
| [Icon](https://gpui-kit.com/component/icon/) | 已有 | [内置矢量图标、自定义图标、尺寸与颜色](../../ui/kit/icon.go) | 按应用需要扩充图标 |
| [Image](https://gpui-kit.com/component/image/) | 已有 | [异步缓存、适配/裁剪/拉伸、圆角、预览、失败重试](../../ui/kit/image.go) | — |
| [Input Group](https://gpui-kit.com/component/input-group/) | 已有 | [独立组合容器、前后内容、统一边框、标签聚焦](../../ui/kit/input_group.go) | — |
| [Input](https://gpui-kit.com/component/input/) | 已有 | [单行、密码、长度、前后缀、清空、校验、禁用、标签聚焦](../../ui/kit/input.go) | — |
| [Kbd](https://gpui-kit.com/component/kbd/) | 已有 | [平台键帽、动态文案、尺寸、无边框、`KbdFor` 按动作显示键位表绑定](../../ui/kit/kbd.go) | — |
| [Label](https://gpui-kit.com/component/label/) | 已有 | [Text、字号/颜色、For 标签关联聚焦](../../ui/el/element.go) | — |
| [List](https://gpui-kit.com/component/list/) | 已有 | [列表项、稳定 ID、单项禁用、多选/范围、键盘、拖动](../../ui/kit/list.go) | — |
| [Marker](https://gpui-kit.com/component/marker/) | 已有 | [标记形状、大小和颜色](../../ui/kit/marker.go) | — |
| [Menu](https://gpui-kit.com/component/menu/) | 已有 | [菜单、子菜单、分隔线、长内容、方向键、焦点恢复](../../ui/kit/menu.go) | — |
| [MessageScroller](https://gpui-kit.com/component/message-scroller/) | 已有 | [可变高度虚拟化、跟随尾部、流式增高、历史加载锚点](../../ui/kit/message_scroller.go) | — |
| [Message](https://gpui-kit.com/component/message/) | 已有 | [消息内容、状态、操作栏、反应、失败重试](../../ui/kit/message.go) | — |
| [Notification](https://gpui-kit.com/component/notification/) | 已有 | [通知队列、超时、关闭、暂停与原位更新](../../ui/kit/notification.go) | — |
| [NumberInput](https://gpui-kit.com/component/number-input/) | 已有 | [数值解析、范围/步长/精度、草稿提交与取消](../../ui/kit/number_input.go) | — |
| [OtpInput](https://gpui-kit.com/component/otp-input/) | 已有 | [分格输入、粘贴、退格、焦点移动、窄布局](../../ui/kit/otp_input.go) | — |
| [Pagination](https://gpui-kit.com/component/pagination/) | 已有 | [页码、前后翻页、总数、窄布局换行](../../ui/kit/pagination.go) | — |
| [Plot](https://gpui-kit.com/component/plot/) | 已有 | [缩放、平移、数据拾取、键盘、异常数据保护](../../ui/kit/plot.go) | — |
| [Popover](https://gpui-kit.com/component/popover/) | 已有 | [锚点定位、避让、长内容、外部点击/Esc、焦点恢复](../../ui/kit/popover.go) | — |
| [Progress](https://gpui-kit.com/component/progress/) | 已有 | [确定进度、不确定动画、模式切换](../../ui/kit/progress.go) | — |
| [Questionnaire](https://gpui-kit.com/component/questionnaire/) | 已有 | [题型、答案模型、校验、分页、禁用与提交快照](../../ui/kit/questionnaire.go) | — |
| [Radio](https://gpui-kit.com/component/radio/) | 已有 | [单选组、横纵布局、独立 Item、单项禁用、键盘](../../ui/kit/radio_group.go) | — |
| [Rating](https://gpui-kit.com/component/rating/) | 已有 | [评分、半星/小数展示、只读、键盘](../../ui/kit/rating.go) | — |
| [Resizable](https://gpui-kit.com/component/resizable/) | 已有 | [横纵分割、最小尺寸、拖动、键盘、取消与禁用](../../ui/kit/resizable.go) | — |
| [Root View](https://gpui-kit.com/component/root/) | 已有 | [根布局、统一浮层宿主、窗口快捷键](../../ui/el/root.go) | — |
| [Scrollable](https://gpui-kit.com/component/scrollable/) | 已有 | [ScrollX/ScrollY、滚动条拖动/轨道点击、定位与尾部跟随](../../ui/el/viewport.go) | 通过 el 组合，没有独立 kit.Scrollable 类型 |
| [Select](https://gpui-kit.com/component/select/) | 已有 | [过滤、分组、多选、禁用项、万条虚拟化](../../ui/kit/select.go) | — |
| [Settings](https://gpui-kit.com/component/settings/) | 已有 | [设置分组、导航、搜索、窄布局](../../ui/kit/settings.go) | — |
| [Sheet](https://gpui-kit.com/component/sheet/) | 已有 | [侧边抽屉、遮罩、长内容、焦点与禁用继承](../../ui/kit/sheet.go) | — |
| [Shimmer](https://gpui-kit.com/component/shimmer/) | 已有 | [Skeleton 的扫光选项、减少动画](../../ui/kit/skeleton.go) | — |
| [Sidebar](https://gpui-kit.com/component/sidebar/) | 已有 | [嵌套分组、收起、选中、固定头尾、键盘滚动](../../ui/kit/sidebar.go) | — |
| [Skeleton](https://gpui-kit.com/component/skeleton/) | 已有 | [占位形状、尺寸、加载展示](../../ui/kit/skeleton.go) | — |
| [Slider](https://gpui-kit.com/component/slider/) | 已有 | [单值/双端范围、横向/竖向、步长、拖动与键盘](../../ui/kit/slider.go) | — |
| [Spinner](https://gpui-kit.com/component/spinner/) | 已有 | [不确定动画、减少动画、可访问名称](../../ui/kit/spinner.go) | — |
| [StatusBar](https://gpui-kit.com/component/status-bar/) | 已有 | [固定状态栏、左右内容组、按优先级收起的溢出菜单](../../ui/kit/status_bar.go) | — |
| [Stepper](https://gpui-kit.com/component/stepper/) | 已有 | [步骤状态、导航、键盘、横向滚动与禁用](../../ui/kit/stepper.go) | — |
| [Switch](https://gpui-kit.com/component/switch/) | 已有 | [布尔开关、尺寸、加载、禁用](../../ui/kit/switch.go) | — |
| [Table](https://gpui-kit.com/component/table/) | 已有 | [排序、行选择、单元格插槽、列宽调整；高级能力同 DataTable](../../ui/kit/table.go) | — |
| [Tabs](https://gpui-kit.com/component/tabs/) | 已有 | [页面状态、溢出、关闭与焦点恢复、拖动排序](../../ui/kit/tabs.go) | — |
| [Tag](https://gpui-kit.com/component/tag/) | 已有 | [颜色、移除、选中](../../ui/kit/tag.go) | — |
| [TextView](https://gpui-kit.com/component/text-view/) | 已有 | [Markdown、HTML 富文本、扩展 TeX、图片、选择复制、代码块、流式渲染](../../ui/markdown) | 不是完整 TeX 引擎；不支持 CSS 样式 |
| [Textarea](https://gpui-kit.com/component/textarea/) | 已有 | [多行、只读、自动高度、最大可见行数、错误恢复](../../ui/kit/input.go) | — |
| [Theme](https://gpui-kit.com/component/theme/) | 已有 | [语义配色、间距/字号/圆角/阴影刻度、浅深切换、注册与 JSON 主题、局部作用域、渐变、目录监听](../../ui/theme/registry.go) | — |
| [TimeField](https://gpui-kit.com/component/time-field/) | 已有 | [时分秒、步进/进位、12/24 小时、Tab 与本地化](../../ui/kit/time_field.go) | — |
| [TitleBar](https://gpui-kit.com/component/title-bar/) | 已有 | [自定义标题栏、窗口控制、macOS 双击偏好与失焦外观](../../ui/kit/title_bar.go) | 实现已有，原生行为验收仍待 F3 |
| [Toggle](https://gpui-kit.com/component/toggle/) | 已有 | [状态按钮、图标、尺寸、单选/多选组](../../ui/kit/toggle_group.go) | — |
| [Toolbar](https://gpui-kit.com/component/toolbar/) | 已有 | [左右区域、尺寸、工具分组、溢出与键盘](../../ui/kit/toolbar.go) | — |
| [Tooltip](https://gpui-kit.com/component/tooltip/) | 已有 | [通用提示、键盘焦点、延迟与取消](../../ui/kit/tooltip.go) | — |
| [Tree](https://gpui-kit.com/component/tree/) | 已有 | [虚拟化、展开、多选、单项禁用、键盘、动态数据与拖动](../../ui/kit/tree.go) | 拖动时自动滚动/展开未实现 |
| [VirtualList](https://gpui-kit.com/component/virtual-list/) | 已有 | [等高及可变高度实现、稳定 key、尺寸缓存、插入保持锚点](../../ui/kit/variable_list.go) | — |

验证记录见 [组件验收](component-acceptance-2026-10-02.md)。
