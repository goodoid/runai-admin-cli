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
	systemWorkerLabel = "node-role.kubernetes.io/runai-system"
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
			nodesInCluster := labelNodesWithRolesAndGetNodesInCluster(client, flags, args, true)
			updateRunaiConfigurations(client, flags, nodesInCluster)
			deleteResourcesIfNeeded(flags, client, nodesInCluster)
			log.Info("Successfully updated nodes and set configurations")
		},
	}

	command.Flags().BoolVar(&flags.AllNodes, "all", false, "set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}

func deletePodsIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node) {
	if !flags.RunaiSystemWorker && !flags.CpuWorker && !flags.GpuWorker {
		return
	}
	log.Debugf("Deleting runai pods if needed")
	runaiPods, err := client.GetClientset().CoreV1().Pods("runai").List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to list pods from runai namespace")
		os.Exit(1)
	}

	for _, pod := range runaiPods.Items {
		deletePodIfNeeded(pod, nodesInCluster, client)
	}
}

func deletePodIfNeeded(pod v1.Pod, nodesInCluster map[string]v1.Node, client *client.Client) {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		return
	}
	for _, nodeSelectorTerms := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, matchExpressions := range nodeSelectorTerms.MatchExpressions {
			if checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, systemWorkerLabel) {
				return
			}
			if checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, cpuWorkerLabel) {
				return
			}
			if checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, gpuWorkerLabel) {
				return
			}
		}
	}
}

func checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod v1.Pod, nodesInCluster map[string]v1.Node, client *client.Client, matchExpressions v1.NodeSelectorRequirement, labelToCheck string) bool {
	if matchExpressions.Key == labelToCheck {
		if len(pod.Spec.NodeName) == 0 {
			return true
		}
		_, found := nodesInCluster[pod.Spec.NodeName].Labels[labelToCheck]
		if found {
			return true
		}
		err := client.GetClientset().CoreV1().Pods("runai").Delete(pod.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Debugf("Failed to delete pod: %v, error: %v", pod.Name, err)
		}
		log.Debugf("Deleted runai pod: %v", pod.Name)
	}
	return false
}

func deleteResourcesIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node) {
	log.Info("Updating RunAi resources")
	deletePVCIfNeeded(flags, client, nodesInCluster)
	deleteJobsIfNeeded(client)
	deletePodsIfNeeded(flags, client, nodesInCluster)
}

func deleteJobsIfNeeded(client *client.Client) {
	client.GetClientset().BatchV1().Jobs("runai").Delete("runai-db-migrations", &metav1.DeleteOptions{})
}

func deletePVCIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node) {
	if !flags.RunaiSystemWorker {
		return
	}
	pvc, err := client.GetClientset().CoreV1().PersistentVolumeClaims("runai").Get("data-runai-db-0", metav1.GetOptions{})
	if err != nil {
		log.Debugf("Failed to list PVCs from runai namespace")
		return
	}

	pvcNode, found := pvc.Annotations["volume.kubernetes.io/selected-node"]
	if !found {
		return
	}

	nodeInfo, found := nodesInCluster[pvcNode]
	if !found {
		fmt.Printf("Failed to find pvc node in cluster, node: %v\n", pvcNode)
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

func updateRunaiConfigurations(client *client.Client, flags nodeRoleTypes, nodesInCluster map[string]v1.Node) {
	log.Info("Updating RunAi configurations")
	nodeWithRestrictSchedulingExist := false
	nodeWithRestrictRunaiSystemExist := false
	for _, nodeInfo := range nodesInCluster {
		_, foundCpu := nodeInfo.Labels[cpuWorkerLabel]
		_, foundGpu := nodeInfo.Labels[gpuWorkerLabel]
		_, foundSystem := nodeInfo.Labels[systemWorkerLabel]
		if foundCpu || foundGpu {
			nodeWithRestrictSchedulingExist = foundCpu || foundGpu
		}
		if foundSystem {
			nodeWithRestrictRunaiSystemExist = foundSystem
		}
	}
	log.Debugf("nodes with cpu or gpu workers exists: %v", nodeWithRestrictSchedulingExist)
	log.Debugf("nodes with runai system workers exists: %v", nodeWithRestrictRunaiSystemExist)
	updateRunaiConfigIfNeeded(client, flags, nodeWithRestrictSchedulingExist, nodeWithRestrictRunaiSystemExist)
	updateRunaiDeploymentIfNeeded(client, flags, nodeWithRestrictRunaiSystemExist)
}

func updateRunaiDeploymentIfNeeded(client *client.Client, flags nodeRoleTypes, nodeWithRestrictRunaiSystemExist bool) {
	if !flags.RunaiSystemWorker {
		return
	}
	deployment, err := client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
	if err != nil {
		fmt.Println("Failed to get runai-operator, error: %v", err)
		os.Exit(1)
	}
	if nodeWithRestrictRunaiSystemExist {
		deployment.Spec.Template.Spec.Affinity = &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      systemWorkerLabel,
									Operator: v1.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
		}
	} else {
		deployment.Spec.Template.Spec.Affinity = nil
	}

	_, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
	if err != nil {
		fmt.Println("Failed to update runai-operator, error: %v", err)
		os.Exit(1)
	}

	log.Debugf("Updated runai operator to have node affinity")
}

func updateRunaiConfigIfNeeded(client *client.Client, flags nodeRoleTypes, nodeWithRestrictSchedulingExist, nodeWithRestrictRunaiSystemExist bool) {
	runaiconfigResource := schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}
	runaiConfig, err := client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Get("runai", metav1.GetOptions{})
	var needToUpdateRunaiConfig bool
	if err != nil {
		fmt.Println("Failed to get RunaiConfig, RunAI isn't installed on the cluster")
		os.Exit(1)
	}
	nodeAffinityMap, _, err := unstructured.NestedMap(runaiConfig.Object, "spec.global.nodeAffinity")
	if err != nil {
		fmt.Printf("Failed to get nodeAffinityMap from runaiConfig, error: %v", err)
		os.Exit(1)
	}

	if flags.CpuWorker || flags.GpuWorker {
		nodeAffinityMap["restrictScheduling"] = nodeWithRestrictSchedulingExist != nodeAffinityMap["restrictScheduling"]
	}

	if flags.RunaiSystemWorker {

		nodeAffinityMap["restrictRunaiSystem"] = needToUpdateRunaiConfig || nodeWithRestrictSchedulingExist != nodeAffinityMap["restrictRunaiSystem"]

		if err != nil {
			fmt.Printf("Failed to get restrictRunaiSystem from runaiConfig, error: %v", err)
			os.Exit(1)
		}

	}

	if needToUpdateRunaiConfig {
		err = unstructured.SetNestedMap(runaiConfig.Object, nodeAffinityMap, "spec.global.nodeAffinity")
		_, err := client.GetDynamicClient().Resource(runaiconfigResource).Namespace("runai").Update(runaiConfig, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("Failed to update runaiconfig")
			os.Exit(1)
		}
		log.Debugf("Updated RunaiConfig")
	}
}

func labelNodesWithRolesAndGetNodesInCluster(client *client.Client, flags nodeRoleTypes, args []string, shouldEnableLabel bool) map[string]v1.Node {
	log.Info("Updating nodes with roles")

	allNodeClusters := map[string]v1.Node{}
	nodesInCluster, err := client.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil || len(nodesInCluster.Items) == 0 {
		fmt.Println("Failed to list nodesInCluster")
		os.Exit(1)
	}

	wasAnyNodeUpdated := false
	if flags.AllNodes {
		for _, nodeInfo := range nodesInCluster.Items {
			updateLabelsSingleNode(&nodeInfo, flags, client, shouldEnableLabel)
			allNodeClusters[nodeInfo.Name] = nodeInfo
			wasAnyNodeUpdated = true
		}
		log.Debugf("Successfully updated all nodes with roles")
		return allNodeClusters
	}

	nodesToUpdateMap := map[string]bool{}
	wasNodeUpdated := map[string]bool{}
	for _, nodeToLabel := range args {
		nodesToUpdateMap[nodeToLabel] = true
		wasNodeUpdated[nodeToLabel] = false
	}
	for _, nodeInfo := range nodesInCluster.Items {
		if nodesToUpdateMap[nodeInfo.Name] {
			updateLabelsSingleNode(&nodeInfo, flags, client, shouldEnableLabel)
			wasAnyNodeUpdated = true
			wasNodeUpdated[nodeInfo.Name] = true
			allNodeClusters[nodeInfo.Name] = nodeInfo
		}
		allNodeClusters[nodeInfo.Name] = nodeInfo
	}

	for nodeName := range wasNodeUpdated {
		if !wasNodeUpdated[nodeName] {
			log.Infof("Node: %v was not found in cluster", nodeName)
		}
	}

	if !wasAnyNodeUpdated {
		log.Infof("All nodes are already updated")
		os.Exit(1)
	}

	return allNodeClusters
}

func updateLabelsSingleNode(nodeInfo *v1.Node, o nodeRoleTypes, client *client.Client, shouldEnableLabel bool) {
	if nodeInfo.Labels == nil {
		nodeInfo.Labels = map[string]string{}
	}
	if o.GpuWorker {
		if shouldEnableLabel {
			nodeInfo.Labels[gpuWorkerLabel] = ""
		} else {
			delete(nodeInfo.Labels, gpuWorkerLabel)
		}
	}
	if o.CpuWorker {
		if shouldEnableLabel {
			nodeInfo.Labels[cpuWorkerLabel] = ""
		} else {
			delete(nodeInfo.Labels, cpuWorkerLabel)
		}
	}
	if o.RunaiSystemWorker {
		if shouldEnableLabel {
			nodeInfo.Labels[systemWorkerLabel] = ""
		} else {
			delete(nodeInfo.Labels, systemWorkerLabel)
		}
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
			nodesInCluster := labelNodesWithRolesAndGetNodesInCluster(client, flags, args, false)
			updateRunaiConfigurations(client, flags, nodesInCluster)
			deleteResourcesIfNeeded(flags, client, nodesInCluster)

			log.Infof("Successfully update nodes with roles")
		},
	}

	command.Flags().BoolVar(&flags.AllNodes, "all", false, "set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}
