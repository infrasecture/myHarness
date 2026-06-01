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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TAP_SUBMODULE="homebrew-tap"
TAP_REPO_URL="https://github.com/infrasecture/homebrew-tap.git"

update_tap() {
  local current_branch formula tag_name stable formula_name formula_class tap_path tar_amd64 tar_arm64 sha_amd64 sha_arm64
  tag_name="$1"
  stable="$2"
  formula_name="myharness"
  if [[ "${stable}" != "1" ]]; then
    formula_name="myharness-nightly"
    formula_class="MyharnessNightly"
  else
    formula_class="Myharness"
  fi
  tar_amd64="${SCRIPT_DIR}/dist/myharness-brew-darwin-amd64.tar.gz"
  tar_arm64="${SCRIPT_DIR}/dist/myharness-brew-darwin-arm64.tar.gz"
  [[ -f "${tar_amd64}" && -f "${tar_arm64}" ]] || { echo "Skipping tap update: missing Homebrew bundles." >&2; return; }
  sha_amd64="$(sha256sum "${tar_amd64}" | awk '{print $1}')"
  sha_arm64="$(sha256sum "${tar_arm64}" | awk '{print $1}')"

  tap_path="$(git -C "${SCRIPT_DIR}" config -f .gitmodules --get "submodule.${TAP_SUBMODULE}.path" || true)"
  [[ "${tap_path}" == "${TAP_SUBMODULE}" ]] || { echo "Missing ${TAP_SUBMODULE} submodule configuration." >&2; exit 1; }

  if [[ -n "${TAP_TOKEN:-}" ]]; then
    git config --global url."https://x-access-token:${TAP_TOKEN}@github.com/".insteadOf "https://github.com/"
  elif [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "HOMEBREW_TAP_TOKEN is required to push ${TAP_REPO_URL} from GitHub Actions." >&2
    exit 1
  fi

  git -C "${SCRIPT_DIR}" submodule update --init --checkout -- "${TAP_SUBMODULE}"
  [[ -z "$(git -C "${SCRIPT_DIR}/${TAP_SUBMODULE}" status --porcelain)" ]] || { echo "${TAP_SUBMODULE} has uncommitted changes." >&2; exit 1; }
  git -C "${SCRIPT_DIR}/${TAP_SUBMODULE}" fetch origin main
  git -C "${SCRIPT_DIR}/${TAP_SUBMODULE}" checkout -B main origin/main

  mkdir -p "${SCRIPT_DIR}/${TAP_SUBMODULE}/Formula"
  formula="${SCRIPT_DIR}/${TAP_SUBMODULE}/Formula/${formula_name}.rb"
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
    cd "${SCRIPT_DIR}/${TAP_SUBMODULE}"
    git add "Formula/${formula_name}.rb"
    if git diff --cached --quiet; then
      echo "Tap formula already current: ${formula_name}"
    else
      git commit -m "chore(formula): update ${formula_name} to ${tag_name}"
      git push origin HEAD
    fi
  )

  current_branch="$(git -C "${SCRIPT_DIR}" symbolic-ref --short HEAD || true)"
  [[ -n "${current_branch}" ]] || { echo "Cannot update ${TAP_SUBMODULE} pointer from a detached HEAD." >&2; exit 1; }
  git -C "${SCRIPT_DIR}" add "${TAP_SUBMODULE}"
  if git -C "${SCRIPT_DIR}" diff --cached --quiet; then
    echo "Homebrew tap submodule already current."
  else
    git -C "${SCRIPT_DIR}" commit -m "chore(submodule): bump homebrew-tap after ${tag_name} release"
    git -C "${SCRIPT_DIR}" push origin "HEAD:${current_branch}"
  fi
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

mapfile -t assets < <(find dist -maxdepth 1 -type f ! -name myharness-local ! -name 'nfpm-*.yaml' | sort)
args=(release create "${TAG}" "${assets[@]}" --title "${TITLE:-${TAG}}" "${PRERELEASE[@]}")
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
