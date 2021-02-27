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

package uninstall

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"

	"github.com/run-ai/runai-cli/cmd/common"
	"github.com/run-ai/runai-cli/pkg/client"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type uninstallFlags struct {
	deleteAll bool
}

func Command() *cobra.Command {
	uninstallFlags := uninstallFlags{}
	var command = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Run:AI cluster",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := client.GetClient()
			if uninstallFlags.deleteAll {
				log.Infof("Deleting RunaiConfig")
				deleteRunaiConfig(client)
			} else {
				common.ScaleRunaiOperator(client, 0)
			}
			deleteAllResources(client, uninstallFlags)
			deleteResourcesByKubectlCommand()

			if uninstallFlags.deleteAll {
				err := client.GetClientset().CoreV1().Namespaces().Delete("runai", &metav1.DeleteOptions{})
				if err != nil {
					log.Infof("Failed to delete namespace runai, error %v", err)
					os.Exit(1)
				}
				log.Infof("Deleted namespace runai")
			}
			log.Println("Successfully uninstalled Run:AI Cluster")
		},
	}
	command.Flags().BoolVarP(&uninstallFlags.deleteAll, "all", "A", false, "use flag to delete: Runai Namespace, RunaiConfig, Runai Operator")

	return command
}

func deleteAllResources(client *client.Client, uninstallFlags uninstallFlags) {
	deployments, err := client.GetClientset().AppsV1().Deployments("runai").List(metav1.ListOptions{})
	if err == nil {
		for _, deployment := range deployments.Items {
			if !uninstallFlags.deleteAll && deployment.Name == "runai-operator" {
				log.Infof("Keeping RunAI Operator with 0 replicas")
				continue
			}
			client.GetClientset().AppsV1().Deployments("runai").Delete(deployment.Name, &metav1.DeleteOptions{})
			log.Debugf("deleted deployment %v", deployment.Name)
		}
	}

	dss, err := client.GetClientset().AppsV1().DaemonSets("runai").List(metav1.ListOptions{})
	if err == nil {
		for _, ds := range dss.Items {
			client.GetClientset().AppsV1().DaemonSets("runai").Delete(ds.Name, &metav1.DeleteOptions{})
			log.Debugf("deleted ds %v", ds.Name)
		}
	}

	stss, err := client.GetClientset().AppsV1().StatefulSets("runai").List(metav1.ListOptions{})
	if err == nil {
		for _, sts := range stss.Items {
			client.GetClientset().AppsV1().StatefulSets("runai").Delete(sts.Name, &metav1.DeleteOptions{})
			log.Debugf("deleted sts %v", sts.Name)
		}
	}

	jobs, err := client.GetClientset().BatchV1().Jobs("runai").List(metav1.ListOptions{})
	if err == nil {
		for _, job := range jobs.Items {
			client.GetClientset().AppsV1().StatefulSets("runai").Delete(job.Name, &metav1.DeleteOptions{})
			log.Debugf("deleted job %v", job.Name)
		}
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

func deleteRunaiConfig(client *client.Client) {
	runaiconfigResource := schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}
	var error error
	var runaiConfig *unstructured.Unstructured
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		runaiConfig, error = client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Get("runai", metav1.GetOptions{})
		if error != nil {
			fmt.Println("Failed to get RunaiConfig")
			return
		}
		var emptyMap []string
		err := unstructured.SetNestedStringSlice(runaiConfig.Object, emptyMap, "metadata", "finalizers")
		if err != nil {
			fmt.Printf("Failed to update RunaiConfig finalizer, error: %v", err)
			os.Exit(1)
		}
		_, error = client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Update(runaiConfig, metav1.UpdateOptions{})
		if error != nil {
			log.Debugf("Failed to update runaiconfig, attempt: %v, error: %v", i, error)
			continue
		}
		error = client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Delete("runai", &metav1.DeleteOptions{})
		if error != nil {
			log.Debugf("Failed to delete runaiconfig, attempt: %v, error: %v", i, error)
			continue
		}

		break
	}

	if error != nil {
		log.Infof("Failed to update runaiconfig, error: %v", error)
		os.Exit(1)
	}

	log.Infof("Deleted runaiconfig")
}

func deleteResourcesByKubectlCommand() {
	pspToDelete := []string{"psp", "runai-admission-controller", "runai-grafana", "runai-grafana-test", "runai-init-ca", "runai-kube-state-metrics", "runai-local-path-provisioner", "runai-prometheus-node-exporter", "runai-prometheus-operator-operator", "runai-prometheus-operator-prometheus", "runai-prometheus-pushgateway", "runai-nginx-ingress", "runai-nginx-ingress-backend", "mpi-operator", "runai-job-controller", "runai-prometheus-operator-admission", "runai-project-controller", "runai-kube-prometheus-stac-prometheus", "nfd-master", "runai-job-viewer", "runai-job-executor"}
	kubectl.Delete(pspToDelete)

	clusterRoleToDelete := []string{"clusterrole", "init-ca", "psp-runai-kube-state-metrics", "psp-runai-prometheus-node-exporter", "runai", "runai-admission-controller", "runai-grafana-clusterrole", "runai-kube-state-metrics", "runai-prometheus-operator-operator", "runai-prometheus-operator-operator-psp", "runai-prometheus-operator-prometheus", "runai-prometheus-operator-prometheus-psp", "runai-local-path-provisioner", "mpi-operator", "runai-nginx-ingress", "runai-job-controller", "runai-nfs-client-provisioner-runner", "runai-project-controller", "runai-kube-prometheus-stac-operator", "runai-kube-prometheus-stac-operator-psp", "runai-kube-prometheus-stac-prometheus", "runai-kube-prometheus-stac-prometheus-psp", "nfd-master", "runai-job-viewer", "runai-job-executor", "runai-cli-index-map-editor", "runai-scheduler-rw", "runai-scheduler-ro", "runai-project-controller-project", "runai-project-controller-administrator", "runai-operator", "runai-nvidia-device-plugin", "runai-job-controller-project", "runai-agent", "researcher-service"}
	kubectl.Delete(clusterRoleToDelete)

	clusterRoleBindingToDelete := []string{"clusterrolebinding", "default-sa-admin", "init-ca", "psp-runai-kube-state-metrics", "psp-runai-prometheus-node-exporter", "runai", "runai-admission-controller", "runai-grafana-clusterrolebinding", "runai-kube-state-metrics", "runai-prometheus-operator-operator", "runai-prometheus-operator-operator-psp", "runai-prometheus-operator-prometheus", "runai-prometheus-operator-prometheus-psp", "runai-local-path-provisioner", "mpi-operator", "runai-nginx-ingress", "runai-job-controller", "run-runai-nfs-client-provisioner", "runai-project-controller", "runai-kube-prometheus-stac-operator", "runai-kube-prometheus-stac-operator-psp", "runai-kube-prometheus-stac-prometheus", "runai-kube-prometheus-stac-prometheus-psp", "nfd-master", "runai-job-viewer", "runai-job-executor", "researcher-service", "runai-agent", "runai-nvidia-device-plugin", "runai-operator", "runai-project-controller-administrator", "runai-scheduler-ro", "runai-scheduler-rw"}
	kubectl.Delete(clusterRoleBindingToDelete)

	mutatingWebhookConfigurationToDelete := []string{"MutatingWebhookConfiguration", "runai-fractional-gpus", "runai-label-project", "runai-mutating-webhook", "runai-prometheus-operator-admission", "runai-reporter-library", "runai-node-affinity", "runai-resource-gpu-factor", "runai-kube-prometheus-stac-admission"}
	kubectl.Delete(mutatingWebhookConfigurationToDelete)

	validatingWebhookConfiguration := []string{"ValidatingWebhookConfiguration", "runai-prometheus-operator-admission", "runai-validate-elastic", "runai-validate-fractional", "runai-kube-prometheus-stac-admission"}
	kubectl.Delete(validatingWebhookConfiguration)

	pcToDelete := []string{"pc", "build", "interactive-preemptible", "train", "runai-critical"}
	kubectl.Delete(pcToDelete)

	scToDelete := []string{"sc", "local-path", "nfs-client"}
	kubectl.Delete(scToDelete)

	departmentToDelete := []string{"department", "default"}
	kubectl.Delete(departmentToDelete)

	services := []string{"service", "-n", "kube-system", "runai-prometheus-operator-coredns", "runai-prometheus-operator-kube-controller-manager", "runai-prometheus-operator-kube-etcd", "runai-prometheus-operator-kube-proxy", "runai-prometheus-operator-kube-scheduler", "runai-prometheus-operator-kubelet", "kube-prometheus-stack-kubelet", "prom-kube-prometheus-stack-kubelet", "runai-kube-prometheus-stac-kubelet"}
	kubectl.Delete(services)

	roles := []string{"roles", "-n", "runai", "--all"}
	kubectl.Delete(roles)

	svc := []string{"services", "-n", "runai", "--all"}
	kubectl.Delete(svc)

	mutatingwebhookconfigurations := []string{"mutatingwebhookconfigurations.admissionregistration.k8s.io", "-n", "runai", "--all"}
	kubectl.Delete(mutatingwebhookconfigurations)

	serviceaccounts := []string{"serviceaccount", "-n", "runai", "--all"}
	kubectl.Delete(serviceaccounts)

	servicemonitor := []string{"servicemonitor", "-n", "runai", "--all"}
	kubectl.Delete(servicemonitor)

	rolebindings := []string{"rolebinding", "-n", "runai", "--all"}
	kubectl.Delete(rolebindings)
}
