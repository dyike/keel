# Keel、Gio 与 typesetting 统一模块

[English](single-module.md) | 简体中文

状态：已实现并通过本地验收，见[验证记录](desktop-cgo-free-validation.zh-CN.md)。生产发布与桌面 CGO 迁移仍待完成。用户要求统一维护一份 `go.mod`，本设计取代此前独立 Fork 分发方案以及保留上游模块路径的约束。“typing” 按仓库实际指 `typesetting`（`github.com/go-text/typesetting`）。当前步骤合并模块，桌面 CGO 移除仍是[桌面计划](desktop-cgo-free.zh-CN.md)中的后续工作。

## 目标结构

```text
github.com/dyike/keel/
├── go.mod                   # 唯一维护的模块清单
├── go.sum
├── ui/
├── native/
├── cmd/
├── gio/                     # 普通包，不再有 go.mod/go.sum
├── typesetting/             # 普通包，不再有 go.mod/go.sum
└── third_party/             # 仅保存来源与补丁记录
```

两个库位于根级 gio/ 和 typesetting/ 目录，保留许可证、来源说明、补丁历史和上游代码风格。两个库进入 Keel 根模块 zip，随 Keel 统一版本发布。不再需要子模块发布、Fork 锁定清单、上游替换模式或应用端 replace。不引入 go.work、符号链接、替换用的桥接模块或重复源码。外部依赖仍拥有各自模块；一份清单指本仓库维护的源码。

| 历史 import 前缀 | 当前 import 前缀 |
| --- | --- |
| `gioui.org`（Gio 根模块的包） | `github.com/dyike/keel/gio` |
| `github.com/go-text/typesetting` | `github.com/dyike/keel/typesetting` |
| `github.com/dyike/keel/third_party/gio`（未发布的中间路径） | `github.com/dyike/keel/gio` |
| `github.com/dyike/keel/third_party/typesetting`（未发布的中间路径） | `github.com/dyike/keel/typesetting` |

`gioui.org/shader` 是独立外部模块，连同 `gioui.org/shader/gio` 保持不变，不能无差别替换所有 `gioui.org/` 字符串。目前使用的 shader 包不要求第二套 Gio；`shader/piet` 另有 `gioui.org/cpu` 导入，不能因过度替换将其引入。按各支持平台/tag 验证包依赖图。

## 兼容性与架构边界

包路径变化会改变 Go 命名类型的身份。保持组件名称、行为与方法形态，但不能向直接导入上游 Gio/go-text 的应用承诺完全源码兼容。例如 `core.C`/`core.D` 将成为新路径 layout 类型的别名；应用若实现 `Layout(gioui.org/layout.Context)`，必须迁移 import。发布双前缀迁移表，增加使用新 Gio layout/unit/op 的外部自定义组件编译夹具。第三方上游 Gio 组件需要迁移或显式适配；本次不提供隐式类型互通或完整别名兼容层。旧 Keel 发布不变。

源码迁移前同步 `AGENTS.md`、架构和依赖检查：

- ui 按现有职责使用 Gio/typesetting；native 和 CLI 仍须独立于 GUI，不能直接或间接导入它。
- Gio 可依赖 typesetting 和自身 internal；typesetting 不能反向依赖 Gio。两个库都不能依赖 Keel 的 ui、native、cmd 或应用层辅助包，即使现已共用模块前缀。
- 保持 ui/native 隔离、原有分层规则和重依赖可选加载。`internal/deps` 显式识别两个基础库子树，不能笼统豁免基础库依赖。
- Go 的 internal 可见性仍按目录祖先限定；保留各库 internal 层级，一致修改库内导入。

## 实施顺序

1. **记录基线。** 保存当前工作区与构建/测试结果，包括已知 CGO/MCP 失败。保留并发修改及 `gio/internal/nativeabi` 探针。路径迁移不夹带渲染器重写，也不能声称解决了已有 CGO 问题。
2. **一次完整合并依赖与路径。** 合并三份清单的依赖到根目录，初始保留当前已选择版本，不降级根 x/image、x/sys、x/text。改写源码、测试、生成文件、示例和模板的 import，更新路径注释和生成输入。移除两个子目录 go.mod/go.sum 以及根 require/replace 中两个旧模块，执行根 tidy 并审查依赖变化。保持 build tag、shader 内容、原生代码和运行行为。
3. **迁移工具中的包路径。** 更新 `cmd/keel/run.go`、测试、`scripts/build-ios-hello.sh` 中的 `-X gioui.org/app.ID`；检查其他链接符号、包发现、权限识别、生成器、文档示例。不能只改 import，还要扫描字符串和原生/汇编符号。
4. **简化脚手架与交付。** `keel new` 仅写入 Keel require 后 tidy；`new -replace <checkout>` 仅写 Keel replace。离线生成指定精确版本不需要 Fork 元数据且不访问网络；latest 等版本查询需要缓存解析，否则给明确错误。保留当前工作的事务生成、失败回滚、路径引用和并发文件保护。撤下未发布的 backend-lock 清单、Fork 解析器、Fork 专用 deps/doctor、`-backend` 选择及过时测试，保留通用 doctor。已有应用自定义 replace 不自动修改，文档说明迁移 import 后清理过时条目。不需要新增依赖管理命令。历史独立模块发布使用对应历史 CLI。
5. **更新打包与 CI。** 完成下节打包适配；将 `scripts/check-backend-release.py` 改为单根模块归档夹具。根 `./...` 已覆盖两个库，更新 `scripts/check-desktop-cgo-free.sh`、workflows 和测试；定向命令从根执行 `go test ./gio/... ./typesetting/...`。旧“去掉 replace 验证上游”改为检查应用依赖图无上游核心包。只发布根 Keel 版本，不发子目录标签。

## 打包是路径迁移的必要部分

当前缓存中的 `gioui.org/cmd/gogio@v0.10.0` 不只硬编码 app.ID：`build_info.go` 还设置 app/internal/log.appID、app.extraArgs；`androidbuild.go` 查找 `gioui.org/app/` 和权限前缀；`iosbuild.go` 也按原路径定位 app。只额外传一个链接参数不能修复这些发现逻辑。

将所需 gogio Go 打包器源码以普通包形式维护在 `cmd/keel/internal/gogio`，由现有 CLI 通过可测试入口在进程内调用。保留许可证，在双语文档记录上游版本与补丁。参数化 Gio 包根路径并设置为新路径，统一用于链接符号、资源定位和权限导入识别；打包器自身不导入 UI。必要工具依赖并入根 go.mod，不引入新模块、不改模块缓存、不增加原生源码、不重写平台打包算法。CLI 和站点脚本的外部 gogio 调用改为 Keel 打包入口，覆盖 macOS、Android、iOS、Web 产物及 Windows 资源行为。新增辅助包前更新架构与依赖表。

即使移动端 CGO 移除延期，此适配仍必需：Gio 路径变化不能破坏现有移动端打包。保留独立 shader 模块和 iOS shader 适配，直到其需求另行变化。验收实际 app ID、启动参数、权限和资源，不能只检查链接成功。

## 验收与交接

- 本仓库只有一对维护中的 go.mod/go.sum；两个迁移后包树的 `go list -json` 中 `Module.Path` 归属根 Keel 模块。普通应用模块图不 require/replace `gioui.org`、`github.com/go-text/typesetting` 两个根模块，允许外部 `gioui.org/shader`。扫描所有平台/tag 的旧核心导入，历史/来源说明有明确例外。
- 根格式、vet、test、race 覆盖两个库；Windows amd64/arm64 无 CGO 构建及 native-only 交叉构建保持通过。macOS/Linux 在绑定迁移前保留当前工具链要求。已有失败单列，新失败必须修复。
- 文件模块代理夹具提供真实根模块归档（包含两个原子模块），使用全新模块缓存和 `GOWORK=off`。外部应用仅 require Keel，无任何本地 replace 即可 tidy/build；核对两个库包的 `Module.Path` 和文件。另将 `cmd/keel@<fixture-version>` 安装到临时 GOBIN，验证去掉根 replace 后恢复正常的指定版本 CLI 安装。真实远程安装属于发布门禁，不能宣称未发布版本已存在。
- 编译外部自定义组件夹具；覆盖默认、离线、本地脚手架模式及含空格路径，保留生成失败/回滚测试，验证各目标打包发现逻辑与新链接符号。
- 对比同机 hello/组件截图，迁移后运行真实窗口夹具。包身份改变不应改变渲染和生命周期。本结构改动无需新增 benchmark 体系或完成 CGO 迁移才能验收；保留已有性能补丁和测试。

`impl` 先完成此次单模块合并，再继续桌面计划 M0/M1/M2/L1/L2。替换原 D1 分发工作，不同时维护两套方案；保留其中有价值的事务生成测试和 D0/ABI 工作。旧验证记录仅描述历史结果，不能作为新架构已通过的证据。完成标准是外部应用通过一个 Keel 模块得到两个库，并保留现有打包行为。
