# ResizableGroup

[English](resizable_group.md) | 简体中文

独立多面板分割容器，支持横纵排列和嵌套。与双面板 Resizable 并存。

```go
g := kit.ResizableGroup(
    kit.ResizablePanel{ID: "files", Content: files, Size: 180, Min: 100, Max: 300},
    kit.ResizablePanel{ID: "editor", Content: editor, Min: 160},
    kit.ResizablePanel{ID: "preview", Content: preview, Size: 240, Min: 100},
)
```

默认横排；Vertical 改为纵排。ID 必须非空且唯一，空 ID 和重复 ID 后续项忽略。Size 为初始 dp，0 初始均分指定尺寸面板之外的剩余空间；Min 默认 0，Max 为 0 表示无上限，上限小于 Min 时以 Min 为准。放入有确定大小的父容器。

拖动只改变相邻可见两面板，其他面板不变。方向键每次 16dp，Home/End 到相邻范围边界；取消保留最后有效尺寸。SetDisabled 同时禁用把手和内容，HandleAppearance 使用与 Resizable 相同的把手配置。

窗口变化从最后一个可见面板向前分配余量，保留前面板尺寸；全部达到上限时尾部留空。空间不足以容纳最小尺寸时按各 Min 比例压缩，仍保留每条 6dp 把手。首次测量后请求下一帧收敛。

Sizes 返回按 ID 索引的尺寸副本。SetSizes 设置已知 ID 的有限非负尺寸，不触发回调，下一帧按容器范围收敛；显式 0 不再作为自动均分。OnChange 仅在用户实际调整后返回完整尺寸副本。

SetVisible(id, visible) 控制单面板显隐，隐藏内容继续声明并保留状态及尺寸；剩余面板重新分配空间，单侧仍遵守自身 Max。SetPanels 替换/重排配置，保留相同 ID 的尺寸，新 ID 使用 Size；更改已有 ID 的大小请用 SetSizes。被移除的面板及把手状态清理。显隐、配置和窗口变化不触发 OnChange。
