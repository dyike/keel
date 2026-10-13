#!/usr/bin/env bash
# Strict compile-only gate. Runtime and packaging acceptance are separate.
set -euo pipefail
root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
export CGO_ENABLED=0 GOWORK=off GOFLAGS=
export GOOS="${GOOS:-$(go env GOOS)}"
export GOARCH="${GOARCH:-$(go env GOARCH)}"
case "$GOOS/$GOARCH" in
  darwin/arm64|darwin/amd64|windows/arm64|windows/amd64|linux/arm64|linux/amd64) ;;
  *) printf 'Unsupported desktop target: %s/%s\n' "$GOOS" "$GOARCH" >&2; exit 2 ;;
esac
mode="${1:-all}"
case "$mode" in all|native) ;; *) printf 'Usage: %s [all|native]\n' "$0" >&2; exit 2 ;; esac
report="$(mktemp -d)"
trap 'rm -rf "$report"' EXIT
failed=0
check() {
  local dir="$1" pattern="$2" label="$3"
  printf '\nChecking %s (%s/%s, CGO_ENABLED=0, GOWORK=off)\n' "$label" "$GOOS" "$GOARCH"
  # -e exposes errors in every selected dependency, not only the first compiler failure.
  if ! (cd "$dir" && go list -e -deps -f '{{if .Error}}ERROR {{.ImportPath}}: {{.Error.Err}}{{end}}{{if .CgoFiles}}CGO {{.ImportPath}}: {{.CgoFiles}}{{end}}' "$pattern") > "$report/$label.list"; then
    failed=1
  fi
  if grep -v '^$' "$report/$label.list"; then
    failed=1
  fi
  if ! (cd "$dir" && go build "$pattern"); then
    failed=1
  fi
}
check "$root" './native/...' native
if [[ "$mode" == all ]]; then
  check "$root" './...' keel
fi
if [[ "$failed" != 0 ]]; then
  printf '\nDesktop no-cgo compilation gate FAILED; this target is not accepted.\n' >&2
  exit 1
fi
printf '\nNo-cgo compilation passed. Native GPU, IME and lifecycle remain separate gates.\n'
