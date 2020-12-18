// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package root

import (
	getversion "github.com/run-ai/runai-cli/cmd/get"
	"github.com/run-ai/runai-cli/cmd/install"
	"github.com/run-ai/runai-cli/cmd/remove"
	"github.com/run-ai/runai-cli/cmd/set"
	"github.com/run-ai/runai-cli/cmd/uninstall"
	"github.com/run-ai/runai-cli/cmd/update"
	"github.com/run-ai/runai-cli/cmd/upgrade"
	"github.com/run-ai/runai-cli/cmd/version"

	"github.com/run-ai/runai-cli/pkg/config"
	"github.com/run-ai/runai-cli/pkg/util"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var LogLevel string

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   config.CLIName,
		Short: "runai-adm is a command line interface to a RunAI cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		// Would be run before any child command
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(LogLevel)
		},
	}

	// enable logging
	command.PersistentFlags().StringVar(&LogLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")

	command.AddCommand(set.Command())
	command.AddCommand(remove.Command())
	command.AddCommand(upgrade.Command())
	command.AddCommand(version.Command())
	command.AddCommand(update.Command())
	command.AddCommand(getversion.Command())
	command.AddCommand(install.Command())
	command.AddCommand(uninstall.Command())

	return command
}
