#!/bin/bash

# Copyright 2014 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')

# check if golangci-lint installed
function check_golangci-lint() {
  echo "checking whether golangci-lint has been installed"
  command -v golangci-lint >/dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo "installing golangci-lint ."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.40.0
    if [[ $? -ne 0 ]]; then
      echo "golangci-lint installed failed, exiting."
      exit 1
    fi

    export PATH=$PATH:$GOPATH/bin
  else
    echo "found golangci-lint"
  fi
}

# run golangci-lint run to check codes
function golangci-lint_run() {
  echo "begin run golangci-lint"
  cd ${KUBE_ROOT}
  golangci-lint run -v
}

set +e
check_golangci-lint
set -e

set +e
golangci-lint_run
set -e
