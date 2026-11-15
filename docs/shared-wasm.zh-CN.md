# 共享 Go WebAssembly 组件库

[English](shared-wasm.md) | 简体中文

## 默认构建与 CI（2026-11-15）

`gallerybuild` 现在默认执行细粒度共享运行时流水线（28 个叶子分片加 `shared_01`..`shared_13` 共享依赖分片，把 Markdown/Goldmark/HTML、CodeEditor、NumberInput、Table/CSV、AppShell、Command、ImageCache、Tree 等跨分片子系统也从主模块拆出）。将 `BINARYEN_DIR` 指向 Binaryen version_133 的 bin 目录，并使用全新的输出目录。`-mode independent` 保留仅依赖 Go 的构建，`-categories` 选择独立分类开发子集，`-analysis` 仍是内部分析入口。参见[构建说明](../internal/gallerybuild/README.zh-CN.md)。

Site 工作流同时覆盖 PR 和发布事件：固定 Go 1.27.4，并固定及校验 Binaryen 压缩包 SHA-256，无 cgo 构建，执行加载器/归属测试，再用 **Google Chrome** 验证全部 118 个组件首帧、冷分类、输入保留、完整回退、重试及历史导航。失败阻止发布并保留验收产物，PR 不部署。推送前这些只是本地 CI 配置，托管 CI 和部署实跑需要单独记录。

Node 可找到 Playwright 后可复现加载验收：

```sh
node scripts/verify-wasm-loading.cjs /tmp/site-new work/wasm-loading
```

脚本提供真实 gzip-9 HTTP 响应和带哈希资源的不可变缓存。共享 Input 与同等优化的完整 Input 交替各跑五次全新普通浏览器配置目录（默认磁盘缓存容量），固定 390×844、6 Mbit/s、40 ms 延迟及 2 倍 CPU 降速。记录应用就绪、分类加载完成后两帧、实际 WASM 传输/压缩/解压字节、缓存重载、首次 Dialog 切换及不重复下载的回访。门槛为浏览器错误为零，首次总传输量和 WASM 传输量**含首个子模块**均至少降低 5%，冷/热加载中位数不超过完整包的 1.15 倍加 100 ms，首次分类切换少于 3 秒。缓存命中和请求模块数量必须精确满足断言。报告保留浏览器版本、全部样本和截图，时序失败不会静默重跑。

这些是受控加载门槛，不代表手机真机、线上 CDN、持续帧率或整个进程/GPU 内存验收。报告的共享 WASM 内存长度仅是线性内存。本地缓存头不能证明 Pages 的缓存策略。若要求更新后旧页面仍能加载子模块，部署必须保留旧哈希资源；Pages 替换整个发布快照，本工作流不保留历史发布。仍保留完整包回退以便排障。

2026-11-15 本地 Chromium 159.0.8154.6 对细粒度叶子分片与共享依赖分片（共 41 个子模块：28 个叶子分片与 13 个共享依赖分片，从 `primary.wasm` 拆出 4,248 个函数）验收通过全部六项门槛。`GOOS=js` 排除了桌面专用的 `ui/window` 自动化套接字、热重载与尺寸持久化，将 `registerSection` 构建闭包提升为顶层具名函数 `section_<name>`，并将 `ui/highlight/golang` 词法注册改为惰性初始化。同等优化（`wasm-opt -Oz --strip-debug`）后，`baseline-opt.wasm` 为 **29,006,511 字节（gzip-9 为 7,632,520 字节）**，而 `primary.wasm` 降至 **22,803,133 字节（gzip-9 为 5,930,064 字节）**（相比基线减少 6,203,378 原始字节 / 1,702,456 gzip-9 字节，主模块压缩体积**降低 22.3%**；相比此前 12 分片方案的 6,980,693 gzip-9 字节再减少 **1,050,629 字节**）。叶子分片 gzip-9 大小为 **8,265–118,284 字节**（`inputs.wasm` 为 **24,924 字节**，`overlays.wasm` 为 **14,246 字节**），共享依赖分片 gzip-9 大小为 **9,897–296,485 字节**。五组配对样本的首次总传输量为 **7,653,684 对 9,331,465 字节**（减少 18.0%），含 Input 的 WASM 传输量为 **6,016,294 对 7,699,119 字节**（减少 21.9%），冷加载中位数为 **10.78 对 13.01 秒**（提速 17.1%）。缓存重载为 **381 对 379 ms**，首次 Dialog 切换为 **115 对 39 ms**。118 个首帧、28 个冷启动叶子分片、输入保留、回退、重试和历史导航也全部通过，浏览器错误为零。本次本地改动尚未执行托管 CI。

下文保留历史实验数据与限制，其中关于旧默认构建和 CI 的描述属于当时版本。

构建同时生成完整程序和独立分类程序。完整组件库现默认使用完整程序，保留跨分类状态；嵌入式组件预览仍选择分类程序。这以更大的首次下载换取状态连续性，不等于完成共享运行时拆分。本实验验证另一条路线：使用 [Binaryen wasm-split](https://github.com/WebAssembly/binaryen/tree/version_133/src/tools/wasm-split) 拆分已经链接好的 Go/Gio 程序，保留同一份 Go 运行时、内存、可变全局量和分发表。这不是新的 Go 动态链接 ABI，也没有成为生产默认方式。

## 复现

使用 Binaryen version_133 的原生工具（仅构建阶段使用，应用不引入 cgo）：

```sh
python3 scripts/probe-shared-wasm.py --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-shared --optimize
# 中文需要时，把站点 font.otf 复制到 /tmp/keel-shared。
python3 -m http.server 8767 --bind 127.0.0.1 --directory /tmp/keel-shared
# 另一个终端，确保 Node 能找到 Playwright：
node scripts/verify-shared-gallery.cjs http://127.0.0.1:8767/ work/gallery-shared
```

浏览器验证通过触控聚焦和实际输入事件录入文本，加载 Dialog 子模块并打开、关闭对话框，然后返回 Input，使用选中和复制操作从剪贴板确认原文本仍在。验证同时检查内存对象未更换、子模块只在显式请求时下载、重复加载不会再次请求。报告包含浏览器版本、网络请求和截图。使用编译器和标准库版本一致的 Go 工具链即可，无需覆盖编译器版本。

## 已观察到的范围与限制

本地 Chromium 159 已通过单个子模块包含 26 个 Dialog／组件库函数的流程。子模块从主实例导入内存、表、可变全局量和函数，没有创建第二份 Go 运行时。实验宿主现已安装可选的组件构建门控：常规导航和深链接会等待子模块，显示加载／错误页面并提供重试。独立分类包及普通完整包不安装此钩子。迟到的完成事件只重绘当前匹配页面，不会改变导航目标；重复加载合并为同一请求，失败后可重试而不重启 Go。原先只覆盖 Dialog 分区；下文的扩展实验仍不证明完整分类依赖归属已经解决。

这个小分区证明路线可行，不证明下载体积改善。测得未优化原模块为 32,119,816 字节（gzip-9 为 7,808,048），重写后的主模块为 32,143,687（gzip-9 为 8,389,066），还要加上子模块（gzip-9 为 15,231）。Binaryen 重写会显著改变压缩结果，不能凭这些结果替换现有分类构建。

生产门槛仍包括：每个分类及共享依赖的确定性函数归属；启动、GC、异步事件及嵌套组件不调用未加载函数；带进度、错误与重试的异步导航；加载中的深链接及前进后退；跨分类数据／状态保留；缓存与版本一致性；首传体积实测下降，首次调用、帧时间和内存成本可接受；重复浏览器与真机验收。在这些条件满足前保留独立包／完整包回退，绝不能混用不同构建的主模块与子模块。

可复现的 `--optimize` 选项对未拆分基线、主模块和子模块统一使用 `wasm-opt -Oz --strip-debug`，并把原始／gzip-9 大小写入 `sizes.json`。新导航门控构建中，基线为 29,157,010 字节／gzip 7,668,807，主模块为 29,114,017／gzip 7,652,877，子模块为 46,478／gzip 13,033。首传仅减少 15,930 个压缩字节，加载 Dialog 后总量仅减少 2,897 字节，不能据此认定完整分类拆分有实用收益。启动、首次调用、帧时间及物理内存仍未完成对比测量。

原型现为主模块、子模块、加载器和启动脚本生成内容哈希文件名，并在 `shared-build.json` 中记录配对。部署新索引时应保留旧索引引用的资源；哈希文件名不能证明生产缓存响应头或原子部署正确。Chromium 159 测试暂停子模块请求，返回 Input，再释放请求，通过复制检查输入保留且路由未改变，之后前进可打开已加载的 Dialog。同一轮还验证了请求中断后的触控重试。这些报告是本地验证结果，不代表已部署或通过实体手机验收。

## 多个按需分区

增加 `--partitions dialog,nav_stack,emoji_picker,terminal` 参数，通过 Binaryen 的 `--multi-split` 构建四个子模块。清单记录显式的函数归属，各子模块导入同一个主实例。浏览器验证在重试、历史记录之外，检查各分片独立按需下载、内存对象不变，以及访问四个分区后复制出的输入仍保留。

Dialog 的共享 kit 方法保留在主模块：Motion 和 Image 即使没有访问 Dialog 示例也会使用它们。原先的 26 函数原型错误地把这些共享方法分给 Dialog；新分区只包含 Dialog 示例工厂和方法。回归流程增加了下载 Dialog 之前访问 Motion。另外三个分区目前包含各自视图的方法，其支持库仍驻留主模块。这些显式选择仍是实验，不是通用依赖归属算法，也不代表生产体积已经改善。

保留共享 Dialog 方法后，Go 1.27.4 四分区构建测得 gzip-9：基线 **7,812,742 字节**、主模块 **7,780,262**，Dialog / NavStack / EmojiPicker / Terminal 子模块分别 **5,833 / 5,722 / 15,148 / 13,804** 字节。首传只减少 **32,480 字节**，全部加载比统一优化的基线多 **8,027 字节**。Chromium 159 已通过四分片按需导航、剪贴板核对输入保留、内存对象一致、Dialog 下载前访问 Motion、失败重试及等待下载时前进后退。这是本地浏览器证据，不代表生产或真机门槛通过。

## 分类归属与站点接入

分类方案对比完整程序、十二个分类和无示例构造函数基线实际被链接保留的函数名。`gallerybuild -analysis` 保留符号，并让所有构建使用一致的可选导入；拒绝只分析部分分类。基线函数、共享函数、初始化闭包及无法归属的函数保留在主模块，只有恰好属于一个分类的函数进入对应子模块。分析记录主程序 SHA-256，拆分时拒绝不同的主程序。这是保守的链接证据，不是完整的动态调用证明。

```sh
go run ./internal/gallerybuild -analysis -out /tmp/keel-ownership
python3 scripts/analyze-gallery-wasm.py --input /tmp/keel-ownership --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-ownership/ownership.json
python3 scripts/probe-shared-wasm.py --ownership-build /tmp/keel-ownership --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-shared-site-input --optimize
# 与普通站点流程一样，在此目录提供 font.otf。
go run ./internal/site -demo /tmp/keel-shared-site-input -out /tmp/keel-shared-site
python3 -m http.server 8774 --bind 127.0.0.1 --directory /tmp/keel-shared-site
# 另一个终端，确保 Node 能找到 Playwright：
node scripts/verify-shared-categories.cjs http://127.0.0.1:8774/demo/ work/gallery-shared-categories
node scripts/verify-shared-gallery.cjs http://127.0.0.1:8774/demo/ work/gallery-shared-retry
```

当前 118 分区构建把 1,469 个函数移入十二个子模块。统一优化后 gzip-9：未拆分基线 **7,812,742 字节**，主模块 **6,972,526 字节**，未加载分类前首传减少 **840,216 字节（10.75%）**。子模块为 **15,508–196,317 字节**；十二个全部加载后，总量比基线多 **44,645 字节**。这些是压缩产物大小，不是实际网络时间或已部署的压缩响应证据。前面的四分片测量保留为历史结果。

站点发布器支持 `shared-build.json`，只发布配对的运行时、子模块、脚本及字体，不发布分析中间产物。进度加载器只请求一次主模块，并在 Go 启动前加载共享宿主。`?bundle=all` 选择单独配对的完整程序，不安装懒加载钩子。子模块缺失或同时存在独立包／共享包清单会导致构建失败。所有引用资源都有内容哈希；部署仍需自行保留旧索引引用的资源。

本地 Chromium 159 已通过 118 个示例首帧、十二次分类冷启动（每次仅请求主模块和本分类子模块）、内存对象一致、跨分类复制确认输入保留，以及完整包回退后的实际输入。没有穷尽每个组件的交互路径。CI 仍使用原构建流程，等待启动、帧时间、物理内存及真机门槛；站点接入已可显式使用，本次未部署。

浏览器验证脚本默认使用本机安装的 Google Chrome；显式设置 `KEEL_BROWSER_CHANNEL=chromium` 可使用 Playwright Chromium。分类嵌入检查暴露了 Go 1.27.4 JS/WASM 的越界：无焦点输入队列在大型处理器 map 中查询 nil 键。独立的标准库程序和实际输入回归测试在修复前失败；`InputHint` 在无焦点时提前返回后测试通过。这是后续 `call to released function` 控制台提示及 Chromium 159 headless shell 崩溃背后的原始异常，见 [Gio 补丁记录](../third_party/README.zh-CN.md)。尝试修改分类布局未修复原因，已撤销。

加入输入修复后的新构建在 Chrome 157 中通过 118 个示例首帧、十二次分类冷启动、单组件／分类嵌入、完整程序回退、精确输入保留，页面及控制台错误均为零。另一轮通过子模块失败重试和加载中前进后退。Chromium 159 headless shell 也能启动此前对照的三个嵌入入口，没有错误或崩溃。这是浏览器回归证据，不是真机或性能验收。重建后 gzip-9 基线为 7,812,756 字节、主模块 6,972,851，主模块首传减少 839,905 字节；全部子模块加载后比未拆分基线多 44,916 字节。
