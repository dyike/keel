# EmojiPicker

[English](emoji_picker.md) | 简体中文

带分类、肤色、最近使用和虚拟网格的表情面板。EmojiPicker() 使用 Unicode 17.0：全部 3,944 个 fully-qualified 序列作为条目或肤色变体提供，混合肤色序列保留为独立条目。目录首次使用时解析，导入包不会提前解析；运行时搜索不访问网络。

消息输入场景使用 EmojiPicker().OnSelect(func(EmojiSelection))：每次选择都会通知，包括连续选择相同表情。EmojiSelection 包含精确输出 Emoji、目录基础表情 Base 和 Name。OnChange(func(string)) 只在 Value 变化时通知。SetValue 静默设置已知序列，忽略未知值，空字符串清空。SetDisabled 禁用交互并关闭弹层；程序赋值仍可用。

Popup(true) 显示触发按钮。低于按文字比例调整的 600dp 视口使用底部面板，更宽视口使用锚定弹出层。底部面板占可用宽度，分类栏位于网格下方。选择表情、Escape 或点击外部关闭。SetOpen 静默设置，Open 返回状态。默认内联显示。Name 设置面板／按钮名称。Height 设置网格视口高度，默认 264dp；外围面板负责容纳其他控件。

SetQuery 和 Query 控制搜索。搜索按空白拆分关键词，对名称和 Keywords 忽略大小写匹配，所有词都必须匹配；也支持精确表情及已登记的肤色序列，精确变体搜索会选择该变体。默认 Unicode 名称为英文；分类及组件界面跟随 locale，应用可以提供本地化表情名称和 Keywords。

SetSkinTone 接受 0（默认）或 1–5（由浅至深）。只输出显式登记的变体，无对应肤色时使用基础表情。改变偏好肤色不改变已选值。Groups 返回独立分类列表；SetGroup 选择分类，空字符串表示全部。ShowRecent 切换最近使用。Recent 返回最多 18 个去重基础表情的副本，最新在前。SetRecent 将已知变体归一到基础表情并过滤未知项。用户选择自动更新最近使用；跨会话持久化由应用负责。最近使用应用当前偏好肤色。

EmojiPicker(entries...) 和 Data([]EmojiEntry) 支持自定义目录。Data 复制条目和关键词切片，忽略空／重复输出，静默移除不再支持的已选值和最近使用。Data(nil) 显式清空目录。EmojiEntry 包含 Emoji、Name、Group、Keywords 和五个 SkinTones。DefaultEmojis 返回独立目录副本供本地化。scripts/generate-emoji.py 支持通过 --input 使用缓存源文件并校验固定哈希，--check 检查生成结果一致性。来源与许可证记录在 third_party/unicode 下。

每个网格单元提供完整名称、精确值及选中／禁用语义。方向键移动，Home/End 到首／末项，Enter/Space 选择；Tab 进入当前活动单元。搜索框仅通过常规焦点／输入操作获取焦点。单元至少 44dp，随文字比例增大，列数根据可用宽度调整。分类和肤色栏横向滚动；底部预览悬停／聚焦的名称。网格只构建可见行及附近预绘制行。

使用系统 emoji 字体，不打包表情图片。各系统字体覆盖和颜色字形支持决定实际外观，新表情缺少字体时可能显示缺字。真机／输入法和跨平台字体一致性需要单独验收。

验证：go test ./ui/kit ./ui/window -run TestEmojiPicker；go run ./examples/components -section emoji_picker。
