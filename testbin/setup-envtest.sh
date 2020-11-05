#!/usr/bin/env bash

#  Copyright 2020 The Kubernetes Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

set -o errexit
set -o pipefail

tmp_root=/tmp
envtest_root_dir=$tmp_root/envtest
dest_dir="${1}"

k8s_version="${ENVTEST_K8S_VERSION:-1.16.4}"
goarch="$(go env GOARCH)"
goos="$(go env GOOS)"


# use the pre-existing version in the temporary folder if it matches our k8s version
if [[ -x "${dest_dir}/bin/kube-apiserver" ]]; then
  version=$("${dest_dir}"/bin/kube-apiserver --version)
  if [[ $version == *"${k8s_version}"* ]]; then
    echo "Using cached envtest tools from ${dest_dir}"
    exit 0
  fi
fi

echo "fetching envtest tools@${k8s_version} (into '${dest_dir}')"
envtest_tools_archive_name="kubebuilder-tools-$k8s_version-$goos-$goarch.tar.gz"
envtest_tools_download_url="https://storage.googleapis.com/kubebuilder-tools/$envtest_tools_archive_name"

envtest_tools_archive_path="$tmp_root/$envtest_tools_archive_name"
if [ ! -f $envtest_tools_archive_path ]; then
  curl -sL ${envtest_tools_download_url} -o "$envtest_tools_archive_path"
fi

mkdir -p "${dest_dir}"
tar -C "${dest_dir}" --strip-components=1 -zvxf "$envtest_tools_archive_path"

echo "setting up env vars"

# Setup env vars
KUBEBUILDER_ASSETS=${KUBEBUILDER_ASSETS:-""}
if [[ -z "${KUBEBUILDER_ASSETS}" ]]; then
  export KUBEBUILDER_ASSETS="$1/bin"
fi
