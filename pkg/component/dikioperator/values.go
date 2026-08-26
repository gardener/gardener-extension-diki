// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator

const (
	configVolumeName = "diki-operator-config"
	configMountPath  = "/etc/diki-operator/config"
	configFileName   = "config.yaml"

	webhookTLSVolumeName = "webhook-tls"
	webhookTLSMountPath  = "/etc/diki-operator/webhooks/tls"

	portHealth  int32 = 8081
	portMetrics int32 = 8080
	portWebhook int32 = 10443
)

// Values contains the configuration values for the diki-operator component.
type Values struct {
	// Image is the container image for the diki-operator.
	Image string
	// Replicas is the number of diki-operator replicas.
	Replicas int32
	// Namespace is the shoot namespace on the seed.
	Namespace string
	// GenericTokenKubeconfigSecretName is the name of the generic token kubeconfig secret.
	GenericTokenKubeconfigSecretName string
	// OperatorShootAccessSecretName is the name of the shoot access secret for the diki-operator.
	OperatorShootAccessSecretName string
	// OperatorShootAccessServiceAccountName is the name of the service account in the shoot
	// for which the gardenlet populates the token for the diki-operator.
	OperatorShootAccessServiceAccountName string
	// RunnerShootAccessSecretName is the name of the shoot access secret for the diki-runner.
	RunnerShootAccessSecretName string
	// RunnerShootAccessServiceAccountName is the name of the service account in the shoot
	// for which the gardenlet populates the token for the diki-runner.
	RunnerShootAccessServiceAccountName string
	// ServerTLSSecretName is the name of the TLS secret for the webhook server.
	ServerTLSSecretName string
	// WebhookCABundle is the CA bundle used to verify the webhook server certificate.
	WebhookCABundle []byte
}
