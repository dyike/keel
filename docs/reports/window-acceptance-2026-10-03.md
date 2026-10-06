# Window acceptance · 2026-10-03

English | [简体中文](window-acceptance-2026-10-03.zh-CN.md)

Native window life cycle and initial position are verified with desktop testing:

```sh
KEEL_DESKTOP=1 go test ./ui/window -run TestRealWindows -count=1 -v
```

2026-10-03 Actual operation on macOS passed: bringing multiple windows to front/closing, immediately closing and then reopening, and normal/borderless windows centered according to the available area of the screen. The position test also moves the window and requests subsequent frames to confirm it won't be re-centered. This result does not cover multi-monitor switching, Windows/Linux window managers, and system preference notifications.
