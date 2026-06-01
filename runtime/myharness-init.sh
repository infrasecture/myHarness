#!/usr/bin/env bash
set -euo pipefail

USER_NAME="harness"
GROUP_NAME="harness"
HOME_DIR="${MYHARNESS_HOME:-/home/myharness}"
HOST_UID="${MYHARNESS_HOST_UID:-1000}"
HOST_GID="${MYHARNESS_HOST_GID:-1000}"
SESSION="${MYHARNESS_SESSION:-myharness}"
HARNESS="${MYHARNESS_HARNESS:-codex}"

if ! getent group "${GROUP_NAME}" >/dev/null; then
  groupadd --gid "${HOST_GID}" "${GROUP_NAME}"
else
  groupmod --gid "${HOST_GID}" "${GROUP_NAME}" 2>/dev/null || true
fi

if ! id -u "${USER_NAME}" >/dev/null 2>&1; then
  useradd --uid "${HOST_UID}" --gid "${GROUP_NAME}" --home-dir "${HOME_DIR}" --create-home --shell /bin/bash "${USER_NAME}"
else
  usermod --uid "${HOST_UID}" --gid "${HOST_GID}" --home "${HOME_DIR}" "${USER_NAME}" 2>/dev/null || true
fi

mkdir -p "${HOME_DIR}" /workspace

if [[ -f /etc/myharness/profile.json ]]; then
  python3 <<'PY'
import json
import os
import pathlib
import stat

home = pathlib.Path(os.environ.get("MYHARNESS_HOME", "/home/myharness"))
profile_path = pathlib.Path("/etc/myharness/profile.json")
profile = json.loads(profile_path.read_text())

for directory in profile.get("homeDirs") or []:
    path = pathlib.Path(directory)
    if path == home or home in path.parents:
        path.mkdir(parents=True, exist_ok=True)

for item in (profile.get("config") or {}).get("files") or []:
    path = pathlib.Path(item["path"])
    if path == home or home in path.parents:
        path.parent.mkdir(parents=True, exist_ok=True)
        if not path.exists():
            path.write_text(item.get("content", ""))
        mode = int(item.get("mode", "0600"), 8)
        path.chmod(mode)

banner = profile.get("banner")
if banner:
    pathlib.Path("/etc/myharness/profile-banner.txt").write_text(str(banner) + "\n")
PY
fi

chown "${USER_NAME}:${GROUP_NAME}" "${HOME_DIR}" || true
find "${HOME_DIR}" -maxdepth 2 -type d -exec chown "${USER_NAME}:${GROUP_NAME}" {} + 2>/dev/null || true
find "${HOME_DIR}" -maxdepth 3 -type f -exec chown "${USER_NAME}:${GROUP_NAME}" {} + 2>/dev/null || true

if command -v byobu-ctrl-a >/dev/null; then
  runuser -u "${USER_NAME}" -- byobu-ctrl-a screen >/dev/null 2>&1 || true
fi

if ! runuser -u "${USER_NAME}" -- byobu-tmux has-session -t "${SESSION}" 2>/dev/null; then
  STARTUP_CMD="$(cat <<EOF
clear
if [[ -f /etc/myharness/session-banner.txt ]]; then
  cat /etc/myharness/session-banner.txt
fi
if [[ -f /etc/myharness/profile-banner.txt ]]; then
  echo
  cat /etc/myharness/profile-banner.txt
fi
echo
echo "Harness: ${HARNESS}"
echo "Workspace: /workspace"
echo "Session: ${SESSION}"
echo
exec bash --login
EOF
)"
  runuser -u "${USER_NAME}" -- byobu-tmux new-session -d -s "${SESSION}" bash --login -lc "${STARTUP_CMD}"
fi
