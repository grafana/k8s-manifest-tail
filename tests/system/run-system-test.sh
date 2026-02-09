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

export CREATE_CLUSTER=${CREATE_CLUSTER:=true}
export DELETE_CLUSTER=${DELETE_CLUSTER:=false}

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
  testName=$(basename "${testDir}")
  echo "==> Starting list test ${testName}"
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
      if ! diff -u "${expectedList}" "${tmpList}"; then
        echo "ERROR: list output mismatch for ${testName}" >&2
        exit 1
      fi
    else
      echo "  - No expected list output found; creating baseline"
      cp "${tmpList}" "${expectedList}"
    fi

    echo "<== List test ${testName} passed"
  )
}

describe_test() {
  local testDir="$1"
  testName=$(basename "${testDir}")
  echo "==> Starting describe test ${testName}"
  (
    set -euo pipefail
    local config="${testDir}/config.yaml"
    local expectedDescribe="${testDir}/expected-describe-output.txt"

    tmpDescribe="$(mktemp)"
    cleanup() {
      if [[ -n "${tmpDescribe}" && -f "${tmpDescribe}" ]]; then
        rm -f "${tmpDescribe}"
      fi
    }
    trap cleanup EXIT

    (
      cd "${testDir}"
      "${BINARY}" describe --config "${config}" > "${tmpDescribe}"
    )

    if [[ -f "${expectedDescribe}" ]]; then
      echo "  - Validating describe output"
      if ! diff -u "${expectedDescribe}" "${tmpDescribe}"; then
        echo "ERROR: describe output mismatch for ${testName}" >&2
        exit 1
      fi
    else
      echo "  - No expected describe output found; creating baseline"
      cp "${tmpDescribe}" "${expectedDescribe}"
    fi

    echo "<== Describe test ${testName} passed"
  )
}

run_once_test() {
  local testDir="$1"
  testName=$(basename "${testDir}")
  echo "==> Starting run test ${testName}"
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
      if ! diff -ru "${expectedOutput}" "${tmpOutput}"; then
        echo "ERROR: run output mismatch for ${testName}" >&2
        exit 1
      fi
    else
      echo "  - No expected run output found; creating baseline"
      cp -rf "${tmpOutput}" "${expectedOutput}"
    fi

    echo "<== Run test ${testName} passed"
  )
}

if [[ "${CREATE_CLUSTER}" == "true" ]]; then
  create_cluster
fi

overall_rc=0
failed_tests=()
while IFS= read -r -d '' configFile; do
  testDir="$(dirname "${configFile}")"
  if ! describe_test "${testDir}"; then
    failed_tests+=("describe:${testDir}")
    overall_rc=1
  fi
  if ! list_test "${testDir}"; then
    failed_tests+=("list:${testDir}")
    overall_rc=1
  fi
  if ! run_once_test "${testDir}"; then
    failed_tests+=("run:${testDir}")
    overall_rc=1
  fi
done < <(find "${SCRIPT_DIR}" -name config.yaml -print0)

if [[ "${DELETE_CLUSTER}" == "true" ]]; then
  delete_cluster
fi

if [[ ${#failed_tests[@]} -gt 0 ]]; then
  echo
  echo "The following system test checks failed:"
  for entry in "${failed_tests[@]}"; do
    IFS=":" read -r testType testPath <<< "${entry}"
    echo "  - ${testType} (${testPath})"
  done
fi

exit "${overall_rc}"
