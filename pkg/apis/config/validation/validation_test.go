// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-diki/pkg/apis/config"
	. "github.com/gardener/gardener-extension-diki/pkg/apis/config/validation"
)

var _ = Describe("ValidateConfiguration", func() {
	var cfg *config.Configuration

	BeforeEach(func() {
		cfg = &config.Configuration{}
	})

	It("should allow an empty configuration", func() {
		errs := ValidateConfiguration(cfg)
		Expect(errs).To(BeEmpty())
	})

	It("should allow a valid baseDikiOptions", func() {
		cfg.BaseDikiOptions = &config.BaseDikiOptionsConfig{
			Data: "providers:\n- id: managedk8s\n",
		}
		errs := ValidateConfiguration(cfg)
		Expect(errs).To(BeEmpty())
	})

	It("should reject baseDikiOptions with empty data", func() {
		cfg.BaseDikiOptions = &config.BaseDikiOptionsConfig{
			Data: "",
		}
		errs := ValidateConfiguration(cfg)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeRequired))
		Expect(errs[0].Field).To(Equal("baseDikiOptions.data"))
	})
})
