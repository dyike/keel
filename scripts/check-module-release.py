#!/usr/bin/env python3
"""Exercise release-backed scaffolding with the real module in a local Go proxy.

Synthetic fixture versions are never published. This checks that an app
requiring only Keel gets Keel's Gio and go-text, a Windows cross-build and go
install of the CLI, not production remote releases or native
    GPU/IME parity. Other dependencies use the configured Go module proxy.
"""

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import zipfile


ROOT_MODULE = "github.com/dyike/keel"
ROOT_VERSION = "v0.0.0-fixture"
FOUNDATIONS = (ROOT_MODULE + "/gio", ROOT_MODULE + "/typesetting")
CUSTOM_WIDGET = '''package main

import (
    "image"
    "image/color"

    "github.com/dyike/keel/gio/layout"
    "github.com/dyike/keel/gio/op/clip"
    "github.com/dyike/keel/gio/op/paint"
    "github.com/dyike/keel/gio/unit"
    "github.com/dyike/keel/ui/core"
)

type releaseWidget struct{}

var _ core.Widget = releaseWidget{}

func (releaseWidget) Layout(context layout.Context) layout.Dimensions {
    size := image.Pt(context.Dp(unit.Dp(24)), context.Dp(unit.Dp(24)))
    paint.FillShape(context.Ops, color.NRGBA{A: 255}, clip.Rect{Max: size}.Op())
    return layout.Dimensions{Size: size}
}
'''
IDENTITY_PROBE = '''package main

import (
    "encoding/json"
    "os"

    gioapp "github.com/dyike/keel/gio/app"
)

func init() {
    if os.Getenv("KEEL_RELEASE_IDENTITY") != "1" {
        return
    }
    json.NewEncoder(os.Stdout).Encode(struct {
        ID string
        Args []string
    }{gioapp.ID, os.Args[1:]})
    os.Exit(0)
}
'''
# gioui.org/shader stays an upstream module; Gio itself and go-text must not.
UPSTREAM = ("gioui.org", "github.com/go-text/typesetting")


def run(args, cwd, env, capture=False):
    print("+ " + " ".join(str(arg) for arg in args), flush=True)
    result = subprocess.run(
        [str(arg) for arg in args], cwd=cwd, env=env,
        stdout=subprocess.PIPE if capture else None,
        text=True, check=True,
    )
    return result.stdout if capture else None


def nested_modules(source):
    """Go module archives omit submodules, including test/example modules."""
    found = []
    for directory, dirs, files in os.walk(source):
        dirs[:] = [name for name in dirs if name not in (".git", "vendor")]
        current = Path(directory)
        if current != source and "go.mod" in files:
            found.append(current.relative_to(source))
            dirs[:] = []
    return found


def archive_files(source, candidates):
    omitted = nested_modules(source)
    for relative in sorted(set(candidates)):
        path = source / relative
        if not path.is_file() or path.is_symlink():
            continue
        if ".git" in relative.parts or "vendor" in relative.parts:
            continue
        if any(relative.is_relative_to(module) for module in omitted):
            continue
        yield relative, path.read_bytes()


def write_release(proxy, module, version, files):
    directory = proxy / module / "@v"
    directory.mkdir(parents=True, exist_ok=True)
    prefix = module + "@" + version + "/"
    go_mod = None
    count = 0
    with zipfile.ZipFile(directory / (version + ".zip"), "w", zipfile.ZIP_DEFLATED) as archive:
        for relative, data in files:
            name = relative.as_posix()
            if name == "go.mod":
                go_mod = data
            archive.writestr(prefix + name, data)
            count += 1
    if go_mod is None:
        raise RuntimeError("source archive has no go.mod: " + module)
    (directory / (version + ".mod")).write_bytes(go_mod)
    (directory / (version + ".info")).write_text(json.dumps({
        "Version": version, "Time": "2026-10-12T00:27:05Z",
    }) + "\n", encoding="utf-8")
    (directory / "list").write_text(version + "\n", encoding="utf-8")
    print("Archived %s@%s (%d real source files)" % (module, version, count), flush=True)


def json_objects(text):
    decoder = json.JSONDecoder()
    while text.strip():
        text = text.lstrip()
        value, end = decoder.raw_decode(text)
        yield value
        text = text[end:]


def check_release(root, temporary):
    env = os.environ.copy()
    env.update({"GOWORK": "off", "GOFLAGS": "", "CGO_ENABLED": "0"})
    # Build the host CLI before selecting the isolated fixture module cache.
    env.pop("GOOS", None)
    env.pop("GOARCH", None)
    binary = temporary / ("keel.exe" if os.name == "nt" else "keel")
    run(["go", "build", "-o", binary, "./cmd/keel"], root, env)

    external_proxy = run(["go", "env", "GOPROXY"], root, env, capture=True).strip()
    proxy = temporary / "proxy"
    proxy.mkdir()
    tracked = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"],
        cwd=root,
    )
    candidates = [Path(os.fsdecode(name)) for name in tracked.split(b"\0") if name]
    write_release(proxy, ROOT_MODULE, ROOT_VERSION, archive_files(root, candidates))

    cache = temporary / "module cache"
    cache.mkdir()
    env.update({
        "GOMODCACHE": str(cache), "GOWORK": "off", "GOFLAGS": "",
        "GONOPROXY": "none", "GOPRIVATE": "",
        "GOPROXY": proxy.as_uri() + "," + external_proxy,
        # Only this synthetic local fixture disables checksum verification.
        "GOSUMDB": "off",
    })
    offline = temporary / "offline app"
    offline_env = dict(env, GOPROXY="off", GOSUMDB="off", GONOPROXY="none", GOTOOLCHAIN="local")
    run([binary, "new", offline, "-keel", ROOT_VERSION, "-offline"], temporary, offline_env)
    if (offline / "go.sum").exists() or "replace" in (offline / "go.mod").read_text():
        raise RuntimeError("offline generation fetched dependencies or added replacements")

    app = temporary / "external app"
    run([binary, "new", app, "-keel", ROOT_VERSION], temporary, env)
    go_mod = (app / "go.mod").read_text(encoding="utf-8")
    if "replace" in go_mod:
        raise RuntimeError("generated go.mod has a replace directive:\n" + go_mod)
    output = run(["go", "list", "-m", "-json", "all"], app, env, capture=True)
    modules = {module["Path"]: module for module in json_objects(output)}
    selected = modules.get(ROOT_MODULE, {})
    if selected.get("Version") != ROOT_VERSION or "Replace" in selected:
        raise RuntimeError("unexpected root resolution: " + json.dumps(selected))
    for path in modules:
        if path in UPSTREAM:
            raise RuntimeError("external app selects upstream " + path)
        if path.startswith(ROOT_MODULE + "/third_party/"):
            raise RuntimeError("external app selects an obsolete foundation module " + path)
    module_dir = Path(selected["Dir"])
    if list((module_dir / "third_party").rglob("*.go")):
        raise RuntimeError("release archive contains obsolete foundation sources")
    for prefix in FOUNDATIONS:
        relative = prefix.removeprefix(ROOT_MODULE + "/")
        if not list((module_dir / relative).rglob("*.go")):
            raise RuntimeError("release archive omits foundation sources: " + prefix)
    if list(module_dir.rglob("go.mod")) != [module_dir / "go.mod"]:
        raise RuntimeError("release archive contains nested modules")
    output = run(["go", "list", "-deps", "-f", "{{.ImportPath}} {{with .Module}}{{.Path}}{{end}}", "."],
                 app, dict(env, GOOS="windows", GOARCH="amd64"), capture=True)
    found = set()
    for line in output.splitlines():
        package, _, module = line.partition(" ")
        for prefix in FOUNDATIONS:
            if package == prefix or package.startswith(prefix + "/"):
                if module != ROOT_MODULE:
                    raise RuntimeError("%s comes from module %s" % (package, module))
                found.add(prefix)
        if package.startswith("gioui.org/shader"):
            continue
        if any(package == path or package.startswith(path + "/") for path in UPSTREAM):
            raise RuntimeError("external app imports upstream " + package)
    if found != set(FOUNDATIONS):
        raise RuntimeError("external app does not use Keel's copies: " + ", ".join(sorted(found)))
    print("Resolved Gio and go-text from %s@%s" % (ROOT_MODULE, ROOT_VERSION), flush=True)

    (app / "custom_widget.go").write_text(CUSTOM_WIDGET, encoding="utf-8")
    build_env = dict(env, CGO_ENABLED="0", GOOS="windows", GOARCH="amd64")
    run(["go", "build", "-o", temporary / "fixture-app.exe", "."], app, build_env)
    gobin = temporary / "bin"
    run(["go", "install", ROOT_MODULE + "/cmd/keel@" + ROOT_VERSION], temporary,
        dict(env, GOBIN=str(gobin)))
    installed = gobin / binary.name
    default_app = temporary / "default app"
    run([installed, "new", default_app], temporary, env)
    if ROOT_VERSION not in (default_app / "go.mod").read_text():
        raise RuntimeError("installed CLI did not default to its own release")
    local_app = temporary / "local app"
    run([installed, "new", local_app, "-replace", module_dir], temporary, offline_env)
    local_mod = json.loads(run(["go", "mod", "edit", "-json"], local_app, env, capture=True))
    replacements = local_mod.get("Replace", [])
    if len(replacements) != 1 or replacements[0]["Old"]["Path"] != ROOT_MODULE:
        raise RuntimeError("local generation did not use exactly one Keel replacement")
    run([installed, "build", "-target", "windows", "-arch", "amd64,arm64"], app, env)
    run([installed, "build", "-target", "js"], app, env)
    for name in ("main.wasm", "wasm.js", "index.html"):
        output = app / "dist" / "web" / name
        if not output.is_file() or output.stat().st_size == 0:
            raise RuntimeError("web package missing " + name)
    if sys.platform == "darwin":
        (app / "identity_probe.go").write_text(IDENTITY_PROBE, encoding="utf-8")
        native_env = dict(env, CGO_ENABLED="1")
        architecture = run(["go", "env", "GOARCH"], app, native_env, capture=True).strip()
        bundle = temporary / "Identity Probe.app"
        run(["go", "run", "-mod=mod", ROOT_MODULE + "/cmd/gogio",
             "-target", "macos", "-arch", architecture, "-appid", "dev.keel.releaseprobe",
             "-o", bundle, ".", "fixture-launch-argument"], app, native_env)
        executable = bundle / "Contents" / "MacOS" / "Identity Probe"
        identity = json.loads(run([executable, "native-argument"], app,
                                  dict(native_env, KEEL_RELEASE_IDENTITY="1"), capture=True))
        if identity != {"ID": "dev.keel.releaseprobe", "Args": ["native-argument", "fixture-launch-argument"]}:
            raise RuntimeError("packaged app lost identity or arguments: " + json.dumps(identity))
        print("PASS: native macOS packaged application ID and launch arguments (existing cgo backend).")
    print("PASS: single-module release, custom widget, CLI installation, default/offline/local scaffolds and Windows/Web packaging.")
    print("NOT ACCEPTED: production releases, native GPU/IME parity or no-cgo macOS/Linux.")


def remove_readonly(function, path, exception):
    # Go's extracted module cache directories are read-only.
    os.chmod(path, 0o700)
    parent = Path(path).parent
    if parent.exists():
        os.chmod(parent, 0o700)
    function(path)


def main():
    root = Path(__file__).resolve().parent.parent
    temporary = Path(tempfile.mkdtemp(prefix="keel-module-release-"))
    try:
        check_release(root, temporary)
    finally:
        import shutil
        shutil.rmtree(temporary, onerror=remove_readonly)


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
        print("FAIL: " + str(error), file=sys.stderr)
        sys.exit(1)
