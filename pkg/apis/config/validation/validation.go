// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-diki/pkg/apis/config"
)

// ValidateConfiguration validates the passed configuration instance.
func ValidateConfiguration(cfg *config.Configuration) field.ErrorList {
	allErrs := field.ErrorList{}

	if cfg.BaseDikiOptions != nil {
		fldPath := field.NewPath("baseDikiOptions")
		if len(cfg.BaseDikiOptions.Data) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("data"), "data must not be empty when baseDikiOptions is specified"))
		}
	}

	return allErrs
}
