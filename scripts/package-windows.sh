#!/usr/bin/env bash
set -euo pipefail

version="${1:-0.0.1}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"

rm -rf "$dist"
mkdir -p "$dist"

for arch in amd64 arm64; do
  package="vmware-mcp-server_${version}_windows_${arch}"
  out="$dist/$package"
  mkdir -p "$out"

  GOOS=windows GOARCH="$arch" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o "$out/vmware-mcp-agent.exe" \
    "$root/cmd/vmware-mcp-agent"

  GOOS=windows GOARCH="$arch" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o "$out/vmware-mcp-install.exe" \
    "$root/cmd/vmware-mcp-install"

  cp "$root/README.md" "$out/README.md"
  cp "$root/LICENSE" "$out/LICENSE"
  cp "$root/NOTICE" "$out/NOTICE"

  (
    cd "$dist"
    zip -qr "$package.zip" "$package"
  )
done
