// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-diki/pkg/constants"
)

var (
	//go:embed crds/diki.gardener.cloud_compliancescans.yaml
	crdComplianceScans []byte
	//go:embed crds/diki.gardener.cloud_reportoutputs.yaml
	crdReportOutputs []byte
	//go:embed crds/diki.gardener.cloud_scheduledcompliancescans.yaml
	crdScheduledComplianceScans []byte
)

// New creates a new dikioperator component.
func New(c client.Client, values Values) *Component {
	return &Component{
		client: c,
		values: values,
	}
}

// Component deploys and manages the diki-operator on the seed and shoot.
type Component struct {
	client client.Client
	values Values
}

// Deploy creates or updates the diki-operator seed and shoot resources.
func (c *Component) Deploy(ctx context.Context) error {
	shootResources, err := c.shootResources()
	if err != nil {
		return fmt.Errorf("failed to build shoot resources: %w", err)
	}

	if err := managedresources.CreateForShoot(ctx, c.client, c.values.Namespace, constants.ManagedResourceNameShoot, constants.Origin, false, shootResources); err != nil {
		return fmt.Errorf("failed to create shoot ManagedResource: %w", err)
	}

	seedResources, err := c.seedResources()
	if err != nil {
		return fmt.Errorf("failed to build seed resources: %w", err)
	}

	if err := managedresources.CreateForSeed(ctx, c.client, c.values.Namespace, constants.ManagedResourceNameSeed, false, seedResources); err != nil {
		return fmt.Errorf("failed to create seed ManagedResource: %w", err)
	}

	return nil
}

// Destroy removes the diki-operator shoot and seed resources.
// Shoot resources are deleted and awaited before the seed resources so that
// the operator can handle finalizer-based cleanup before it is removed.
// When forceDelete is true, waiting for resource deletion is skipped because
// ManagedResources are finalized by gardenlet in a later step.
func (c *Component) Destroy(ctx context.Context, forceDelete bool) error {
	if err := managedresources.DeleteForShoot(ctx, c.client, c.values.Namespace, constants.ManagedResourceNameShoot); err != nil {
		return err
	}

	if !forceDelete {
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		if err := managedresources.WaitUntilDeleted(timeoutCtx, c.client, c.values.Namespace, constants.ManagedResourceNameShoot); err != nil {
			return err
		}
	}

	if err := managedresources.DeleteForSeed(ctx, c.client, c.values.Namespace, constants.ManagedResourceNameSeed); err != nil {
		return err
	}

	if !forceDelete {
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		return managedresources.WaitUntilDeleted(timeoutCtx, c.client, c.values.Namespace, constants.ManagedResourceNameSeed)
	}

	return nil
}

// Wait waits until both ManagedResources are healthy.
func (c *Component) Wait(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := managedresources.WaitUntilHealthy(timeoutCtx, c.client, c.values.Namespace, constants.ManagedResourceNameSeed); err != nil {
		return fmt.Errorf("waiting for seed ManagedResource to be healthy: %w", err)
	}

	return managedresources.WaitUntilHealthy(timeoutCtx, c.client, c.values.Namespace, constants.ManagedResourceNameShoot)
}

func (c *Component) seedResources() (map[string][]byte, error) {
	registry := managedresources.NewRegistry(gardenerkubernetes.SeedScheme, gardenerkubernetes.SeedCodec, gardenerkubernetes.SeedSerializer)

	deployment := c.deployment()
	if err := gutil.InjectGenericKubeconfig(deployment, c.values.GenericTokenKubeconfigSecretName, c.values.OperatorShootAccessSecretName); err != nil {
		return nil, fmt.Errorf("failed to inject generic kubeconfig: %w", err)
	}

	service := c.service()
	if err := gutil.InjectNetworkPolicyAnnotationsForScrapeTargets(service, networkingv1.NetworkPolicyPort{
		Port:     ptr.To(intstr.FromInt32(portMetrics)),
		Protocol: ptr.To(corev1.ProtocolTCP),
	}); err != nil {
		return nil, fmt.Errorf("failed to inject network policy annotations: %w", err)
	}
	if err := gutil.InjectNetworkPolicyAnnotationsForWebhookTargets(service, networkingv1.NetworkPolicyPort{
		Port:     ptr.To(intstr.FromInt32(portWebhook)),
		Protocol: ptr.To(corev1.ProtocolTCP),
	}); err != nil {
		return nil, fmt.Errorf("failed to inject webhook network policy annotations: %w", err)
	}

	configMap, err := c.configMap()
	if err != nil {
		return nil, err
	}

	return registry.AddAllAndSerialize(
		c.serviceAccount(),
		c.runServiceAccount(),
		configMap,
		deployment,
		service,
		c.serviceMonitor(),
		c.verticalPodAutoscaler(),
		c.role(),
		c.roleBinding(),
	)
}

func (c *Component) shootResources() (map[string][]byte, error) {
	registry := managedresources.NewRegistry(gardenerkubernetes.ShootScheme, gardenerkubernetes.ShootCodec, gardenerkubernetes.ShootSerializer)

	resources, err := registry.AddAllAndSerialize(
		c.operatorClusterRole(),
		c.operatorClusterRoleBinding(),
		c.scannerClusterRole(),
		c.scannerClusterRoleBinding(),
		c.scannerRole(),
		c.scannerRoleBinding(),
		c.exporterClusterRole(),
		c.exporterClusterRoleBinding(),
		c.roleLeaderElection(),
		c.roleBindingLeaderElection(),
		c.validatingWebhookConfiguration(),
		c.mutatingWebhookConfiguration(),
	)
	if err != nil {
		return nil, err
	}

	resources["crd-compliancescans.yaml"] = crdComplianceScans
	resources["crd-reportoutputs.yaml"] = crdReportOutputs
	resources["crd-scheduledcompliancescans.yaml"] = crdScheduledComplianceScans

	return resources, nil
}

func (c *Component) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": constants.DikiOperatorName,
	}
}

func (c *Component) haLabels() map[string]string {
	return map[string]string{
		resourcesv1alpha1.HighAvailabilityConfigType: resourcesv1alpha1.HighAvailabilityConfigTypeServer,
	}
}
