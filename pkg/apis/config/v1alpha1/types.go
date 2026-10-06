// SPDX-FileCopyrightText: Contributors to the Gardener project
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

	// BaseDikiConfig is the YAML content for base diki options.
	// +optional
	BaseDikiConfig *string `json:"baseDikiConfig,omitempty"`
	// DefaultScheduledScan configures a default ScheduledComplianceScan that is injected into
	// every shoot cluster that has the diki extension enabled.
	// +optional
	DefaultScheduledScan *DefaultScheduledScanSpec `json:"defaultScheduledScan,omitempty"`
}

// DefaultScheduledScanSpec describes the spec of the default ScheduledComplianceScan.
type DefaultScheduledScanSpec struct {
	// Schedule is a cron expression for how often the scan should run.
	// If the shoot has a maintenance time window configured, the hour and minute fields of this
	// expression are replaced with values derived from the window's begin time; the remaining
	// fields (day-of-month, month, day-of-week) are preserved as configured.
	Schedule string `json:"schedule"`
	// SuccessfulScansHistoryLimit is the number of successfully completed scans to retain.
	// +optional
	SuccessfulScansHistoryLimit *int32 `json:"successfulScansHistoryLimit,omitempty"`
	// FailedScansHistoryLimit is the number of failed scans to retain.
	// +optional
	FailedScansHistoryLimit *int32 `json:"failedScansHistoryLimit,omitempty"`
	// Rulesets describes the rulesets to be applied during each scheduled scan.
	Rulesets []RulesetConfig `json:"rulesets"`
}

// RulesetConfig describes the configuration of a ruleset.
type RulesetConfig struct {
	// ID is the identifier of the ruleset.
	ID string `json:"id"`
	// Version is the version of the ruleset.
	Version string `json:"version"`
}
