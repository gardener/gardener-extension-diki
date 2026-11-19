// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"time"

	extensionssecretsmanager "github.com/gardener/gardener/extensions/pkg/util/secret/manager"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	secretsutils "github.com/gardener/gardener/pkg/utils/secrets"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"

	"github.com/gardener/gardener-extension-diki/pkg/constants"
)

const (
	// ManagerIdentity is the identity used for the secrets manager.
	ManagerIdentity = "extension-" + constants.ExtensionType
	// CAName is the name of the CA secret.
	CAName = "ca-extension-" + constants.ExtensionType

	// caValidity is the validity of the extension CA certificate.
	caValidity = 2 * 365 * 24 * time.Hour
	// serverCertValidity is the validity of the webhook server certificate.
	serverCertValidity = 90 * 24 * time.Hour
)

// ConfigsFor returns configurations for the secrets manager for the given namespace.
func ConfigsFor(namespace string) []extensionssecretsmanager.SecretConfigWithOptions {
	return []extensionssecretsmanager.SecretConfigWithOptions{
		{
			Config: &secretsutils.CertificateSecretConfig{
				Name:       CAName,
				CommonName: CAName,
				CertType:   secretsutils.CACert,
				Validity:   new(caValidity),
			},
			Options: []secretsmanager.GenerateOption{secretsmanager.Persist()},
		},
		{
			Config: &secretsutils.CertificateSecretConfig{
				Name:                        constants.WebhookTLSSecretName,
				CommonName:                  constants.DikiOperatorName,
				DNSNames:                    kubernetesutils.DNSNamesForService(constants.DikiOperatorName, namespace),
				CertType:                    secretsutils.ServerCert,
				SkipPublishingCACertificate: true,
				Validity:                    new(serverCertValidity),
			},
			Options: []secretsmanager.GenerateOption{
				secretsmanager.SignedByCA(CAName, secretsmanager.UseCurrentCA),
			},
		},
	}
}
