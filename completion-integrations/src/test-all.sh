#!/usr/bin/env bash

set -euo pipefail

# This script runs completion tests in different environments and different shells.

# Get path to docker or podman binary
CONTAINER_ENGINE="$(command -v podman docker | head -n1)"

if [[ -z "$CONTAINER_ENGINE" ]]; then
  echo "Missing 'docker' or 'podman' which is required for these tests"
  exit 2
fi

engine_args=()
[[ $CONTAINER_ENGINE == */docker ]] && engine_args+=("--load")

BASE_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." &>/dev/null && pwd)

export TESTS_DIR="${BASE_DIR}/tests"
export TESTPROG_DIR="${BASE_DIR}/testprog"
export TESTING_DIR="${BASE_DIR}/testingdir"

# Run all tests, even if there is a failure.
# But remember if there was any failure to report it at the end.
set +e
GOT_FAILURE=0
trap "GOT_FAILURE=1" ERR

mapfile -d '' shellTypes < <(find "${BASE_DIR}" -name "comp-tests.*" -printf '%f\0' | cut -z -c 12-)
getTestShellType() {
  for shell in "${shellTypes[@]}"; do
    if [[ $1 == *"-$shell-"* ]]; then
      printf "%s" "$shell"
      break
    fi
  done
}

declare -A test_cases=()

for testName in "$@"; do
  testFile="${BASE_DIR}/tests/Dockerfile.${testName}"

  imageName="comp-test:$testName"
  test_cases[$imageName]="$(getTestShellType "$testName")"

  (
    exec > >(trap "" INT TERM; sed 's/^/'"$testName"': /')
    exec 2> >(trap "" INT TERM; sed 's/^/'"$testName"': /' >&2)
    $CONTAINER_ENGINE build "${engine_args[@]}" -t "${imageName}" "${BASE_DIR}" -f "$testFile"
  ) &
done

wait

for imageName in "${!test_cases[@]}"; do
  shellType="${test_cases[$imageName]}"
  "$CONTAINER_ENGINE" run --rm "${imageName}" "tests/comp-tests.$shellType"
done
# Indicate if anything failed during the run
exit ${GOT_FAILURE}
