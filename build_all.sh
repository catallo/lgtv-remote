#!/bin/bash
set -e
PLATFORMS=(
  "linux/amd64" "linux/arm64" "linux/arm" "linux/386"
  "linux/mips" "linux/mips64" "linux/mipsle" "linux/mips64le"
  "linux/ppc64" "linux/ppc64le" "linux/riscv64" "linux/s390x" "linux/loong64"
  "darwin/amd64" "darwin/arm64"
  "windows/amd64" "windows/arm64" "windows/386"
  "freebsd/amd64" "freebsd/arm64" "freebsd/arm" "freebsd/386"
  "openbsd/amd64" "openbsd/arm64" "openbsd/arm" "openbsd/386" "openbsd/ppc64" "openbsd/riscv64"
  "netbsd/amd64" "netbsd/arm64" "netbsd/arm" "netbsd/386"
)
for p in "${PLATFORMS[@]}"; do
  GOOS=${p%/*} GOARCH=${p#*/}
  OUT="dist/lgtv-remote-${GOOS}-${GOARCH}"
  [ "$GOOS" = "windows" ] && OUT="${OUT}.exe"
  echo "Building $GOOS/$GOARCH..."
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags='-s -w' -o "$OUT" .
done
echo "Done! $(ls dist/ | wc -l) binaries built."
