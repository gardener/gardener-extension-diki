// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	configapi "github.com/gardener/gardener-extension-diki/pkg/apis/config"
	configv1alpha1 "github.com/gardener/gardener-extension-diki/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-diki/pkg/apis/config/validation"
	"github.com/gardener/gardener-extension-diki/pkg/controller/lifecycle"
)

var (
	configScheme  *runtime.Scheme
	configDecoder runtime.Decoder
)

func init() {
	configScheme = runtime.NewScheme()
	utilruntime.Must(configapi.AddToScheme(configScheme))
	utilruntime.Must(configv1alpha1.AddToScheme(configScheme))

	configDecoder = serializer.NewCodecFactory(configScheme).UniversalDecoder()
}

// DikiOptions holds options related to the diki extension configuration.
type DikiOptions struct {
	ConfigLocation string
	config         *DikiServiceConfig
}

// AddFlags implements Flagger.AddFlags.
func (o *DikiOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigLocation, "config", "", "Path to diki extension configuration")
}

// Complete implements Completer.Complete.
func (o *DikiOptions) Complete() error {
	if len(o.ConfigLocation) == 0 {
		o.config = &DikiServiceConfig{}
		return nil
	}

	data, err := os.ReadFile(o.ConfigLocation)
	if err != nil {
		return err
	}

	cfg := configapi.Configuration{}
	if err := runtime.DecodeInto(configDecoder, data, &cfg); err != nil {
		return err
	}

	if errs := validation.ValidateConfiguration(&cfg); len(errs) > 0 {
		return errs.ToAggregate()
	}

	o.config = &DikiServiceConfig{config: cfg}
	return nil
}

// Completed returns the decoded DikiServiceConfig instance. Only call this if `Complete` was successful.
func (o *DikiOptions) Completed() *DikiServiceConfig {
	return o.config
}

// DikiServiceConfig contains configuration information about the diki extension.
type DikiServiceConfig struct {
	config configapi.Configuration
}

// Apply applies the DikiServiceConfig to the passed Configuration instance.
func (c *DikiServiceConfig) Apply(cfg *configapi.Configuration) {
	*cfg = c.config
}

// ControllerSwitches are the cmd.SwitchOptions for the diki extension controllers.
func ControllerSwitches() *extensionscmdcontroller.SwitchOptions {
	return extensionscmdcontroller.NewSwitchOptions(
		extensionscmdcontroller.Switch(lifecycle.ControllerName, lifecycle.AddToManager),
		extensionscmdcontroller.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
