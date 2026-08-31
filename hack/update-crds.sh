#!/usr/bin/env bash

# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

IMAGES_YAML="$REPO_ROOT/imagevector/images.yaml"
CRD_DIR="$REPO_ROOT/pkg/component/dikioperator/crds"

DIKI_OPERATOR_REPO="gardener/diki-operator"
# Path inside the diki-operator repository where CRDs are maintained.
DIKI_OPERATOR_CRD_PATH="pkg/apis/diki/crds"

CRD_FILES=(
  "diki.gardener.cloud_compliancescans.yaml"
  "diki.gardener.cloud_reportoutputs.yaml"
  "diki.gardener.cloud_scheduledcompliancescans.yaml"
)

# Extract the diki-operator image tag from the image vector.
version=$(sed -n '/name: diki-operator/,/^-/{s/.*tag: *//p}' "$IMAGES_YAML" | head -1)
if [[ -z "$version" ]]; then
  echo "Error: could not determine diki-operator version from $IMAGES_YAML"
  exit 1
fi

echo "> Updating CRDs from diki-operator $version"

for crd in "${CRD_FILES[@]}"; do
  echo "  Downloading $crd ..."
  curl -sSfL "https://raw.githubusercontent.com/${DIKI_OPERATOR_REPO}/${version}/${DIKI_OPERATOR_CRD_PATH}/${crd}" \
    -o "$CRD_DIR/$crd"
done

echo "> CRDs updated successfully"
