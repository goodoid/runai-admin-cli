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

	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
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

			log.Infof("Upgrading CRDs")
			kubectl.Apply("https://raw.githubusercontent.com/run-ai/docs/master/install/runai_new_crds.yaml")
			if upgradeFlags.filePath != "" {
				log.Infof("Installing from file: %v", upgradeFlags.filePath)
				for i := 0; i < 3; i++ {
					kubectl.Apply(upgradeFlags.filePath)
				}
			}

			client := client.GetClient()
			deployment, err := client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Failed to get runai-operator, error: %v", err)
				os.Exit(1)
			}
			zeroReplicas := int32(0)
			deployment.Spec.Replicas = &zeroReplicas
			deployment, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
			if err != nil {
				fmt.Printf("Failed to update runai-operator, error: %v", err)
				os.Exit(1)
			}

			josList, err := client.GetClientset().BatchV1().Jobs("runai").List(metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to list jobs in runai namespace, error: %v", err)
				os.Exit(1)
			}
			for _, job := range josList.Items {
				client.GetClientset().BatchV1().Jobs("runai").Delete(job.Name, &metav1.DeleteOptions{})
				log.Debugf("Deleted Job: %v", job.Name)
			}

			upgradeVersionIfNeeded(deployment, client, upgradeFlags)
			deployment, err = client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Failed to get runai-operator, error: %v", err)
				os.Exit(1)
			}
			oneReplicas := int32(1)
			deployment.Spec.Replicas = &oneReplicas
			deployment, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)

			log.Println("Succesfully upgraded RunAi Cluster")
		},
	}

	command.Flags().StringVarP(&upgradeFlags.filePath, "file", "f", "", "path of runai config .yaml file")
	command.Flags().StringVarP(&upgradeFlags.operatorVersion, "version", "v", "", "set version of runai operator")

	return command
}

func upgradeVersionIfNeeded(deployment *v1.Deployment, client *client.Client, upgradeFlags upgradeFlags) {
	if upgradeFlags.operatorVersion == "" {
		return
	}

	currentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, ":")
	currentTag := currentImage[1]
	currentMinorVersion := strings.Split(currentTag, ".")
	currentMinorInt, _ := strconv.Atoi(currentMinorVersion[2])
	deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", currentImage, upgradeFlags.operatorVersion)
	_, err := client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
	if err != nil {
		fmt.Printf("Failed to update runai-operator with new tag, error: %v", err)
		os.Exit(1)
	}

	if currentMinorInt <= 77 {
		stsList, err := client.GetClientset().AppsV1().StatefulSets("runai").List(metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Failed to list statefulsets in runai namespace, error: %v", err)
			os.Exit(1)
		}
		for _, sts := range stsList.Items {
			client.GetClientset().AppsV1().StatefulSets("runai").Delete(sts.Name, &metav1.DeleteOptions{})
			log.Debugf("Deleted Statefulset: %v", sts.Name)
		}

		pvcList, err := client.GetClientset().CoreV1().PersistentVolumeClaims("runai").List(metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Failed to list pvc in runai namespace, error: %v", err)
			os.Exit(1)
		}
		for _, pvc := range pvcList.Items {
			client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete(pvc.Name, &metav1.DeleteOptions{})
			log.Debugf("Deleted PVC: %v", pvc.Name)
		}
	}
}
