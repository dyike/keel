# animation

[English](README.md) | 简体中文

`ui/el` 与 `ui/motion` 共用的帧时间数值采样器，只依赖标准库，不持有 UI 状态、不调度帧、不创建 goroutine 或原生计时器。调用方持有 Value，仅在 Sample 报告活动时请求后续帧。
