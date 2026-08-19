// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionssecretsmanager "github.com/gardener/gardener/extensions/pkg/util/secret/manager"
	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/extensions"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	secretsutils "github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-diki/imagevector"
	"github.com/gardener/gardener-extension-diki/pkg/component/dikioperator"
	"github.com/gardener/gardener-extension-diki/pkg/constants"
	"github.com/gardener/gardener-extension-diki/pkg/secrets"
)

const (
	operatorShootAccessSecretNamePrefix = "diki-operator"
	runnerShootAccessSecretNamePrefix   = "diki-runner"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(c client.Client) extension.Actuator {
	return &actuator{
		client: c,
	}
}

type actuator struct {
	client client.Client
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	var (
		replicas  int32 = 1
		namespace       = ex.GetNamespace()
	)

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	if v1beta1helper.HibernationIsEnabled(cluster.Shoot) {
		replicas = 0
	}

	operatorShootAccessSecret := gutil.NewShootAccessSecret(gutil.SecretNamePrefixShootAccess+operatorShootAccessSecretNamePrefix, namespace)
	if err := operatorShootAccessSecret.Reconcile(ctx, a.client); err != nil {
		return fmt.Errorf("failed to reconcile operator shoot access secret: %w", err)
	}

	runnerShootAccessSecret := gutil.NewShootAccessSecret(gutil.SecretNamePrefixShootAccess+runnerShootAccessSecretNamePrefix, namespace)
	if err := runnerShootAccessSecret.Reconcile(ctx, a.client); err != nil {
		return fmt.Errorf("failed to reconcile runner shoot access secret: %w", err)
	}

	configs := secrets.ConfigsFor(namespace)
	secretsManager, err := extensionssecretsmanager.SecretsManagerForCluster(ctx, logger.WithName("secretsmanager"), clock.RealClock{}, a.client, cluster, secrets.ManagerIdentity, configs)
	if err != nil {
		return fmt.Errorf("failed to create secrets manager: %w", err)
	}

	generatedSecrets, err := extensionssecretsmanager.GenerateAllSecrets(ctx, secretsManager, configs)
	if err != nil {
		return fmt.Errorf("failed to generate secrets: %w", err)
	}

	caBundleSecret, found := secretsManager.Get(secrets.CAName)
	if !found {
		return fmt.Errorf("secret %q not found", secrets.CAName)
	}

	comp, err := a.newComponent(namespace, cluster, replicas, operatorShootAccessSecret, runnerShootAccessSecret, generatedSecrets[constants.WebhookTLSSecretName].Name, caBundleSecret.Data[secretsutils.DataKeyCertificateBundle])
	if err != nil {
		return err
	}

	if err := comp.Deploy(ctx); err != nil {
		return fmt.Errorf("failed to deploy diki-operator: %w", err)
	}

	if err := comp.Wait(ctx); err != nil {
		return err
	}

	return secretsManager.Cleanup(ctx)
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.delete(ctx, logger, ex, false, false)
}

// ForceDelete the Extension resource.
//
// We don't need to wait for the ManagedResource deletion because ManagedResources are finalized by gardenlet
// in a later step in the Shoot force deletion flow.
func (a *actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.delete(ctx, logger, ex, false, true)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.delete(ctx, logger, ex, true, false)
}

func (a *actuator) delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension, skipSecretsManagerCleanup, forceDelete bool) error {
	namespace := ex.GetNamespace()

	cluster, err := controller.GetCluster(ctx, a.client, namespace)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	operatorShootAccessSecret := gutil.NewShootAccessSecret(gutil.SecretNamePrefixShootAccess+operatorShootAccessSecretNamePrefix, namespace)
	runnerShootAccessSecret := gutil.NewShootAccessSecret(gutil.SecretNamePrefixShootAccess+runnerShootAccessSecretNamePrefix, namespace)

	comp, err := a.newComponent(namespace, cluster, 0, operatorShootAccessSecret, runnerShootAccessSecret, "", nil)
	if err != nil {
		return err
	}

	if skipSecretsManagerCleanup {
		if err := managedresources.SetKeepObjects(ctx, a.client, namespace, constants.ManagedResourceNameShoot, true); err != nil {
			return fmt.Errorf("failed to set keep objects on shoot ManagedResource: %w", err)
		}
	}

	if err := comp.Destroy(ctx, forceDelete); err != nil {
		return fmt.Errorf("failed to destroy diki-operator: %w", err)
	}

	if err := client.IgnoreNotFound(a.client.Delete(ctx, operatorShootAccessSecret.Secret)); err != nil {
		return err
	}
	if err := client.IgnoreNotFound(a.client.Delete(ctx, runnerShootAccessSecret.Secret)); err != nil {
		return err
	}

	if skipSecretsManagerCleanup {
		return nil
	}

	secretsManager, err := extensionssecretsmanager.SecretsManagerForCluster(ctx, logger.WithName("secretsmanager"), clock.RealClock{}, a.client, cluster, secrets.ManagerIdentity, nil)
	if err != nil {
		return fmt.Errorf("failed to create secrets manager: %w", err)
	}

	return secretsManager.Cleanup(ctx)
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}

func (a *actuator) newComponent(namespace string, cluster *extensions.Cluster, replicas int32, operatorShootAccessSecret, runnerShootAccessSecret *gutil.AccessSecret, serverTLSSecretName string, webhookCABundle []byte) (*dikioperator.Component, error) {
	image, err := imagevector.ImageVector().FindImage(constants.ImageNameDikiOperator)
	if err != nil {
		return nil, fmt.Errorf("failed to find image %s: %w", constants.ImageNameDikiOperator, err)
	}

	return dikioperator.New(a.client, dikioperator.Values{
		Image:                                 image.String(),
		Replicas:                              replicas,
		Namespace:                             namespace,
		GenericTokenKubeconfigSecretName:      extensions.GenericTokenKubeconfigSecretNameFromCluster(cluster),
		OperatorShootAccessSecretName:         operatorShootAccessSecret.Secret.Name,
		OperatorShootAccessServiceAccountName: operatorShootAccessSecret.ServiceAccountName,
		RunnerShootAccessSecretName:           runnerShootAccessSecret.Secret.Name,
		RunnerShootAccessServiceAccountName:   runnerShootAccessSecret.ServiceAccountName,
		ServerTLSSecretName:                   serverTLSSecretName,
		WebhookCABundle:                       webhookCABundle,
	}), nil
}
