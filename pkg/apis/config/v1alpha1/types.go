// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the diki extension configuration.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`

	// BaseDikiOptions holds the base diki options configuration.
	// +optional
	BaseDikiOptions *BaseDikiOptionsConfig `json:"baseDikiOptions,omitempty"`
}

// BaseDikiOptionsConfig holds the base diki options content.
type BaseDikiOptionsConfig struct {
	// Data is the raw diki configuration YAML content.
	Data string `json:"data"`
}
