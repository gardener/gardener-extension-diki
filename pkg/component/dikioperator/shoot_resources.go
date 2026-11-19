// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator

import (
	"encoding/json"
	"fmt"

	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	"go.yaml.in/yaml/v4"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-diki/pkg/constants"
)

func (c *Component) operatorClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "diki-operator",
			Labels: c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"diki.gardener.cloud"},
				Resources: []string{"compliancescans"},
				Verbs:     []string{"create", "delete"},
			},
			{
				APIGroups: []string{"diki.gardener.cloud"},
				Resources: []string{"compliancescans", "compliancescans/status", "scheduledcompliancescans", "scheduledcompliancescans/status"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"diki.gardener.cloud"},
				Resources: []string{"reportoutputs"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func (c *Component) operatorClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "diki-operator",
			Labels: c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "diki-operator",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.values.OperatorShootAccessServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func (c *Component) scannerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "scanner.diki.gardener.cloud",
			Labels: c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "nodes", "nodes/proxy", "namespaces", "pods", "replicationcontrollers", "services"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"daemonsets", "deployments", "replicasets", "statefulsets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"autoscaling"},
				Resources: []string{"horizontalpodautoscalers"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"networkpolicies"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"roles", "clusterroles"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
}

func (c *Component) scannerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "diki-scanner",
			Labels: c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "scanner.diki.gardener.cloud",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.values.RunnerShootAccessServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func (c *Component) scannerRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scanner.diki.gardener.cloud",
			Namespace: metav1.NamespaceSystem,
			Labels:    c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/exec"},
				Verbs:     []string{"create"},
			},
		},
	}
}

func (c *Component) scannerRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "diki-scanner",
			Namespace: metav1.NamespaceSystem,
			Labels:    c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "scanner.diki.gardener.cloud",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.values.RunnerShootAccessServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func (c *Component) exporterClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "exporter.diki.gardener.cloud",
			Labels: c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"diki.gardener.cloud"},
				Resources: []string{"compliancescans"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"diki.gardener.cloud"},
				Resources: []string{"compliancescans/status"},
				Verbs:     []string{"get", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
		},
	}
}

func (c *Component) exporterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "diki-exporter",
			Labels: c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "exporter.diki.gardener.cloud",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.values.RunnerShootAccessServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func (c *Component) roleLeaderElection() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName + "-leader-election",
			Namespace: metav1.NamespaceSystem,
			Labels:    c.labels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"", "events.k8s.io"},
				Resources: []string{"events"},
				Verbs:     []string{"create", "update", "patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "create", "update"},
			},
		},
	}
}

func (c *Component) roleBindingLeaderElection() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.DikiOperatorName + "-leader-election",
			Namespace: metav1.NamespaceSystem,
			Labels:    c.labels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     constants.DikiOperatorName + "-leader-election",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.values.OperatorShootAccessServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func (c *Component) validatingWebhookConfiguration() *admissionregistrationv1.ValidatingWebhookConfiguration {
	var (
		failurePolicy  = admissionregistrationv1.Fail
		matchPolicy    = admissionregistrationv1.Equivalent
		sideEffects    = admissionregistrationv1.SideEffectClassNone
		timeoutSeconds = ptr.To[int32](10)
		webhookURL     = fmt.Sprintf("https://%s.%s.svc", constants.DikiOperatorName, c.values.Namespace)
	)

	return &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gardener-extension-diki",
			Labels: c.labels(),
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name:                    "compliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"diki.gardener.cloud"},
						APIVersions: []string{"v1alpha1"},
						Resources:   []string{"compliancescans"},
					},
				}},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(webhookURL + "/webhooks/compliancescan"),
					CABundle: c.values.WebhookCABundle,
				},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: timeoutSeconds,
			},
			{
				Name:                    "scheduledcompliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"diki.gardener.cloud"},
						APIVersions: []string{"v1alpha1"},
						Resources:   []string{"scheduledcompliancescans"},
					},
				}},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(webhookURL + "/webhooks/scheduledcompliancescan/validate"),
					CABundle: c.values.WebhookCABundle,
				},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: timeoutSeconds,
			},
		},
	}
}

func (c *Component) mutatingWebhookConfiguration() *admissionregistrationv1.MutatingWebhookConfiguration {
	var (
		failurePolicy  = admissionregistrationv1.Fail
		matchPolicy    = admissionregistrationv1.Equivalent
		sideEffects    = admissionregistrationv1.SideEffectClassNone
		timeoutSeconds = ptr.To[int32](10)
		webhookURL     = fmt.Sprintf("https://%s.%s.svc", constants.DikiOperatorName, c.values.Namespace)
	)

	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gardener-extension-diki",
			Labels: c.labels(),
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name:                    "scheduledcompliancescans.diki.gardener.cloud",
				AdmissionReviewVersions: []string{"v1"},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"diki.gardener.cloud"},
						APIVersions: []string{"v1alpha1"},
						Resources:   []string{"scheduledcompliancescans"},
					},
				}},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					URL:      ptr.To(webhookURL + "/webhooks/scheduledcompliancescan/mutate"),
					CABundle: c.values.WebhookCABundle,
				},
				FailurePolicy:  &failurePolicy,
				MatchPolicy:    &matchPolicy,
				SideEffects:    &sideEffects,
				TimeoutSeconds: timeoutSeconds,
			},
		},
	}
}

func (c *Component) defaultScheduledComplianceScanYAML() ([]byte, error) {
	scan := &dikiv1alpha1.ScheduledComplianceScan{
		TypeMeta: metav1.TypeMeta{
			APIVersion: dikiv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ScheduledComplianceScan",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			// This annotation is needed since the ComplianceScanSpec is immutable for ScheduledComplianceScans.
			Annotations: map[string]string{"resources.gardener.cloud/delete-on-invalid-update": "true"},
		},
		Spec: *c.values.DefaultScheduledScanSpec,
	}

	// Marshal to JSON first because go.yaml.in/yaml/v4 ignores json: struct tags
	// and the diki-operator types only have json: tags.
	jsonData, err := json.Marshal(scan)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ScheduledComplianceScan to JSON: %w", err)
	}
	var raw any
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ScheduledComplianceScan JSON: %w", err)
	}
	yamlData, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ScheduledComplianceScan to YAML: %w", err)
	}
	return yamlData, nil
}
