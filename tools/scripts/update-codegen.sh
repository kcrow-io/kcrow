#!/bin/bash


set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(git rev-parse --show-toplevel)
CODEGEN_PKG=${CODEGEN_PKG_PATH:-$(cd ${PROJECT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
MODULE_NAME=$(cat ${PROJECT_ROOT}/go.mod | grep -e "module[[:space:]][^[:space:]]*" | awk '{print $2}')

SPDX_COPYRIGHT_HEADER="${PROJECT_ROOT}/tools/spdx-copyright-header.txt"
LICENSE_FILE="${PROJECT_ROOT}/tools/boilerplate.go.txt"
go_path="${PROJECT_ROOT}/_go"

cleanup() {
  rm -rf ${go_path}
  rm -f ${LICENSE_FILE}
}
trap "cleanup" EXIT SIGINT
cleanup

touch ${LICENSE_FILE}

while read -r line || [[ -n ${line} ]]
do
    echo "// ${line}" >>${LICENSE_FILE}
done < ${SPDX_COPYRIGHT_HEADER}

APIS_PKG="pkg/k8s/apis"
OUTPUT_PKG="pkg/k8s/client"
GROUPS_WITH_VERSIONS="kcrow.io:v1alpha1"

echo "change directory: ${PROJECT_ROOT}"
cd "${PROJECT_ROOT}"

bash ${PROJECT_ROOT}/${CODEGEN_PKG}/kube_codegen.sh kube::codegen::gen_client\
    --with-watch \
    --input-pkg-root ${MODULE_NAME}/${OUTPUT_PKG} \
    --output-pkg-root ${MODULE_NAME}/${APIS_PKG} \
    --output-base ${PROJECT_ROOT} \
    --boilerplate ${LICENSE_FILE}

rm -f ${LICENSE_FILE}
