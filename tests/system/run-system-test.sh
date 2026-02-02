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

run_test() {
  local testDir="$1"
  (
    set -euo pipefail

    export KUBECONFIG="${testDir}/kubeconfig.yaml"
    local testPlan="${testDir}/test-plan.yaml"
    local config="${testDir}/config.yaml"
    local expectedList="${testDir}/expected/list-output.txt"
    local expectedOutput="${testDir}/expected/output"

    if [[ ! -f "${testPlan}" || ! -f "${config}" ]]; then
      echo "Skipping ${testDir}: missing test-plan.yaml or config.yaml" >&2
      exit 0
    fi

    if [[ ! -f "${expectedList}" || ! -d "${expectedOutput}" ]]; then
      echo "Skipping ${testDir}: missing expected outputs" >&2
      exit 0
    fi

    local testName="$(yq -r '.name' "${testPlan}")"
    echo "==> Running system test: ${testName}"

    local tmpList=""
    local tmpOutput=""
    cleanup() {
      if [[ -n "${tmpList}" && -f "${tmpList}" ]]; then
        rm -f "${tmpList}"
      fi
      if [[ -n "${tmpOutput}" && -d "${tmpOutput}" ]]; then
        rm -rf "${tmpOutput}"
      fi
      echo "--> Deleting cluster for ${testName}"
      delete-cluster "${testDir}" || true
      rm -f "${KUBECONFIG}"
    }
    trap cleanup EXIT

    echo "--> Creating cluster for ${testName}"
    create-cluster "${testDir}"

    echo "--> Deploying dependencies for ${testName}"
    deploy-dependencies "${testDir}"

    echo "--> Validating list output"
    tmpList="$(mktemp)"
    (
      cd "${testDir}"
      "${BINARY}" list --config "${config}" >"${tmpList}"
    )
    diff -u "${expectedList}" "${tmpList}"

    echo "--> Running manifest collection"
    tmpOutput="$(mktemp -d)"
    (
      cd "${testDir}"
      "${BINARY}" run --config "${config}" --output-directory "${tmpOutput}"
    )
    diff -ru "${expectedOutput}" "${tmpOutput}"

    echo "==> ${testName} passed"
  )
}

overall_rc=0
while IFS= read -r -d '' planFile; do
  testDir="$(dirname "${planFile}")"
  if ! run_test "${testDir}"; then
    overall_rc=1
  fi
done < <(find "${SCRIPT_DIR}" -name test-plan.yaml -print0)

exit "${overall_rc}"
