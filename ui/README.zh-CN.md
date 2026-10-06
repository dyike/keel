# ui

[English](README.md) | 简体中文

界面模块集合。和 `native/` 一样，每个子目录是一个职责单一的模块。

| 模块 | 做什么 | 依赖 |
| --- | --- | --- |
| [core](core/README.zh-CN.md) | 地基：`Widget` 接口、回调、线程规则（`Update`） | 无 |
| [theme](theme/README.zh-CN.md) | 颜色、字号、字体 | 无 |
| [locale](locale/README.zh-CN.md) | 框架文字，中英文切换 | 无 |
| [base](base/README.zh-CN.md) | 组件行为，不含外观：键盘导航、首字母跳转、多选、打开状态 | 无 |
| [el](el/README.zh-CN.md) | GPUI 风格：视图 + 链式样式元素 + flexbox | core、theme、locale |
| [kit](kit/README.zh-CN.md) | 组件：按钮、表单、表格、浮层、应用外壳、图表 | base、el、core、theme、locale |
| [window](window/README.zh-CN.md) | 窗口：`Open`、`Main`、快捷键、截图 | core、theme |
| [markdown](markdown/README.zh-CN.md) | Markdown 渲染，针对 AI 流式输出优化 | el、core、theme、locale |

模块边界和依赖图见 [架构](../docs/architecture.zh-CN.md#模块)。应用的最小写法见 [快速开始](../docs/getting-started.zh-CN.md#写第一个窗口)。

`internal/` 保存帧锁、编辑状态、文字绘制、图片加载与无界面测试工具，外部应用不能直接引用。
