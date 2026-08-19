// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator

import (
	"encoding/json"
	"fmt"

	dikiconfigv1alpha1 "github.com/gardener/diki-operator/pkg/apis/config/v1alpha1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	kubeapiserverconstants "github.com/gardener/gardener/pkg/component/kubernetes/apiserver/constants"
	monitoringutils "github.com/gardener/gardener/pkg/component/observability/monitoring/utils"
	"github.com/gardener/gardener/pkg/utils"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.yaml.in/yaml/v4"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-diki/pkg/constants"
)

func (c *Component) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
}

func (c *Component) configMap() (*corev1.ConfigMap, error) {
	config, err := c.operatorConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate operator config: %w", err)
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		Data: map[string]string{
			configFileName: config,
		},
	}, nil
}

func (c *Component) operatorConfig() (string, error) {
	cfg := &dikiconfigv1alpha1.DikiOperatorConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: dikiconfigv1alpha1.SchemeGroupVersion.String(),
			Kind:       "DikiOperatorConfiguration",
		},
		Log: dikiconfigv1alpha1.Log{
			Level:  "info",
			Format: "json",
		},
		LeaderElection: &componentbaseconfigv1alpha1.LeaderElectionConfiguration{
			LeaderElect:       ptr.To(true),
			ResourceLock:      "leases",
			ResourceName:      dikiconfigv1alpha1.DefaultLockObjectName,
			ResourceNamespace: dikiconfigv1alpha1.DefaultLockObjectNamespace,
		},
		Controllers: dikiconfigv1alpha1.ControllerConfiguration{
			ComplianceScan: dikiconfigv1alpha1.ComplianceScanConfig{
				DikiRunner: dikiconfigv1alpha1.DikiRunnerConfig{
					Namespace: c.values.Namespace,
					TargetKubeconfig: &dikiconfigv1alpha1.KubeconfigConfig{
						SecretRef: dikiconfigv1alpha1.SecretRef{
							Name: c.values.GenericTokenKubeconfigSecretName,
							Key:  ptr.To("kubeconfig"),
						},
						TokenSecretRef: &dikiconfigv1alpha1.SecretRef{
							Name: c.values.RunnerShootAccessSecretName,
							Key:  ptr.To("token"),
						},
						MountPath: gutil.VolumeMountPathGenericKubeconfig,
					},
				},
			},
		},
		Server: dikiconfigv1alpha1.ServerConfiguration{
			HealthProbes: &dikiconfigv1alpha1.Server{Port: portHealth},
			Metrics:      &dikiconfigv1alpha1.Server{Port: portMetrics},
			Webhooks: dikiconfigv1alpha1.HTTPSServer{
				Server: dikiconfigv1alpha1.Server{Port: portWebhook},
				TLS: dikiconfigv1alpha1.TLS{
					ServerCertDir: webhookTLSMountPath,
				},
			},
		},
	}

	// Marshal to JSON first because go.yaml.in/yaml/v4 ignores json: struct tags
	// and the config types only have json: tags.
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal operator config to JSON: %w", err)
	}

	var raw any
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return "", fmt.Errorf("failed to unmarshal operator config JSON: %w", err)
	}

	yamlData, err := yaml.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal operator config to YAML: %w", err)
	}

	return string(yamlData), nil
}

func (c *Component) service() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: c.labels(),
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       portMetrics,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(portMetrics),
				},
				{
					Name:       "webhooks",
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(portWebhook),
				},
			},
		},
	}
}

func (c *Component) serviceMonitor() *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: monitoringutils.ConfigObjectMeta(constants.DikiOperatorName, c.values.Namespace, "shoot"),
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{MatchLabels: c.labels()},
			Endpoints: []monitoringv1.Endpoint{{
				Port: "metrics",
			}},
		},
	}
}

func (c *Component) deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    utils.MergeStringMaps(c.labels(), c.haLabels()),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             ptr.To(c.values.Replicas),
			RevisionHistoryLimit: ptr.To[int32](2),
			Selector:             &metav1.LabelSelector{MatchLabels: c.labels()},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: c.podLabels(),
				},
				Spec: corev1.PodSpec{
					PriorityClassName:  v1beta1constants.PriorityClassNameShootControlPlane300,
					ServiceAccountName: constants.DikiOperatorName,
					SecurityContext: &corev1.PodSecurityContext{
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
						RunAsNonRoot: ptr.To(true),
						RunAsUser:    ptr.To[int64](65532),
						RunAsGroup:   ptr.To[int64](65532),
						FSGroup:      ptr.To[int64](65532),
					},
					AutomountServiceAccountToken: ptr.To(true),
					Containers: []corev1.Container{{
						Name:            constants.DikiOperatorName,
						Image:           c.values.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Args: []string{
							fmt.Sprintf("--config=%s/%s", configMountPath, configFileName),
							fmt.Sprintf("--kubeconfig=%s/kubeconfig", gutil.VolumeMountPathGenericKubeconfig),
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "health",
								ContainerPort: portHealth,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "metrics",
								ContainerPort: portMetrics,
								Protocol:      corev1.ProtocolTCP,
							},
							{
								Name:          "webhooks",
								ContainerPort: portWebhook,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
							ReadOnlyRootFilesystem: ptr.To(true),
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/healthz",
									Port:   intstr.FromInt32(portHealth),
									Scheme: corev1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: 10,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/readyz",
									Port:   intstr.FromInt32(portHealth),
									Scheme: corev1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: 5,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      configVolumeName,
								MountPath: configMountPath,
								ReadOnly:  true,
							},
							{
								Name:      webhookTLSVolumeName,
								MountPath: webhookTLSMountPath,
								ReadOnly:  true,
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: configVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: constants.DikiOperatorName,
									},
									DefaultMode: ptr.To[int32](0440),
								},
							},
						},
						{
							Name: webhookTLSVolumeName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  c.values.ServerTLSSecretName,
									DefaultMode: ptr.To[int32](0440),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Component) role() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		},
	}
}

func (c *Component) roleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     constants.DikiOperatorName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      constants.DikiOperatorName,
				Namespace: c.values.Namespace,
			},
		},
	}
}

func (c *Component) runServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "diki-run",
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
}

func (c *Component) verticalPodAutoscaler() *vpaautoscalingv1.VerticalPodAutoscaler {
	return &vpaautoscalingv1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName,
			Namespace: c.values.Namespace,
			Labels:    c.labels(),
		},
		Spec: vpaautoscalingv1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "Deployment",
				Name:       constants.DikiOperatorName,
			},
			UpdatePolicy: &vpaautoscalingv1.PodUpdatePolicy{
				UpdateMode: ptr.To(vpaautoscalingv1.UpdateModeInPlaceOrRecreate),
			},
			ResourcePolicy: &vpaautoscalingv1.PodResourcePolicy{
				ContainerPolicies: []vpaautoscalingv1.ContainerResourcePolicy{
					{
						ContainerName:    constants.DikiOperatorName,
						ControlledValues: ptr.To(vpaautoscalingv1.ContainerControlledValuesRequestsOnly),
						MinAllowed: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("32Mi"),
						},
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

func (c *Component) podLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":                              constants.DikiOperatorName,
		v1beta1constants.LabelNetworkPolicyToDNS:              v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToRuntimeAPIServer: v1beta1constants.LabelNetworkPolicyAllowed,
		gutil.NetworkPolicyLabel(v1beta1constants.DeploymentNameKubeAPIServer, kubeapiserverconstants.Port): v1beta1constants.LabelNetworkPolicyAllowed,
	}
}
