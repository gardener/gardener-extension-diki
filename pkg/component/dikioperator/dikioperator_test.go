// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator_test

import (
	"context"
	"fmt"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	kubeapiserverconstants "github.com/gardener/gardener/pkg/component/kubernetes/apiserver/constants"
	monitoringutils "github.com/gardener/gardener/pkg/component/observability/monitoring/utils"
	"github.com/gardener/gardener/pkg/utils"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/retry"
	retryfake "github.com/gardener/gardener/pkg/utils/retry/fake"
	"github.com/gardener/gardener/pkg/utils/test"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener-extension-diki/pkg/component/dikioperator"
	"github.com/gardener/gardener-extension-diki/pkg/constants"
)

var _ = Describe("Component", func() {
	const (
		namespace                  = "shoot--foo--bar"
		image                      = "europe-docker.pkg.dev/gardener-project/releases/gardener/diki-operator:v0.1.0"
		genericTokenKubeconfigName = "generic-token-kubeconfig"
		operatorShootAccessSecret  = "shoot-access-diki-operator"
		operatorShootAccessSAName  = "diki-operator"
		runnerShootAccessSecret    = "shoot-access-diki-runner"
		runnerShootAccessSAName    = "diki-runner"
		serverTLSSecretName        = "diki-operator-tls-12345" //#nosec G101 - not credentials
	)

	var (
		ctx        context.Context
		fakeClient client.Client
		comp       *dikioperator.Component
		values     dikioperator.Values

		fakeOps   *retryfake.Ops
		consistOf func(...client.Object) types.GomegaMatcher

		managedResourceSeed  *resourcesv1alpha1.ManagedResource
		managedResourceShoot *resourcesv1alpha1.ManagedResource

		webhookCABundle = []byte("ca-bundle-data")
	)

	BeforeEach(func() {
		ctx = context.Background()

		fakeClient = fakeclient.NewClientBuilder().WithScheme(gardenerkubernetes.SeedScheme).Build()

		fakeOps = &retryfake.Ops{MaxAttempts: 2}
		DeferCleanup(test.WithVars(
			&retry.Until, fakeOps.Until,
			&retry.UntilTimeout, fakeOps.UntilTimeout,
		))

		consistOf = NewManagedResourceConsistOfObjectsMatcher(fakeClient,
			cmpopts.IgnoreFields(corev1.ConfigMap{}, "Data"),
			cmpopts.IgnoreFields(apiextensionsv1.CustomResourceDefinition{}, "Spec", "Annotations"),
		)

		values = dikioperator.Values{
			Image:                                 image,
			Replicas:                              1,
			Namespace:                             namespace,
			GenericTokenKubeconfigSecretName:      genericTokenKubeconfigName,
			OperatorShootAccessSecretName:         operatorShootAccessSecret,
			OperatorShootAccessServiceAccountName: operatorShootAccessSAName,
			RunnerShootAccessSecretName:           runnerShootAccessSecret,
			RunnerShootAccessServiceAccountName:   runnerShootAccessSAName,
			ServerTLSSecretName:                   serverTLSSecretName,
			WebhookCABundle:                       webhookCABundle,
		}

		managedResourceSeed = &resourcesv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ManagedResourceNameSeed,
				Namespace: namespace,
			},
		}
		managedResourceShoot = &resourcesv1alpha1.ManagedResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ManagedResourceNameShoot,
				Namespace: namespace,
			},
		}
	})

	JustBeforeEach(func() {
		comp = dikioperator.New(fakeClient, values)
	})

	Describe("#Deploy", func() {
		Context("seed resources", func() {
			JustBeforeEach(func() {
				Expect(comp.Deploy(ctx)).To(Succeed())
				Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(managedResourceSeed), managedResourceSeed)).To(Succeed())
			})

			It("should deploy all seed resources", func() {
				Expect(managedResourceSeed).To(consistOf(
					expectedServiceAccount(),
					expectedRunServiceAccount(),
					expectedConfigMap(),
					expectedDeployment(1),
					expectedService(),
					expectedServiceMonitor(),
					expectedVPA(),
					expectedRole(),
					expectedRoleBinding(),
				))
			})

			It("should set replicas to 0 when hibernated", func() {
				values.Replicas = 0
				comp = dikioperator.New(fakeClient, values)

				Expect(comp.Deploy(ctx)).To(Succeed())
				Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(managedResourceSeed), managedResourceSeed)).To(Succeed())

				Expect(managedResourceSeed).To(consistOf(
					expectedServiceAccount(),
					expectedRunServiceAccount(),
					expectedConfigMap(),
					expectedDeployment(0),
					expectedService(),
					expectedServiceMonitor(),
					expectedVPA(),
					expectedRole(),
					expectedRoleBinding(),
				))
			})
		})

		Context("shoot resources", func() {
			JustBeforeEach(func() {
				Expect(comp.Deploy(ctx)).To(Succeed())
				Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(managedResourceShoot), managedResourceShoot)).To(Succeed())
			})

			It("should deploy all shoot RBAC and webhook resources", func() {
				Expect(managedResourceShoot).To(consistOf(
					expectedOperatorClusterRole(),
					expectedOperatorClusterRoleBinding(),
					expectedScannerClusterRole(),
					expectedScannerClusterRoleBinding(),
					expectedScannerRole(),
					expectedScannerRoleBinding(),
					expectedExporterClusterRole(),
					expectedExporterClusterRoleBinding(),
					expectedLeaderElectionRole(),
					expectedLeaderElectionRoleBinding(),
					expectedValidatingWebhookConfiguration(),
					expectedMutatingWebhookConfiguration(),
					&apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "compliancescans.diki.gardener.cloud"}},
					&apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "reportoutputs.diki.gardener.cloud"}},
					&apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "scheduledcompliancescans.diki.gardener.cloud"}},
				))
			})
		})
	})

	Describe("#Destroy", func() {
		It("should delete both ManagedResources", func() {
			Expect(fakeClient.Create(ctx, managedResourceSeed)).To(Succeed())
			Expect(fakeClient.Create(ctx, managedResourceShoot)).To(Succeed())

			comp = dikioperator.New(fakeClient, values)
			Expect(comp.Destroy(ctx)).To(Succeed())

			Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(managedResourceSeed), managedResourceSeed)).To(BeNotFoundError())
			Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(managedResourceShoot), managedResourceShoot)).To(BeNotFoundError())
		})
	})

	Describe("#Wait", func() {
		It("should fail because the seed ManagedResource is not healthy", func() {
			Expect(fakeClient.Create(ctx, &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:       constants.ManagedResourceNameSeed,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: unhealthyManagedResourceStatus,
			})).To(Succeed())

			comp = dikioperator.New(fakeClient, values)
			Expect(comp.Wait(ctx)).To(MatchError(ContainSubstring("is not healthy")))
		})
	})

	Describe("#WaitCleanup", func() {
		It("should fail when the managed resource still exists", func() {
			Expect(fakeClient.Create(ctx, managedResourceSeed)).To(Succeed())

			comp = dikioperator.New(fakeClient, values)
			Expect(comp.WaitCleanup(ctx)).To(MatchError(ContainSubstring("still exists")))
		})

		It("should succeed when managed resources are gone", func() {
			comp = dikioperator.New(fakeClient, values)
			Expect(comp.WaitCleanup(ctx)).To(Succeed())
		})
	})
})

// Expected seed resources

func expectedServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
}

func expectedRunServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "diki-run",
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
}

func expectedConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
	}
}

func expectedDeployment(replicas int32) *appsv1.Deployment {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels: utils.MergeStringMaps(
				map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
				map[string]string{resourcesv1alpha1.HighAvailabilityConfigType: resourcesv1alpha1.HighAvailabilityConfigTypeServer},
			),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             ptr.To(replicas),
			RevisionHistoryLimit: ptr.To[int32](2),
			Selector:             &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":                              constants.DikiOperatorServiceName,
						v1beta1constants.LabelNetworkPolicyToDNS:              v1beta1constants.LabelNetworkPolicyAllowed,
						v1beta1constants.LabelNetworkPolicyToRuntimeAPIServer: v1beta1constants.LabelNetworkPolicyAllowed,
						gutil.NetworkPolicyLabel(v1beta1constants.DeploymentNameKubeAPIServer, kubeapiserverconstants.Port): v1beta1constants.LabelNetworkPolicyAllowed,
					},
				},
				Spec: corev1.PodSpec{
					PriorityClassName:  v1beta1constants.PriorityClassNameShootControlPlane300,
					ServiceAccountName: constants.DikiOperatorServiceName,
					SecurityContext: &corev1.PodSecurityContext{
						SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
						RunAsNonRoot:   ptr.To(true),
						RunAsUser:      ptr.To[int64](65532),
						RunAsGroup:     ptr.To[int64](65532),
						FSGroup:        ptr.To[int64](65532),
					},
					AutomountServiceAccountToken: ptr.To(true),
					Containers: []corev1.Container{{
						Name:            constants.DikiOperatorServiceName,
						Image:           "europe-docker.pkg.dev/gardener-project/releases/gardener/diki-operator:v0.1.0",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Args: []string{
							"--config=/etc/diki-operator/config/config.yaml",
							fmt.Sprintf("--kubeconfig=%s/kubeconfig", gutil.VolumeMountPathGenericKubeconfig),
						},
						Ports: []corev1.ContainerPort{
							{Name: "health", ContainerPort: 8081, Protocol: corev1.ProtocolTCP},
							{Name: "metrics", ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
							{Name: "webhooks", ContainerPort: 10443, Protocol: corev1.ProtocolTCP},
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
							ReadOnlyRootFilesystem:   ptr.To(true),
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromInt32(8081), Scheme: corev1.URISchemeHTTP}},
							InitialDelaySeconds: 10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/readyz", Port: intstr.FromInt32(8081), Scheme: corev1.URISchemeHTTP}},
							InitialDelaySeconds: 5,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "diki-operator-config", MountPath: "/etc/diki-operator/config", ReadOnly: true},
							{Name: "webhook-tls", MountPath: "/etc/diki-operator/webhooks/tls", ReadOnly: true},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "diki-operator-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: constants.DikiOperatorServiceName},
									DefaultMode:          ptr.To[int32](0440),
								},
							},
						},
						{
							Name: "webhook-tls",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "diki-operator-tls-12345", //#nosec G101 - not credentials
									DefaultMode: ptr.To[int32](0440),
								},
							},
						},
					},
				},
			},
		},
	}
	Expect(gutil.InjectGenericKubeconfig(dep, "generic-token-kubeconfig", "shoot-access-diki-operator")).To(Succeed())
	return dep
}

func expectedService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
			Annotations: map[string]string{
				"networking.resources.gardener.cloud/from-all-scrape-targets-allowed-ports":  `[{"protocol":"TCP","port":8080}]`,
				"networking.resources.gardener.cloud/from-all-webhook-targets-allowed-ports": `[{"protocol":"TCP","port":10443}]`,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
			Ports: []corev1.ServicePort{
				{Name: "metrics", Port: 8080, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(8080)},
				{Name: "webhooks", Port: 443, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(10443)},
			},
		},
	}
}

func expectedServiceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: monitoringutils.ConfigObjectMeta(constants.DikiOperatorServiceName, "shoot--foo--bar", "shoot"),
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
			Endpoints: []monitoringv1.Endpoint{{
				Port: "metrics",
			}},
		},
	}
}

func expectedVPA() *vpaautoscalingv1.VerticalPodAutoscaler {
	return &vpaautoscalingv1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName + "-vpa",
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
		Spec: vpaautoscalingv1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "Deployment",
				Name:       constants.DikiOperatorServiceName,
			},
			UpdatePolicy: &vpaautoscalingv1.PodUpdatePolicy{
				UpdateMode: ptr.To(vpaautoscalingv1.UpdateModeInPlaceOrRecreate),
			},
			ResourcePolicy: &vpaautoscalingv1.PodResourcePolicy{
				ContainerPolicies: []vpaautoscalingv1.ContainerResourcePolicy{
					{
						ContainerName:    constants.DikiOperatorServiceName,
						ControlledValues: ptr.To(vpaautoscalingv1.ContainerControlledValuesRequestsOnly),
						MinAllowed:       corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("32Mi")},
					},
					{
						ContainerName: vpaautoscalingv1.DefaultContainerResourcePolicy,
						Mode:          ptr.To(vpaautoscalingv1.ContainerScalingModeOff),
					},
				},
			},
		},
	}
}

func expectedRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"configmaps"}, Verbs: []string{"create"}},
			{APIGroups: []string{"batch"}, Resources: []string{"jobs"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		},
	}
}

func expectedRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorServiceName,
			Namespace: "shoot--foo--bar",
			Labels:    map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName},
		},
		RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: constants.DikiOperatorServiceName},
		Subjects: []rbacv1.Subject{
			{Kind: rbacv1.ServiceAccountKind, Name: constants.DikiOperatorServiceName, Namespace: "shoot--foo--bar"},
		},
	}
}

// Expected shoot resources

func expectedOperatorClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "diki-operator", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"diki.gardener.cloud"}, Resources: []string{"compliancescans"}, Verbs: []string{"create", "delete"}},
			{APIGroups: []string{"diki.gardener.cloud"}, Resources: []string{"compliancescans", "compliancescans/status", "scheduledcompliancescans", "scheduledcompliancescans/status"}, Verbs: []string{"get", "list", "watch", "update", "patch"}},
			{APIGroups: []string{""}, Resources: []string{"configmaps"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{"diki.gardener.cloud"}, Resources: []string{"reportoutputs"}, Verbs: []string{"get", "list", "watch"}},
		},
	}
}

func expectedOperatorClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "diki-operator", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "diki-operator"},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: "diki-operator", Namespace: metav1.NamespaceSystem}},
	}
}

func expectedScannerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner.diki.gardener.cloud", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"configmaps", "nodes", "nodes/proxy", "namespaces", "pods", "replicationcontrollers", "services"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"apps"}, Resources: []string{"daemonsets", "deployments", "replicasets", "statefulsets"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"batch"}, Resources: []string{"jobs", "cronjobs"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"autoscaling"}, Resources: []string{"horizontalpodautoscalers"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"storage.k8s.io"}, Resources: []string{"storageclasses"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"networking.k8s.io"}, Resources: []string{"networkpolicies"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"roles", "clusterroles"}, Verbs: []string{"get", "list"}},
		},
	}
}

func expectedScannerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "diki-scanner", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "scanner.diki.gardener.cloud"},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: "diki-runner", Namespace: metav1.NamespaceSystem}},
	}
}

func expectedScannerRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "scanner.diki.gardener.cloud", Namespace: metav1.NamespaceSystem, Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"create", "delete"}},
			{APIGroups: []string{""}, Resources: []string{"pods/exec"}, Verbs: []string{"create"}},
		},
	}
}

func expectedScannerRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "diki-scanner", Namespace: metav1.NamespaceSystem, Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "scanner.diki.gardener.cloud"},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: "diki-runner", Namespace: metav1.NamespaceSystem}},
	}
}

func expectedExporterClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "exporter.diki.gardener.cloud", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"diki.gardener.cloud"}, Resources: []string{"compliancescans"}, Verbs: []string{"get"}},
			{APIGroups: []string{"diki.gardener.cloud"}, Resources: []string{"compliancescans/status"}, Verbs: []string{"get", "patch"}},
			{APIGroups: []string{""}, Resources: []string{"configmaps"}, Verbs: []string{"create"}},
		},
	}
}

func expectedExporterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "diki-exporter", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "exporter.diki.gardener.cloud"},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: "diki-runner", Namespace: metav1.NamespaceSystem}},
	}
}

func expectedLeaderElectionRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: constants.DikiOperatorServiceName + "-leader-election", Namespace: metav1.NamespaceSystem, Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{"", "events.k8s.io"}, Resources: []string{"events"}, Verbs: []string{"create", "update", "patch"}},
			{APIGroups: []string{"coordination.k8s.io"}, Resources: []string{"leases"}, Verbs: []string{"get", "create", "update"}},
		},
	}
}

func expectedLeaderElectionRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: constants.DikiOperatorServiceName + "-leader-election", Namespace: metav1.NamespaceSystem, Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: constants.DikiOperatorServiceName + "-leader-election"},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: "diki-operator", Namespace: metav1.NamespaceSystem}},
	}
}

func expectedValidatingWebhookConfiguration() *admissionregistrationv1.ValidatingWebhookConfiguration {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Equivalent
	sideEffects := admissionregistrationv1.SideEffectClassNone

	return &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "gardener-extension-diki", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name:                    "compliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule:       admissionregistrationv1.Rule{APIGroups: []string{"diki.gardener.cloud"}, APIVersions: []string{"v1alpha1"}, Resources: []string{"compliancescans"}},
				}},
				ClientConfig:   admissionregistrationv1.WebhookClientConfig{URL: ptr.To(fmt.Sprintf("https://%s.%s.svc/webhooks/compliancescan", constants.DikiOperatorServiceName, "shoot--foo--bar")), CABundle: []byte("ca-bundle-data")},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: ptr.To[int32](10),
			},
			{
				Name:                    "scheduledcompliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule:       admissionregistrationv1.Rule{APIGroups: []string{"diki.gardener.cloud"}, APIVersions: []string{"v1alpha1"}, Resources: []string{"scheduledcompliancescans"}},
				}},
				ClientConfig:   admissionregistrationv1.WebhookClientConfig{URL: ptr.To(fmt.Sprintf("https://%s.%s.svc/webhooks/scheduledcompliancescan/validate", constants.DikiOperatorServiceName, "shoot--foo--bar")), CABundle: []byte("ca-bundle-data")},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: ptr.To[int32](10),
			},
		},
	}
}

func expectedMutatingWebhookConfiguration() *admissionregistrationv1.MutatingWebhookConfiguration {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Equivalent
	sideEffects := admissionregistrationv1.SideEffectClassNone

	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "gardener-extension-diki", Labels: map[string]string{"app.kubernetes.io/name": constants.DikiOperatorServiceName}},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name:                    "scheduledcompliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule:       admissionregistrationv1.Rule{APIGroups: []string{"diki.gardener.cloud"}, APIVersions: []string{"v1alpha1"}, Resources: []string{"scheduledcompliancescans"}},
				}},
				ClientConfig:   admissionregistrationv1.WebhookClientConfig{URL: ptr.To(fmt.Sprintf("https://%s.%s.svc/webhooks/scheduledcompliancescan/mutate", constants.DikiOperatorServiceName, "shoot--foo--bar")), CABundle: []byte("ca-bundle-data")},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: ptr.To[int32](10),
			},
		},
	}
}

// Status helpers

var unhealthyManagedResourceStatus = resourcesv1alpha1.ManagedResourceStatus{
	ObservedGeneration: 1,
	Conditions:         []gardencorev1beta1.Condition{{Type: resourcesv1alpha1.ResourcesApplied, Status: gardencorev1beta1.ConditionFalse}},
}
