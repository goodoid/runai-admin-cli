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

package upgrade

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/run-ai/runai-cli/cmd/common"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type upgradeFlags struct {
	filePath        string
	operatorVersion string
}

func Command() *cobra.Command {
	upgradeFlags := upgradeFlags{}
	var command = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade RunAi cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().NFlag() == 0 {
				fmt.Println("No flags were provided")
				cmd.HelpFunc()(cmd, args)
				return
			}

			if upgradeFlags.filePath != "" {
				log.Infof("Installing from file: %v", upgradeFlags.filePath)
				for i := 0; i < 2; i++ {
					kubectl.Apply(upgradeFlags.filePath) // need to remove the crds from this file
				}
			}

			log.Infof("Upgrading CRDs")
			kubectl.Apply("https://raw.githubusercontent.com/run-ai/docs/master/install/runai_new_crds.yaml")

			if upgradeFlags.operatorVersion != "" {
				client := client.GetClient()
				common.ScaleRunaiOperator(client, 0)
				josList, err := client.GetClientset().BatchV1().Jobs("runai").List(metav1.ListOptions{})
				if err != nil {
					fmt.Printf("Failed to list jobs in runai namespace, error: %v", err)
					os.Exit(1)
				}
				for _, job := range josList.Items {
					client.GetClientset().BatchV1().Jobs("runai").Delete(job.Name, &metav1.DeleteOptions{})
					log.Debugf("Deleted Job: %v", job.Name)
				}

				upgradeVersion(client, upgradeFlags)

				common.ScaleRunaiOperator(client, 1)
			}

			log.Println("Successfully upgraded RunAi Cluster")
		},
	}

	command.Flags().StringVarP(&upgradeFlags.filePath, "file", "f", "", "path of runai config .yaml file")
	command.Flags().StringVarP(&upgradeFlags.operatorVersion, "version", "v", "", "set version of runai operator")

	return command
}

func upgradeVersion(client *client.Client, upgradeFlags upgradeFlags) {
	var err error
	var deployment *appsv1.Deployment
	shouldDeleteStsAndPvc := false
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		deployment, err = client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
		if err != nil {
			log.Infof("runai operator doesnt exist on runai namespace, error: %v", err)
			os.Exit(1)
		}
		currentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, ":")
		currentTag := currentImage[1]
		if currentTag == "latest" {
			if upgradeFlags.operatorVersion != "latest" {
				log.Infof("Setting image to 'latest' because old image was 'latest'")
			}
		} else {
			currentMinorVersion := strings.Split(currentTag, ".")
			currentMinorInt, _ := strconv.Atoi(currentMinorVersion[2])
			deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", currentImage[0], upgradeFlags.operatorVersion)
			shouldDeleteStsAndPvc = currentMinorInt <= 77
		}
		_, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
		if err != nil {
			log.Debugf("Failed to update deployment runai operator, attempt: %v, error: %v", i, err)
			continue
		}
		break
	}
	if err != nil {
		log.Infof("Failed to update runai-operator with new tag, error: %v", err)
		os.Exit(1)
	}

	if shouldDeleteStsAndPvc {
		stsList, error := client.GetClientset().AppsV1().StatefulSets("runai").List(metav1.ListOptions{})
		if error != nil {
			log.Infof("Failed to list statefulsets in runai namespace, error: %v", err)
			os.Exit(1)
		}
		for _, sts := range stsList.Items {
			client.GetClientset().AppsV1().StatefulSets("runai").Delete(sts.Name, &metav1.DeleteOptions{})
			log.Debugf("Deleted Statefulset: %v", sts.Name)
		}

		pvcList, err := client.GetClientset().CoreV1().PersistentVolumeClaims("runai").List(metav1.ListOptions{})
		if err != nil {
			log.Infof("Failed to list PVCs in runai namespace, error: %v", err)
			os.Exit(1)
		}
		for _, pvc := range pvcList.Items {
			client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete(pvc.Name, &metav1.DeleteOptions{})
			log.Debugf("Deleted PVC: %v", pvc.Name)
		}
	}
}
