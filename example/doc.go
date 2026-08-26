// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "$TOOLS_BIN_DIR/extension-generator --name=extension-diki --component-category=extension --provider-type=diki --extension-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/diki:$(cat ../VERSION) --destination=./extension/base/extension.yaml"
//go:generate sh -c "$TOOLS_BIN_DIR/kustomize build ./extension -o ./extension.yaml"

package example
