// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package dikioperator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDikiOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DikiOperator Component Suite")
}
