#!/usr/bin/env bash

set -euo pipefail

# Run the completion tests natively, without a container engine. This is used on
# macOS, where the testprog is built for the host instead of for linux.
#
# Usage: test-completion.sh <shell> [shell...]

BASE_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." &>/dev/null && pwd)

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 <shell> [shell...]" >&2
  exit 2
fi

cd "$BASE_DIR"

GOT_FAILURE=0
for shell in "$@"; do
  testScript="tests/comp-tests.$shell"
  if [[ ! -f "$testScript" ]]; then
    echo "No completion test script for shell: $shell" >&2
    GOT_FAILURE=1
    continue
  fi
  echo "===================================================="
  echo "Running $shell completion tests natively"
  echo "===================================================="
  if ! "$testScript"; then
    GOT_FAILURE=1
  fi
done

exit ${GOT_FAILURE}
