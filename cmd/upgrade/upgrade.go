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
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/run-ai/runai-cli/autogenerate"
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
		Short: "Upgrade Run:AI cluster",
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

			upgradeYamlsBeforeRun()

			if upgradeFlags.operatorVersion != "" {
				client := client.GetClient()
				common.ScaleRunaiOperator(client, 0)
				josList, err := client.GetClientset().BatchV1().Jobs("runai").List(metav1.ListOptions{})
				if err != nil {
					fmt.Printf("Failed to list jobs in the runai namespace, error: %v", err)
					os.Exit(1)
				}
				for _, job := range josList.Items {
					client.GetClientset().BatchV1().Jobs("runai").Delete(job.Name, &metav1.DeleteOptions{})
					log.Debugf("Deleted Job: %v", job.Name)
				}

				upgradeVersion(client, upgradeFlags)

				common.ScaleRunaiOperator(client, 1)
			}

			log.Println("Successfully upgraded the Run:AI Cluster")
		},
	}

	command.Flags().StringVarP(&upgradeFlags.filePath, "file", "f", "", "Path of a Run:AI configuration .yaml file")
	command.Flags().StringVarP(&upgradeFlags.operatorVersion, "version", "v", "", "Set a Run:AI version (e.g. 1.0.45)")

	return command
}

func upgradeYamlsBeforeRun() {
	log.Infof("Upgrading yamls before upgrade")
	file, err := ioutil.TempFile("", "pre_upgrade.yaml")
	if err != nil {
		fmt.Println(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write([]byte(autogenerate.PreInstallYaml)); err != nil {
		fmt.Errorf("failed to write file error: %v", err)
		os.Exit(1)
	}

	kubectl.Apply(file.Name())
}

func upgradeVersion(client *client.Client, upgradeFlags upgradeFlags) {
	var err error
	var deployment *appsv1.Deployment
	shouldDeleteStsAndPvc := false
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		deployment, err = client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
		if err != nil {
			log.Infof("Run:AI operator does not exist on runai namespace, error: %v", err)
			os.Exit(1)
		}
		currentImage := strings.Split(deployment.Spec.Template.Spec.Containers[0].Image, ":")
		currentTag := currentImage[1]
		if currentTag == "latest" {
			if upgradeFlags.operatorVersion != "latest" {
				log.Infof("Setting image to 'latest' as an old image was 'latest'")
			}
		} else {
			currentMinorVersion := strings.Split(currentTag, ".")
			currentMinorInt, _ := strconv.Atoi(currentMinorVersion[2])
			deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", currentImage[0], upgradeFlags.operatorVersion)
			shouldDeleteStsAndPvc = currentMinorInt <= 77
		}
		_, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
		if err != nil {
			log.Debugf("Failed to update the deployment of the Run:AI operator, attempt: %v, error: %v", i, err)
			continue
		}
		break
	}
	if err != nil {
		log.Infof("Failed to update Run:AI operator with new tag, error: %v", err)
		os.Exit(1)
	}

	if shouldDeleteStsAndPvc {
		err = client.GetClientset().AppsV1().StatefulSets("runai").Delete("runai-db", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted Statefulset: runai-db")
		}

		err = client.GetClientset().AppsV1().StatefulSets("runai").Delete("runai-prometheus-pushgateway", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted Statefulset: runai-prometheus-pushgateway")
		}

		err = client.GetClientset().AppsV1().StatefulSets("runai").Delete("prometheus-runai-prometheus-operator-prometheus", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted Statefulset: prometheus-runai-prometheus-operator-prometheus")
		}

		err = client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete("data-runai-db-0", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted PVC: data-runai-db-0")
		}

		err = client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete("prometheus-runai-prometheus-operator-prometheus-db-prometheus-runai-prometheus-operator-prometheus-0", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted PVC: prometheus-runai-prometheus-operator-prometheus-db-prometheus-runai-prometheus-operator-prometheus-0")
		}

		err = client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete("storage-volume-runai-prometheus-pushgateway-0", &metav1.DeleteOptions{})
		if err == nil {
			log.Debugf("Deleted PVC: storage-volume-runai-prometheus-pushgateway-0")
		}
	}
}
