# One module for Keel, Gio and typesetting

English | [简体中文](single-module.zh-CN.md)

Status: implemented and locally validated; see the [validation record](desktop-cgo-free-validation.md).
Production release publication and desktop CGO migration remain pending.
The user's request to maintain one
`go.mod` supersedes the former separate-fork delivery design and the instruction
to preserve upstream module paths. Here, “typing” means `typesetting`
(`github.com/go-text/typesetting`). This step merges modules; desktop CGO removal
remains subsequent work described in [the desktop plan](desktop-cgo-free.md).

## Target structure

```text
github.com/dyike/keel/
├── go.mod                   # the only maintained module manifest
├── go.sum
├── ui/
├── native/
├── cmd/
├── gio/                     # ordinary packages, no go.mod/go.sum
├── typesetting/             # ordinary packages, no go.mod/go.sum
└── third_party/             # attribution and patch logs only
```

The libraries live in the root-level gio/ and typesetting/ directories. Preserve
licenses, attribution, patch history and upstream source style. Both libraries
ship inside Keel's module zip and share its version.
No nested releases, fork lock manifest, upstream replacement mode or application
replace is needed. No `go.work`, symlinks, replacement shim modules or duplicate
copies are introduced. External dependencies still retain their own modules;
one manifest refers to the source maintained in this repository.

| Former import prefix | Current import prefix |
| --- | --- |
| `gioui.org` (Gio root module packages) | `github.com/dyike/keel/gio` |
| `github.com/go-text/typesetting` | `github.com/dyike/keel/typesetting` |
| `github.com/dyike/keel/third_party/gio` (unpublished intermediate path) | `github.com/dyike/keel/gio` |
| `github.com/dyike/keel/third_party/typesetting` (unpublished intermediate path) | `github.com/dyike/keel/typesetting` |

`gioui.org/shader` is a separate external module and stays unchanged, including
`gioui.org/shader/gio`. Do not rewrite every `gioui.org/` string indiscriminately.
The currently used shader packages do not require a second Gio implementation;
`shader/piet` has separate `gioui.org/cpu` imports and must not be pulled in by
an overbroad rewrite. Verify package graphs for every supported target/tag.

## Compatibility and architectural boundaries

Changing package paths changes the identity of named Go types. Preserve
component names, behavior and method shapes, but do not promise full source
compatibility to applications directly importing upstream Gio/go-text. For
example `core.C`/`core.D` alias the newly located layout types; an application
with a `Layout(gioui.org/layout.Context)` method must migrate that import.
Publish the two-prefix migration table and compile an external custom widget
using the new Gio layout/unit/op packages. Third-party upstream Gio widgets
must migrate or be adapted explicitly; no implicit interchangeability or full
compatibility alias layer is part of this change. Old Keel releases stay intact.

Update `AGENTS.md`, architecture and dependency checks before source migration:

- `ui` may use the appropriate Gio/typesetting packages; `native` and the CLI
  must still work without importing the GUI, directly or transitively.
- Gio may depend on typesetting and its own internals. Typesetting must not
  depend on Gio. Neither fork may depend on Keel `ui`, `native`, `cmd` or its
  application-level helpers, despite now sharing a module prefix.
- Preserve ui/native separation, all existing layer restrictions and opt-in
  heavy dependencies. Teach `internal/deps` about the two foundation subtrees
  explicitly. Do not exempt foundation dependencies from validation.
- Go's `internal` visibility still follows directory ancestry. Keep each
  library's internal hierarchy, and rewrite its own imports consistently.

## Implementation sequence

1. **Capture the baseline.** Record the current working tree and build/test
   results, including known CGO/MCP failures. Preserve concurrent changes and
   the ABI probe in `gio/internal/nativeabi`. Source relocation must
   not absorb a renderer rewrite or claim to fix the existing CGO failure.
2. **Unify dependencies and imports in one coherent change.** Merge the dependency
   requirements from all three manifests into the root, retaining the currently
   selected versions initially (do not downgrade root x/image/x/sys/x/text).
   Rewrite Go imports in all source/test/generated files, examples and templates;
   update import-path comments and generation inputs. Remove both nested
   go.mod/go.sum files and the root require/replace entries for the two former
   modules. Run root tidy and review the resulting dependency changes. Preserve
   build tags, shader contents, native code and runtime behavior.
3. **Migrate path-bearing tooling.** Update `-X gioui.org/app.ID` in
   `cmd/keel/run.go`, tests and `scripts/build-ios-hello.sh`; audit other linker
   symbols, package discovery, permission detection, generators and doc examples.
   Audit strings and native/assembly symbols, not only import declarations.
4. **Simplify scaffold and delivery.** `keel new` writes one Keel requirement,
   then tidy; `new -replace <checkout>` writes only a Keel replacement. Offline
   generation with an exact version needs no fork metadata and does no network
   work; version queries such as latest require cached resolution or an actionable
   error. Keep transactional generation, failure rollback, path quoting and
   concurrent-file protection from the current work. Remove the unpublished
   backend-lock manifest, fork resolver, fork-specific `deps`/doctor behavior,
   `-backend` selection and their obsolete tests. Retain general doctor checks.
   Do not auto-edit existing applications' custom replacements; document removing
   obsolete ones after migrating their imports. No new dependency-manager CLI is
   needed. Historical separate-module releases use their matching historical CLI.
5. **Update packaging and CI.** Complete the packager adaptation below. Rewrite
   `scripts/check-backend-release.py` as a single-root-module archive fixture;
   root `./...` now includes both libraries, so update
   `scripts/check-desktop-cgo-free.sh`, workflows and tests accordingly. Targeted
   commands use `go test ./gio/... ./typesetting/...`
   from root. Convert the old upstream-replace-removal test into a check for no
   upstream core packages in the application graph. Document and release only
   the root Keel version, with no nested tags.

## Packaging is a required part of the path migration

The installed `gioui.org/cmd/gogio@v0.10.0` hardcodes more than app.ID:
`build_info.go` sets app/internal/log.appID and app.extraArgs;
`androidbuild.go` discovers `gioui.org/app/` and permission prefixes;
`iosbuild.go` discovers the original app path. Passing one additional linker
flag cannot repair those lookups.

Use a small maintained copy of the required gogio Go packager sources as ordinary
packages at `internal/gogio`, invoked in-process through a testable
entrypoint by the existing CLI. Retain its license and record its upstream
version/patches in bilingual documentation. Parameterize the Gio package root
and set it to the new path; use it for all link symbols, resource discovery and
permission imports. Keep it independent of UI imports. Merge necessary tool
dependencies into root go.mod. Do not add another module, patch the module cache,
introduce native sources or rewrite platform packaging algorithms. Replace
external gogio invocations in CLI/site scripts with the Keel packaging entrypoint;
cover macOS, Android, iOS and Web outputs and Windows resource behavior. Update
architecture/dependency tables for the helper before adding it.

This adapter is necessary even though mobile CGO removal is deferred: changing
Gio paths must not break existing mobile packaging. Keep the separate shader
module and iOS shader adapter until their own requirements change. Verify actual
app IDs, launch arguments, permissions and resources, not just successful linking.

## Acceptance and handoff

- Exactly one maintained go.mod/go.sum pair; `go list -json` on both relocated
  package trees reports the root Keel module in `Module.Path`. The ordinary app's module graph
  has no required/replaced `gioui.org` or `github.com/go-text/typesetting` root
  modules; external `gioui.org/shader` is allowed. Scan all platforms/tags for
  stale core imports, with explicit exceptions for historical/attribution text.
- Root formatting, vet, test and race checks include both trees. Windows
  amd64/arm64 no-cgo builds still pass; native-only cross-builds stay intact.
  macOS/Linux retain their current toolchain requirements until binding migration.
  Record pre-existing failures separately; new failures require fixes.
- A file-module-proxy fixture serves the real root archive, including both
  former submodules, with a fresh module cache and `GOWORK=off`. Generate an
  external app with only a Keel requirement, then tidy/build without any local
  replacements. Inspect each fork package's `Module.Path` and files. Also install
  `cmd/keel@<fixture-version>` to a temporary GOBIN to verify that removing the
  root replacements restores normal versioned CLI installation. Real remote
  installation is a release gate, not a claim that an unpublished version exists.
- Compile the custom-widget fixture; exercise default, offline and local scaffold
  modes (including spaces), preserve generator failure/rollback tests, and
  validate packaging discovery and new linker symbols on supported targets.
- Compare same-host hello/component screenshots and run real-window fixtures
  after migration. Package identity changes must not change rendering or lifecycle.
  No new benchmark suite or CGO migration is necessary to accept this structural
  change; existing performance patches and tests remain included.

For `impl`, finish this single-module change before resuming M0/M1/M2/L1/L2 of
the desktop plan. Replace the former D1 delivery work rather than implementing
both approaches. Keep its useful transactional scaffolding tests and the D0/ABI
work. The previous validation report records historical results, not acceptance
of this new architecture. Completion means a real external application receives
both libraries through the single Keel module, with preserved packaging behavior.
