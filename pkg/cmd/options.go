// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"

	"github.com/gardener/gardener-extension-diki/pkg/controller/lifecycle"
)

// ControllerSwitches are the cmd.SwitchOptions for the diki extension controllers.
func ControllerSwitches() *extensionscmdcontroller.SwitchOptions {
	return extensionscmdcontroller.NewSwitchOptions(
		extensionscmdcontroller.Switch(lifecycle.ControllerName, lifecycle.AddToManager),
		extensionscmdcontroller.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
