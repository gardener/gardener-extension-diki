// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(c client.Client) extension.Actuator {
	return &actuator{
		client: c,
	}
}

type actuator struct {
	client client.Client
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(_ context.Context, logger logr.Logger, _ *extensionsv1alpha1.Extension) error {
	logger.Info("Reconciling diki extension")
	return nil
}

// Delete the Extension resource.
func (a *actuator) Delete(_ context.Context, logger logr.Logger, _ *extensionsv1alpha1.Extension) error {
	logger.Info("Deleting diki extension")
	return nil
}

// ForceDelete the Extension resource.
func (a *actuator) ForceDelete(_ context.Context, logger logr.Logger, _ *extensionsv1alpha1.Extension) error {
	logger.Info("Force deleting diki extension")
	return nil
}

// Migrate the Extension resource.
func (a *actuator) Migrate(_ context.Context, logger logr.Logger, _ *extensionsv1alpha1.Extension) error {
	logger.Info("Migrating diki extension")
	return nil
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}
