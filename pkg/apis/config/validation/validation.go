// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-diki/pkg/apis/config"
)

// ValidateConfiguration validates the passed configuration instance.
func ValidateConfiguration(cfg *config.Configuration) field.ErrorList {
	allErrs := field.ErrorList{}

	if cfg.BaseDikiConfig != nil {
		if len(*cfg.BaseDikiConfig) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("baseDikiConfig"), "must not be empty when specified"))
		}
	}

	if cfg.DefaultScheduledScan != nil {
		fldPath := field.NewPath("defaultScheduledScan")
		if len(cfg.DefaultScheduledScan.Schedule) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("schedule"), "must not be empty when defaultScheduledScan is specified"))
		} else if len(strings.Fields(cfg.DefaultScheduledScan.Schedule)) != 5 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("schedule"), cfg.DefaultScheduledScan.Schedule, "must be a standard 5-field cron expression"))
		}
		if len(cfg.DefaultScheduledScan.Rulesets) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("rulesets"), "must not be empty when defaultScheduledScan is specified"))
		}
		for i, ruleset := range cfg.DefaultScheduledScan.Rulesets {
			rFldPath := fldPath.Child("rulesets").Index(i)
			if len(ruleset.ID) == 0 {
				allErrs = append(allErrs, field.Required(rFldPath.Child("id"), "must not be empty"))
			}
			if len(ruleset.Version) == 0 {
				allErrs = append(allErrs, field.Required(rFldPath.Child("version"), "must not be empty"))
			}
		}
	}

	return allErrs
}
