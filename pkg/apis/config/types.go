// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the diki extension configuration.
type Configuration struct {
	metav1.TypeMeta

	// BaseDikiConfig is the YAML content for base diki options.
	// +optional
	BaseDikiConfig *string
	// DefaultScheduledScan configures a default ScheduledComplianceScan that is injected into
	// every shoot cluster that has the diki extension enabled.
	// +optional
	DefaultScheduledScan *DefaultScheduledScanSpec
}

// DefaultScheduledScanSpec describes the spec of the default ScheduledComplianceScan.
type DefaultScheduledScanSpec struct {
	// Schedule is a cron expression for how often the scan should run.
	// If the shoot has a maintenance time window configured, the hour and minute fields of this
	// expression are replaced with values derived from the window's begin time; the remaining
	// fields (day-of-month, month, day-of-week) are preserved as configured.
	Schedule string
	// SuccessfulScansHistoryLimit is the number of successfully completed scans to retain.
	// +optional
	SuccessfulScansHistoryLimit *int32
	// FailedScansHistoryLimit is the number of failed scans to retain.
	// +optional
	FailedScansHistoryLimit *int32
	// Rulesets describes the rulesets to be applied during each scheduled scan.
	Rulesets []RulesetConfig
}

// RulesetConfig describes the configuration of a ruleset.
type RulesetConfig struct {
	// ID is the identifier of the ruleset.
	ID string
	// Version is the version of the ruleset.
	Version string
}
