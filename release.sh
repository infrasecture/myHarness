#!/usr/bin/env bash
set -euo pipefail

NIGHTLY=0
TITLE=""
NOTES_FILE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --nightly) NIGHTLY=1; shift ;;
    --title) TITLE="${2:?missing title}"; shift 2 ;;
    --notes-file) NOTES_FILE="${2:?missing notes file}"; shift 2 ;;
    -h|--help) echo "Usage: ./release.sh [--nightly] [--title <title>] [--notes-file <file>]"; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 1; }
}

require git
require gh

update_tap() {
  local formula tag_name stable formula_name formula_class tar_amd64 tar_arm64 sha_amd64 sha_arm64
  tag_name="$1"
  stable="$2"
  formula_name="myharness"
  if [[ "${stable}" != "1" ]]; then
    formula_name="myharness-nightly"
    formula_class="MyharnessNightly"
  else
    formula_class="Myharness"
  fi
  tar_amd64="dist/myharness-brew-darwin-amd64.tar.gz"
  tar_arm64="dist/myharness-brew-darwin-arm64.tar.gz"
  [[ -f "${tar_amd64}" && -f "${tar_arm64}" ]] || { echo "Skipping tap update: missing Homebrew bundles." >&2; return; }
  sha_amd64="$(sha256sum "${tar_amd64}" | awk '{print $1}')"
  sha_arm64="$(sha256sum "${tar_arm64}" | awk '{print $1}')"

  TAP_REPO="${TAP_REPO:-infrasecture/tap}"
  TAP_DIR="${TAP_DIR:-dist/tap}"
  if [[ ! -d "${TAP_DIR}/.git" ]]; then
    gh repo clone "${TAP_REPO}" "${TAP_DIR}"
  fi
  mkdir -p "${TAP_DIR}/Formula"
  formula="${TAP_DIR}/Formula/${formula_name}.rb"
  cat >"${formula}" <<EOF
class ${formula_class} < Formula
  desc "Profile-driven containerized workstation for autonomous coding agents"
  homepage "https://github.com/infrasecture/myHarness"
  version "${tag_name#v}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/infrasecture/myHarness/releases/download/${tag_name}/myharness-brew-darwin-arm64.tar.gz"
      sha256 "${sha_arm64}"
    else
      url "https://github.com/infrasecture/myHarness/releases/download/${tag_name}/myharness-brew-darwin-amd64.tar.gz"
      sha256 "${sha_amd64}"
    end
  end

  def install
    bin.install "myharness"
    bin.install "myCodex"
    bin.install "myClaude"
    bin.install "myOpenCode"
    bin.install "myHermes"
  end

  test do
    output = shell_output("#{bin}/myharness version")
    assert_match "myharness ", output
  end
end
EOF
  (
    cd "${TAP_DIR}"
    git add "Formula/${formula_name}.rb"
    if git diff --cached --quiet; then
      echo "Tap formula already current: ${formula_name}"
    else
      git commit -m "Update ${formula_name} to ${tag_name}"
      git push
    fi
  )
}

if [[ "${NIGHTLY}" == "1" ]]; then
  TAG="nightly-$(git rev-parse --short HEAD)"
  PRERELEASE=(--prerelease)
  export PUBLISH_LATEST=false
else
  [[ -z "$(git status --porcelain)" ]] || { echo "Stable release requires a clean working tree." >&2; exit 1; }
  TAGS="$(git tag --points-at HEAD | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' || true)"
  [[ "$(printf '%s\n' "${TAGS}" | sed '/^$/d' | wc -l)" == "1" ]] || { echo "Stable release requires exactly one vX.Y.Z tag on HEAD." >&2; exit 1; }
  TAG="$(printf '%s\n' "${TAGS}" | sed '/^$/d')"
  PRERELEASE=()
fi

if gh release view "${TAG}" >/dev/null 2>&1; then
  echo "Release already exists: ${TAG}" >&2
  exit 1
fi

VERSION="${TAG}" ./build.sh --release --packages --push

args=(release create "${TAG}" dist/* --title "${TITLE:-${TAG}}" "${PRERELEASE[@]}")
if [[ -n "${NOTES_FILE}" ]]; then
  args+=(--notes-file "${NOTES_FILE}")
else
  args+=(--generate-notes)
fi
gh "${args[@]}"

if [[ "${NIGHTLY}" == "1" ]]; then
  update_tap "${TAG}" 0
else
  update_tap "${TAG}" 1
fi
