#!/usr/bin/env bash
set -euo pipefail

/usr/local/bin/myharness-init.sh

if [[ $# -gt 0 ]]; then
  exec runuser -u harness -- bash --login -c 'exec "$@"' bash "$@"
fi

if [[ -t 0 && -t 1 && "${MYHARNESS_AUTO_ATTACH:-0}" == "1" ]]; then
  exec runuser -u harness -- byobu -r "${MYHARNESS_SESSION:-myharness}"
fi

exec sleep infinity
