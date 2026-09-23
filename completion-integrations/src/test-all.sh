#!/usr/bin/env bash

set -euo pipefail

# This script runs completion tests in different environments and different shells.
# Every image is built concurrently and every container is run concurrently, so the
# wall-clock time is roughly that of the slowest test rather than the sum of them all.

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

# Derive the shell list from the test files so it never goes stale.
shellTypes=()
for f in "${BASE_DIR}"/tests/comp-tests.*; do
  shellTypes+=("${f##*.}")
done

# Map a test name (e.g. alpine-bash-3.2) to the shell it exercises (bash or fish).
getTestShellType() {
  for shell in "${shellTypes[@]}"; do
    if [[ $1 == *"-$shell-"* ]]; then
      printf "%s" "$shell"
      return
    fi
  done
}

declare -A test_cases=()
declare -A build_pids=()
declare -A run_pids=()

# Build all images concurrently.
for testName in "$@"; do
  testFile="${BASE_DIR}/tests/Dockerfile.${testName}"
  imageName="comp-test:$testName"
  test_cases[$imageName]="$(getTestShellType "$testName")"

  (
    exec > >(
      trap "" INT TERM
      sed 's/^/'"$testName"': /'
    )
    exec 2> >(
      trap "" INT TERM
      sed 's/^/'"$testName"': /' >&2
    )
    $CONTAINER_ENGINE build "${engine_args[@]}" -t "${imageName}" "${BASE_DIR}" -f "$testFile"
  ) &
  build_pids[$testName]=$!
done

# Wait for every build and remember which ones failed. A failed build must not be run.
GOT_FAILURE=0
declare -A built=()
for testName in "${!build_pids[@]}"; do
  if wait "${build_pids[$testName]}"; then
    built[$testName]=1
  else
    echo "${testName}: build failed" >&2
    GOT_FAILURE=1
  fi
done

# Run every successfully built image concurrently.
for testName in "${!built[@]}"; do
  imageName="comp-test:$testName"
  shellType="${test_cases[$imageName]}"
  (
    exec > >(
      trap "" INT TERM
      sed 's/^/'"$testName"': /'
    )
    exec 2> >(
      trap "" INT TERM
      sed 's/^/'"$testName"': /' >&2
    )
    "$CONTAINER_ENGINE" run --rm "${imageName}" "tests/comp-tests.$shellType"
  ) &
  run_pids[$testName]=$!
done

for testName in "${!run_pids[@]}"; do
  if ! wait "${run_pids[$testName]}"; then
    echo "${testName}: test failed" >&2
    GOT_FAILURE=1
  fi
done

# Indicate if anything failed during the build or the run.
exit ${GOT_FAILURE}
