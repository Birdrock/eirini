#!/bin/bash

set -euo pipefail

readonly BASEDIR="$(cd "$(dirname "$0")"/.. && pwd)"
export GO111MODULE=on
export EIRINIUSER_PASSWORD="${EIRINIUSER_PASSWORD:-$(pass eirini/docker-hub)}"

nodes=""
if [[ "${NODES:-}" != "" ]]; then
  nodes="-nodes $NODES"
fi

main() {
  export EIRINI_BINS_PATH
  EIRINI_BINS_PATH=$(mktemp -d)
  trap "rm -rf $EIRINI_BINS_PATH" EXIT

  pushd "$BASEDIR"/tests/integration >/dev/null || exit 1
  ginkgo -mod=vendor -p $nodes -r -keepGoing -tags=integration -randomizeAllSpecs -randomizeSuites -timeout=20m "$@"
  popd >/dev/null || exit 1
}

main "$@"
