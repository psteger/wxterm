#!/usr/bin/env bash
# Build the browser (js/wasm) version of wxterm into OUT_DIR (default: ./wasm/dist):
#   wxterm.wasm   the program
#   wasm_exec.js  Go's WebAssembly loader (must match the Go version that built the .wasm)
#
# Bubble Tea v1 and atotto/clipboard have no js support. Rather than forking them,
# this copies the two modules to a temp dir, adds small js stubs (wasm/stubs/), and
# builds with -modfile so this repo's go.mod/go.sum are untouched.
set -euo pipefail

cd "$(dirname "$0")/.."
OUT_DIR="${1:-wasm/dist}"
WORK="$(mktemp -d)"
trap 'chmod -R u+w "$WORK"; rm -rf "$WORK"' EXIT

# go list only reports source directories for modules already downloaded.
go mod download github.com/charmbracelet/bubbletea github.com/atotto/clipboard
TEA_DIR="$(go list -m -f '{{.Dir}}' github.com/charmbracelet/bubbletea)"
CLIP_DIR="$(go list -m -f '{{.Dir}}' github.com/atotto/clipboard)"
cp -r "$TEA_DIR" "$WORK/bubbletea"
cp -r "$CLIP_DIR" "$WORK/clipboard"
chmod -R u+w "$WORK"
cp wasm/stubs/tea_js.go "$WORK/bubbletea/"
cp wasm/stubs/clipboard_js.go "$WORK/clipboard/"

cp go.mod "$WORK/wasm.mod"
cp go.sum "$WORK/wasm.sum"
{
	echo
	echo "replace github.com/charmbracelet/bubbletea => $WORK/bubbletea"
	echo "replace github.com/atotto/clipboard => $WORK/clipboard"
} >>"$WORK/wasm.mod"

mkdir -p "$OUT_DIR"
VERSION="${WXTERM_VERSION:-dev}"
GOOS=js GOARCH=wasm go build -modfile="$WORK/wasm.mod" -trimpath \
	-ldflags "-s -w -X main.version=$VERSION" \
	-o "$OUT_DIR/wxterm.wasm" .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT_DIR/wasm_exec.js"
echo "built $OUT_DIR/wxterm.wasm ($(du -h "$OUT_DIR/wxterm.wasm" | cut -f1)) with $(go version | cut -d' ' -f3)"
