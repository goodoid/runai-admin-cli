package noderole

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type nodeRoleTypes struct {
	CpuWorker         bool
	AllNodes          bool
	GpuWorker         bool
	RunaiSystemWorker bool
}

const (
	gpuWorkerLabel    = "node-role.kubernetes.io/runai-gpu-worker"
	cpuWorkerLabel    = "node-role.kubernetes.io/runai-cpu-worker"
	systemWorkerLabel = "node-role.kubernetes.io/runai-system-worker"
)

func Set() *cobra.Command {
	flags := nodeRoleTypes{}
	var command = &cobra.Command{
		Use:     "node-role NODE_NAME",
		Aliases: []string{"node-roles"},
		Short:   "Set node with roles",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := client.GetClient()
			nodesInCluster := labelNodesWithRoles(client, flags, args)
			updateRunaiConfig(client, flags)
			deleteResourcesIfNeeded(flags, client, nodesInCluster)
		},
	}

	command.Flags().BoolVar(&flags.AllNodes, "all", false, "set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}

func deletePodsIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]*v1.Node) {
	if !flags.RunaiSystemWorker && !flags.CpuWorker && !flags.GpuWorker {
		return
	}
	runaiPods, err := client.GetClientset().CoreV1().Pods("runai").List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to list pods from runai namespace")
		os.Exit(1)
	}

	for _, pod := range runaiPods.Items {
		deletePodIfNeeded(pod, nodesInCluster, client)
	}
}

func deletePodIfNeeded(pod v1.Pod, nodesInCluster map[string]*v1.Node, client *client.Client) {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		return
	}
	for _, nodeSelectorTerms := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, matchExpressions := range nodeSelectorTerms.MatchExpressions {
			if matchExpressions.Key == systemWorkerLabel {
				if _, found := nodesInCluster[pod.Spec.NodeName].Labels[systemWorkerLabel]; found {
					return
				}
				err := client.GetClientset().CoreV1().Pods("runai").Delete(pod.Name, &metav1.DeleteOptions{})
				if err != nil {
					log.Debugf("Failed to delete pod: %v, error: %v", pod.Name, err)
				}
			}
		}
	}
}

func deleteResourcesIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]*v1.Node) {
	log.Info("Updating RunAi resources")
	deletePVCIfNeeded(flags, client, nodesInCluster)
	deletePodsIfNeeded(flags, client, nodesInCluster)
}
func deletePVCIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]*v1.Node) {
	if !flags.RunaiSystemWorker {
		return
	}
	pvc, err := client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Get("data-runai-db-0", metav1.GetOptions{})
	if err != nil {
		fmt.Println("Failed to list PVCs from runai namespace")
		os.Exit(1)
	}

	pvcNode, found := pvc.Annotations["volume.kubernetes.io/selected-node"]
	if !found {
		return
	}

	nodeInfo, found := nodesInCluster[pvcNode]
	if !found {
		fmt.Println("Failed to find pvc node in cluster, node: %s", pvcNode)
		os.Exit(1)
	}

	if _, found := nodeInfo.Labels[systemWorkerLabel]; found { // no need to delete the pvc - already on a system node
		return
	}

	err = client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Delete("data-runai-db-0", &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println("Failed to delete pvc data-runai-db-0, error: %v", err)
		os.Exit(1)
	}
	log.Debugf("Deleted PVC data-runai-db-0")
}

func updateRunaiConfig(client *client.Client, flags nodeRoleTypes) {
	log.Info("Updating RunAi configurations")
	updateRunaiConfigIfNeeded(client, flags)
	updateRunaiDeploymentIfNeeded(client, flags)
}

func updateRunaiDeploymentIfNeeded(client *client.Client, flags nodeRoleTypes) {
	if !flags.RunaiSystemWorker {
		return
	}
	deployment, err := client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
	if err != nil {
		fmt.Println("Failed to get runai-operator, error: %v", err)
		os.Exit(1)
	}

	nodeSelectorTerms := []v1.NodeSelectorTerm{
		{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{
					Key:      systemWorkerLabel,
					Operator: v1.NodeSelectorOpExists,
				},
			},
		},
	}
	deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = nodeSelectorTerms

	_, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
	if err != nil {
		fmt.Println("Failed to update runai-operator, error: %v", err)
		os.Exit(1)
	}

	log.Debugf("Updated runai operator to have node affinity")
}

func updateRunaiConfigIfNeeded(client *client.Client, flags nodeRoleTypes) {
	runaiconfigResource := schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}
	runaiConfig, err := client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Get("runai", metav1.GetOptions{})
	var needToUpdateRunaiConfig bool
	if err != nil {
		fmt.Println("Failed to get RunaiConfig, RunAI isn't installed on the cluster")
		os.Exit(1)
	}
	if flags.CpuWorker || flags.GpuWorker {
		wasRestrictSchedulingEnabled, found, err := unstructured.NestedBool(runaiConfig.Object, "spec.global.nodeAffinity", "restrictScheduling")
		if err != nil || !found {
			fmt.Printf("Failed to get restrictScheduling from runaiConfig, error: %v", err)
			os.Exit(1)
		}

		needToUpdateRunaiConfig = needToUpdateRunaiConfig || !wasRestrictSchedulingEnabled

		if err := unstructured.SetNestedField(runaiConfig.Object, true, "spec.global.nodeAffinity", "restrictScheduling"); err != nil {
			fmt.Println("Failed to update cluster to support restrictScheduling")
			os.Exit(1)
		}
	}

	if flags.RunaiSystemWorker {
		wasRestrictRunaiSystemEnabled, found, err := unstructured.NestedBool(runaiConfig.Object, "spec.global.nodeAffinity", "restrictRunaiSystem")
		if err != nil || !found {
			fmt.Printf("yodarsdebug %v", runaiConfig.Object)
			fmt.Printf("Failed to get restrictRunaiSystem from runaiConfig, error: %v, found: %v", err, found)
			os.Exit(1)
		}
		needToUpdateRunaiConfig = needToUpdateRunaiConfig || !wasRestrictRunaiSystemEnabled

		if err := unstructured.SetNestedField(runaiConfig.Object, true, "spec.global.nodeAffinity", "restrictRunaiSystem"); err != nil {
			fmt.Println("Failed to update cluster to support restrictRunaiSystem")
			os.Exit(1)
		}

	}

	if needToUpdateRunaiConfig {
		_, err := client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Update(runaiConfig, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("Failed to update runaiconfig")
			os.Exit(1)
		}
	}
}

func labelNodesWithRoles(client *client.Client, flags nodeRoleTypes, args []string) map[string]*v1.Node {
	log.Info("Updating nodes with roles")

	allNodeClusters := map[string]*v1.Node{}
	nodesInCluster, err := client.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil || len(nodesInCluster.Items) == 0 {
		fmt.Println("Failed to list nodesInCluster")
		os.Exit(1)
	}

	wasAnyNodeUpdated := false
	if flags.AllNodes {
		for _, nodeInfo := range nodesInCluster.Items {
			setLabelsSingleNode(&nodeInfo, flags, client)
			allNodeClusters[nodeInfo.Name] = &nodeInfo
			wasAnyNodeUpdated = true
		}
		log.Debugf("Successfully updated all nodes with roles")
		return allNodeClusters
	} else {
		wasNodeUpdate := map[string]bool{}
		for _, nodeToLabel := range args {
			wasNodeUpdate[nodeToLabel] = false
			for _, nodeInfo := range nodesInCluster.Items {
				if wasNodeUpdate[nodeToLabel] == true {
					continue
				}

				if nodeToLabel == nodeInfo.Name {
					setLabelsSingleNode(&nodeInfo, flags, client)
					wasNodeUpdate[nodeToLabel] = true
					wasAnyNodeUpdated = true
					allNodeClusters[nodeInfo.Name] = &nodeInfo
				}
				if _, found := allNodeClusters[nodeInfo.Name]; !found {
					allNodeClusters[nodeInfo.Name] = &nodeInfo
				}
			}
		}
		for nodeName := range wasNodeUpdate {
			if !wasNodeUpdate[nodeName] {
				log.Infof("Node: %v was not found in cluster", nodeName)
			}
		}
	}

	if !wasAnyNodeUpdated {
		log.Infof("All nodes are already updated")
		os.Exit(1)
	}

	return allNodeClusters
}

func setLabelsSingleNode(nodeInfo *v1.Node, o nodeRoleTypes, client *client.Client) {
	if nodeInfo.Labels == nil {
		nodeInfo.Labels = map[string]string{}
	}
	if o.GpuWorker {
		nodeInfo.Labels[gpuWorkerLabel] = ""
	}
	if o.CpuWorker {
		nodeInfo.Labels[cpuWorkerLabel] = ""
	}
	if o.RunaiSystemWorker {
		nodeInfo.Labels[systemWorkerLabel] = ""
	}

	_, err := client.GetClientset().CoreV1().Nodes().Update(nodeInfo)
	if err != nil {
		log.Infof("Failed to update node: %v, ", err)
		os.Exit(1)
	}
}

func Remove() *cobra.Command {
	flags := nodeRoleTypes{}
	var command = &cobra.Command{
		Use:     "node-role NODE_NAME",
		Aliases: []string{"node-roles"},
		Short:   "remove node with roles",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := client.GetClient()
			nodesInCluster, err := client.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
			if err != nil || len(nodesInCluster.Items) == 0 {
				fmt.Println("Failed to list nodesInCluster")
				os.Exit(1)
			}

			if flags.AllNodes {
				for _, nodeInfo := range nodesInCluster.Items {
					removeLabelsFromSingleNode(nodeInfo, flags, client)
				}
				log.Infof("Successfully updated all nodes with roles")
				return
			}

			wasNodeUpdate := map[string]bool{}
			for _, nodeToLabel := range args {
				wasNodeUpdate[nodeToLabel] = false
				for _, nodeInfo := range nodesInCluster.Items {
					if nodeToLabel != nodeInfo.Name {
						continue
					}
					if wasNodeUpdate[nodeToLabel] == true {
						continue
					}

					removeLabelsFromSingleNode(nodeInfo, flags, client)
					wasNodeUpdate[nodeToLabel] = true
				}
			}

			for nodeName := range wasNodeUpdate {
				if !wasNodeUpdate[nodeName] {
					log.Infof("Node: %v was not found in cluster", nodeName)
				}
			}

			log.Infof("Successfully update nodes with roles")
		},
	}

	command.Flags().BoolVar(&flags.AllNodes, "all", false, "set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}

func removeLabelsFromSingleNode(nodeInfo v1.Node, flags nodeRoleTypes, client *client.Client) {
	if nodeInfo.Labels == nil {
		nodeInfo.Labels = map[string]string{}
	}
	if flags.GpuWorker {
		delete(nodeInfo.Labels, gpuWorkerLabel)
	}
	if flags.CpuWorker {
		delete(nodeInfo.Labels, cpuWorkerLabel)
	}
	if flags.RunaiSystemWorker {
		delete(nodeInfo.Labels, systemWorkerLabel)
	}

	_, err := client.GetClientset().CoreV1().Nodes().Update(&nodeInfo)
	if err != nil {
		log.Infof("Failed to update node: %v, ", err)
		os.Exit(1)
	}
}
