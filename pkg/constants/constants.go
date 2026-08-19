// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	// ExtensionType is the type of the diki extension.
	ExtensionType = "diki"
	// ControllerName is the name of the diki lifecycle controller.
	ControllerName = "diki-controller"
	// FinalizerSuffix is the finalizer suffix for the diki extension controller.
	FinalizerSuffix = "diki"
	// Origin is the origin used for the diki ManagedResources.
	Origin = "diki"
	// ImageNameDikiOperator is the image name for the diki-operator.
	ImageNameDikiOperator = "diki-operator"

	// ManagedResourceNameSeed is the name of the seed ManagedResource.
	ManagedResourceNameSeed = "extension-diki-seed"
	// ManagedResourceNameShoot is the name of the shoot ManagedResource.
	ManagedResourceNameShoot = "extension-diki-shoot"

	// DikiOperatorName is the application name for the diki-operator.
	DikiOperatorName = "diki-operator"

	// WebhookTLSSecretName is the name of the TLS secret used by the diki-operator webhook server.
	WebhookTLSSecretName = DikiOperatorName + "-tls"
)
