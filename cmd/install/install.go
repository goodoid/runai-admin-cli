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

package install

import (
	"fmt"

	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type upgradeFlags struct {
	filePath string
}

func Command() *cobra.Command {
	upgradeFlags := upgradeFlags{}
	var command = &cobra.Command{
		Use:   "install",
		Short: "Install RunAi cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().NFlag() == 0 {
				fmt.Println("No flags were provided")
				cmd.HelpFunc()(cmd, args)
				return
			}

			log.Infof("Installing from file: %v", upgradeFlags.filePath)
			for i := 0; i < 2; i++ {
				kubectl.Apply(upgradeFlags.filePath) // need to remove the crds from this file
			}

			log.Println("Successfully installed Run:AI Cluster")
		},
	}

	command.Flags().StringVarP(&upgradeFlags.filePath, "file", "f", "", "path of runai config .yaml file")

	return command
}
