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
package commands

import (
	"fmt"
	"os"

	"github.com/kubeflow/arena/util/helm"
	"github.com/spf13/cobra"
)

func NewLogViewerCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "logviewer job",
		Short: "display Log Viewer URL of a training job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name = args[0]
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			exist, err := helm.CheckRelease(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !exist {
				fmt.Printf("The job %s doesn't exist, please create it first. use 'arena create'\n", name)
				os.Exit(1)
			}
			job, err := getTrainingJob(client, name, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			urls, err := job.GetJobDashboards(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if len(urls) > 0 {
				fmt.Printf("Your LogViewer will be available on:\n")
				for _, url := range urls {
					fmt.Println(url)
				}
			} else {
				fmt.Printf("No LogViewer Installed")
			}

		},
	}

	return command
}
