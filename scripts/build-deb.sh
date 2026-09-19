#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <vMAJOR.MINOR.PATCH>" >&2
  exit 2
fi

tag=$1
if [[ ! $tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "tag must have the form vMAJOR.MINOR.PATCH: $tag" >&2
  exit 2
fi
version=${tag#v}

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
project_dir=$(cd -- "$script_dir/.." && pwd)
cd "$project_dir"

rm -f packagecloud_*.deb

build_and_package() {
  local arch=$1 goos=$2 goarch=$3 goarm=$4
  local work_dir
  work_dir=$(mktemp -d "${TMPDIR:-/tmp}/packagecloud-deb.XXXXXX")
  trap 'rm -rf -- "$work_dir"' RETURN

  env CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" GOARM="$goarm" \
    go build -trimpath -ldflags='-s -w' -o packagecloud .

  go-bin-deb generate \
    --file deb.json \
    --version "$version" \
    --arch "$arch" \
    --wd "$work_dir" \
    --output "$project_dir"
}

build_and_package arm64 linux arm64 ""
build_and_package armhf linux arm 7
build_and_package amd64 linux amd64 ""
