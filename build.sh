#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
ARCHS="${ARCHS:-amd64 arm64}"
CLI_TARGETS="${CLI_TARGETS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64}"
REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_PREFIX="${IMAGE_PREFIX:-ghcr.io/infrasecture}"
GOLANG_IMAGE="${GOLANG_IMAGE:-golang:1.22}"
NFPM_IMAGE="${NFPM_IMAGE:-ghcr.io/goreleaser/nfpm:latest}"
PUBLISH_LATEST="${PUBLISH_LATEST:-true}"
HARNESS_IMAGES="${HARNESS_IMAGES:-base codex claude opencode hermes all}"

DO_CLI=0
DO_IMAGES=0
DO_PACKAGES=0
DO_PUSH=0
DO_MANIFEST=0
REBUILD_IMAGES=0
SELECT_HARNESS=""

usage() {
  cat <<'EOF'
Usage: ./build.sh [--release] [--packages] [--push] [--manifest] [--images] [--cli]
                  [--rebuild-go] [--rebuild-images] [--harness <name>]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --release) DO_CLI=1; DO_IMAGES=1; DO_PACKAGES=1; DO_MANIFEST=1; shift ;;
    --packages) DO_PACKAGES=1; shift ;;
    --push) DO_PUSH=1; shift ;;
    --manifest) DO_MANIFEST=1; shift ;;
    --images) DO_IMAGES=1; shift ;;
    --cli) DO_CLI=1; shift ;;
    --rebuild-go) DO_CLI=1; shift ;;
    --rebuild-images) REBUILD_IMAGES=1; DO_IMAGES=1; shift ;;
    --harness) SELECT_HARNESS="${2:?missing harness}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ "${DO_CLI}${DO_IMAGES}${DO_PACKAGES}${DO_MANIFEST}" == "0000" ]]; then
  DO_CLI=1
fi

mkdir -p dist

build_cli() {
  local target goos goarch out
  for target in ${CLI_TARGETS}; do
    goos="${target%/*}"
    goarch="${target#*/}"
    out="dist/myharness-${goos}-${goarch}"
    echo "Building ${out}"
    GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "${out}" ./cmd/myharness
  done
  cp "dist/myharness-$(go env GOOS)-$(go env GOARCH)" dist/myharness-local 2>/dev/null || go build -o dist/myharness-local ./cmd/myharness
}

build_images() {
  local images image dockerfile tag extra=()
  images="${HARNESS_IMAGES}"
  if [[ -n "${SELECT_HARNESS}" ]]; then
    images="${SELECT_HARNESS}"
  fi
  if [[ "${REBUILD_IMAGES}" == "1" ]]; then
    extra+=(--no-cache)
  fi
  for image in ${images}; do
    dockerfile="images/${image}/Dockerfile"
    tag="${IMAGE_PREFIX}/myharness-${image}:${VERSION}"
    [[ -f "${dockerfile}" ]] || { echo "Missing ${dockerfile}" >&2; exit 1; }
    echo "Building ${tag}"
    docker build "${extra[@]}" -f "${dockerfile}" -t "${tag}" .
    if [[ "${PUBLISH_LATEST}" == "true" ]]; then
      docker tag "${tag}" "${IMAGE_PREFIX}/myharness-${image}:latest"
    fi
    if [[ "${DO_PUSH}" == "1" ]]; then
      push_image "${tag}"
      [[ "${PUBLISH_LATEST}" == "true" ]] && push_image "${IMAGE_PREFIX}/myharness-${image}:latest"
    fi
  done
}

push_image() {
  local image attempt
  image="$1"
  for attempt in 1 2 3; do
    if docker push "${image}"; then
      return 0
    fi
    if [[ "${attempt}" == "3" ]]; then
      break
    fi
    echo "docker push failed for ${image}; retrying (${attempt}/3)" >&2
    sleep $((attempt * 10))
  done
  return 1
}

build_packages() {
  local arch rpm_arch archlinux_arch stage config
  if [[ ! -f dist/myharness-linux-amd64 || ! -f dist/myharness-linux-arm64 ]]; then
    build_cli
  fi
  for arch in ${ARCHS}; do
    case "${arch}" in
      amd64) rpm_arch="x86_64"; archlinux_arch="x86_64" ;;
      arm64) rpm_arch="aarch64"; archlinux_arch="aarch64" ;;
      *) rpm_arch="${arch}"; archlinux_arch="${arch}" ;;
    esac
    stage="dist/package-${arch}"
    rm -rf "${stage}"
    mkdir -p "${stage}/usr/local/bin"
    cp "dist/myharness-linux-${arch}" "${stage}/usr/local/bin/myharness"
    cp bin/myCodex bin/myClaude bin/myOpenCode bin/myHermes "${stage}/usr/local/bin/"
    chmod 0755 "${stage}/usr/local/bin/"*

    config="dist/nfpm-${arch}.yaml"
    cat >"${config}" <<EOF
name: myharness
arch: ${arch}
platform: linux
version: ${VERSION#v}
section: utils
priority: optional
maintainer: infrasecture
description: Profile-driven containerized workstation for autonomous coding agents.
contents:
  - src: ${stage}/usr/local/bin/myharness
    dst: /usr/local/bin/myharness
  - src: ${stage}/usr/local/bin/myCodex
    dst: /usr/local/bin/myCodex
  - src: ${stage}/usr/local/bin/myClaude
    dst: /usr/local/bin/myClaude
  - src: ${stage}/usr/local/bin/myOpenCode
    dst: /usr/local/bin/myOpenCode
  - src: ${stage}/usr/local/bin/myHermes
    dst: /usr/local/bin/myHermes
EOF
    if command -v nfpm >/dev/null 2>&1; then
      nfpm package -f "${config}" -p deb -t "dist/myharness_${VERSION#v}_${arch}.deb"
      nfpm package -f "${config}" -p rpm -t "dist/myharness-${VERSION#v}-1.${rpm_arch}.rpm"
      nfpm package -f "${config}" -p archlinux -t "dist/myharness-${VERSION#v}-1-${archlinux_arch}.pkg.tar.zst"
    elif command -v docker >/dev/null 2>&1; then
      docker run --rm -v "$PWD:/workspace" -w /workspace "${NFPM_IMAGE}" package -f "${config}" -p deb -t "dist/myharness_${VERSION#v}_${arch}.deb"
      docker run --rm -v "$PWD:/workspace" -w /workspace "${NFPM_IMAGE}" package -f "${config}" -p rpm -t "dist/myharness-${VERSION#v}-1.${rpm_arch}.rpm"
      docker run --rm -v "$PWD:/workspace" -w /workspace "${NFPM_IMAGE}" package -f "${config}" -p archlinux -t "dist/myharness-${VERSION#v}-1-${archlinux_arch}.pkg.tar.zst"
    else
      echo "Skipping Linux packages for ${arch}: nfpm or docker is required." >&2
    fi

    if [[ -f "dist/myharness-darwin-${arch}" ]]; then
      cp "dist/myharness-darwin-${arch}" "${stage}/usr/local/bin/myharness"
      chmod 0755 "${stage}/usr/local/bin/myharness"
      tar -C "${stage}/usr/local/bin" -czf "dist/myharness-brew-darwin-${arch}.tar.gz" myharness myCodex myClaude myOpenCode myHermes
    fi
  done
}

write_checksums() {
  local sums
  sums="$(mktemp)"
  (cd dist && find . -maxdepth 1 -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > "${sums}")
  mv "${sums}" dist/SHA256SUMS
}

[[ "${DO_CLI}" == "1" ]] && build_cli
[[ "${DO_IMAGES}" == "1" ]] && build_images
[[ "${DO_PACKAGES}" == "1" ]] && build_packages
write_checksums

if [[ "${DO_MANIFEST}" == "1" ]]; then
  echo "Manifest lists are expected to be assembled by CI buildx for linux/amd64 and linux/arm64."
fi
