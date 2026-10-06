# 窗口验收 · 2026-10-03

[English](window-acceptance-2026-10-03.md) | 简体中文

原生窗口生命周期和初始位置用桌面测试验证：

```sh
KEEL_DESKTOP=1 go test ./ui/window -run TestRealWindows -count=1 -v
```

2026-10-03 在 macOS 实际运行通过：多窗口置前/关闭、立即关闭后重开，以及普通/无边框窗口按屏幕可用区域居中。位置测试还移动窗口并请求后续帧，确认不会被重新居中。此结果未覆盖多显示器切换、Windows/Linux 的窗口管理器及系统偏好通知。
