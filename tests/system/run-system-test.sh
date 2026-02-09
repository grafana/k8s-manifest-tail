#!/usr/bin/env bash

set -euo pipefail

export PATH=/Users/petewall/src/grafana/helm-chart-toolbox/tools/helm-test:$PATH

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BINARY="${REPO_ROOT}/build/k8s-manifest-tail"

if [[ ! -x "${BINARY}" ]]; then
  echo "error: ${BINARY} not found or not executable. Run 'make build' first." >&2
  exit 1
fi

export KUBECONFIG="${SCRIPT_DIR}/kubeconfig.yaml"
clusterName=k8s-manifest-tail-test-cluster

delete_cluster() {
  echo "==> Deleting test cluster"
  kind delete cluster --name "${clusterName}"
  rm -f "${KUBECONFIG}"
  echo "<== Test cluster deleted"
}

create_cluster() {
  echo "==> Creating test cluster"
  if ! kind get clusters | grep "${clusterName}"; then
    kind create cluster --name "${clusterName}"
  fi

  echo "  - Deploying dependencies"
  kubectl apply -f "${SCRIPT_DIR}/dependencies"
  echo "<== Test cluster ready"
}

list_test() {
  local testDir="$1"
  testName=$(basename testDir)
  echo "==> Starting test ${testName}"
  (
    set -euo pipefail
    local config="${testDir}/config.yaml"
    local expectedList="${testDir}/expected-list-output.txt"

    tmpList="$(mktemp)"
    cleanup() {
      if [[ -n "${tmpList}" && -f "${tmpList}" ]]; then
        rm -f "${tmpList}"
      fi
    }
    trap cleanup EXIT

    (
      cd "${testDir}"
      "${BINARY}" list --config "${config}" > "${tmpList}"
    )

    if [[ -f "${expectedList}" ]]; then
      echo "  - Validating list output"
      diff -u "${expectedList}" "${tmpList}"
    else
      cp "${tmpList}" "${expectedList}"
    fi

    echo "<== Test ${testName} passed"
  )
}

run_once_test() {
  local testDir="$1"
  testName=$(basename testDir)
  echo "==> Starting test ${testName}"
  (
    set -euo pipefail
    local config="${testDir}/config.yaml"
    local expectedOutput="${testDir}/expected-output"

    tmpOutput="$(mktemp -d)"
    cleanup() {
      if [[ -n "${tmpOutput}" && -d "${tmpOutput}" ]]; then
        rm -rf "${tmpOutput}"
      fi
    }
    trap cleanup EXIT

    (
      cd "${testDir}"
      "${BINARY}" run-once --config "${config}" --output-directory "${tmpOutput}"
    )

    if [[ -d "${expectedOutput}" ]]; then
      echo "  - Validating run output"
      diff -ru "${expectedOutput}" "${tmpOutput}"
    else
      cp -rf "${tmpOutput}" "${expectedOutput}"
    fi

    echo "<== Test ${testName} passed"
  )
}

create_cluster

overall_rc=0
while IFS= read -r -d '' configFile; do
  testDir="$(dirname "${configFile}")"
  if ! list_test "${testDir}"; then
    overall_rc=1
  fi
  if ! run_once_test "${testDir}"; then
    overall_rc=1
  fi
done < <(find "${SCRIPT_DIR}" -name config.yaml -print0)

#delete_cluster

exit "${overall_rc}"
