// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"github.com/gardener/gardener/extensions/pkg/util"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-diki/pkg/controller/lifecycle"
)

var log = logf.Log.WithName("gardener-extension-diki")

// NewServiceControllerCommand creates a new command that is used to start the diki extension controller.
func NewServiceControllerCommand() *cobra.Command {
	options := NewOptions()

	cmd := &cobra.Command{
		Use:           "gardener-extension-diki",
		Short:         "Diki extension controller manages diki-operator deployments for shoot clusters.",
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			verflag.PrintAndExitIfRequested()

			log.Info("Starting gardener-extension-diki", "version", version.Get())

			if err := options.optionAggregator.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			if err := options.heartbeatOptions.Validate(); err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return options.run(cmd.Context())
		},
	}

	verflag.AddFlags(cmd.Flags())
	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}

func (o *Options) run(ctx context.Context) error {
	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfigv1alpha1.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, o.restOptions.Completed().Config)

	mgrOpts := o.managerOptions.Completed().Options()

	mgrOpts.Client = client.Options{
		Cache: &client.CacheOptions{
			DisableFor: []client.Object{
				&corev1.Secret{},    // applied for ManagedResources
				&corev1.ConfigMap{}, // applied for monitoring config
			},
		},
	}

	mgr, err := manager.New(o.restOptions.Completed().Config, mgrOpts)
	if err != nil {
		return fmt.Errorf("could not instantiate controller-manager: %w", err)
	}

	if err := extensionscontroller.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("could not update manager scheme: %w", err)
	}

	o.controllerOptions.Completed().Apply(&lifecycle.DefaultAddOptions.ControllerOptions)
	o.reconcileOptions.Completed().Apply(&lifecycle.DefaultAddOptions.IgnoreOperationAnnotation)
	o.heartbeatOptions.Completed().Apply(&extensionsheartbeatcontroller.DefaultAddOptions)

	if err := o.controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
		return fmt.Errorf("could not add controllers to manager: %w", err)
	}

	if err := mgr.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
		return fmt.Errorf("could not add ready check for informers: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("could not add health check to manager: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}

	return nil
}
