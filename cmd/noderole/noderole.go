package noderole

import (
	"fmt"
	"os"
	"reflect"

	"github.com/run-ai/runai-cli/cmd/common"
	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
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
	withBackend := false
	var command = &cobra.Command{
		Use:     "node-role NODE_NAME",
		Aliases: []string{"node-roles"},
		Short:   "Set node with roles",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && !flags.AllNodes {
				fmt.Println("No nodes were selected")
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			client := client.GetClient()
			nodesInCluster := labelNodesWithRolesAndGetNodesInCluster(client, flags, args, true)
			updateRunaiConfigurations(client, flags, nodesInCluster, withBackend)

			log.Info("Successfully updated nodes and set configurations")
		},
	}

	command.Flags().BoolVar(&withBackend, "with-backend", false, "Update backend pods (In Air-gapped environment)")
	command.Flags().BoolVar(&flags.AllNodes, "all", false, "Set all nodes.")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "Set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "Set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "Set nodes with node-role of Run:AI System Worker.")
	return command
}

func deletePodsIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist bool, namespace string) {
	if !flags.RunaiSystemWorker && !flags.CpuWorker && !flags.GpuWorker {
		return
	}
	runaiPods, err := client.GetClientset().CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to list pods from the runai namespace")
		os.Exit(1)
	}

	for _, pod := range runaiPods.Items {
		deletePodIfNeeded(pod, nodesInCluster, client, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist, namespace)
	}
}

func deletePodIfNeeded(pod v1.Pod, nodesInCluster map[string]v1.Node, client *client.Client, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist bool, namespace string) {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil || pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		return
	}
	for _, nodeSelectorTerms := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, matchExpressions := range nodeSelectorTerms.MatchExpressions {
			if nodeWithRestrictRunaiSystemExist && checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, systemWorkerLabel, namespace) {
				return
			}
			if nodeWithRestrictSchedulingExist && checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, cpuWorkerLabel, namespace) {
				return
			}
			if nodeWithRestrictSchedulingExist && checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod, nodesInCluster, client, matchExpressions, gpuWorkerLabel, namespace) {
				return
			}
		}
	}
}

func checkIfLabelIsNotStatisfiedAndDeleteIfNeeded(pod v1.Pod, nodesInCluster map[string]v1.Node, client *client.Client, matchExpressions v1.NodeSelectorRequirement, labelToCheck, namespace string) bool {
	if matchExpressions.Key == labelToCheck {
		if len(pod.Spec.NodeName) == 0 {
			return true
		}
		_, found := nodesInCluster[pod.Spec.NodeName].Labels[labelToCheck]
		if found {
			return true
		}
		client.GetClientset().CoreV1().Pods(namespace).Delete(pod.Name, &metav1.DeleteOptions{})
		log.Debugf("Deleted Run:AI pod: %v", pod.Name)
	}
	return false
}

func deleteResourcesIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist, deleteStsAndPvc bool, namespace string) {
	log.Info("Deleting old Run:AI resources")
	if deleteStsAndPvc {
		deletePVCAndStsIfNeeded(flags, client, nodesInCluster, nodeWithRestrictRunaiSystemExist, namespace)
	}
	deleteJobsIfNeeded(client, namespace)
	deletePodsIfNeeded(flags, client, nodesInCluster, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist, namespace)
}

func deleteJobsIfNeeded(client *client.Client, namespace string) {
	jobs, err := client.GetClientset().BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list jobs, error: %v", err)
		os.Exit(1)
	}
	for _, job := range jobs.Items {
		client.GetClientset().BatchV1().Jobs(namespace).Delete(job.Name, &metav1.DeleteOptions{})
		log.Debugf("Deleted Job: %v", job.Name)
	}
}

func deletePVCAndStsIfNeeded(flags nodeRoleTypes, client *client.Client, nodesInCluster map[string]v1.Node, nodeWithRestrictRunaiSystemExist bool, namespace string) {
	if !flags.RunaiSystemWorker || !nodeWithRestrictRunaiSystemExist {
		return
	}

	pvc, err := client.GetClientset().CoreV1().PersistentVolumeClaims(namespace).Get("data-runai-db-0", metav1.GetOptions{})
	pvcNode, found := pvc.Annotations["volume.kubernetes.io/selected-node"]
	if found {
		nodeInfo, found := nodesInCluster[pvcNode]
		if !found {
			fmt.Printf("Failed to find PVC node in cluster, node: %v\n", pvcNode)
			os.Exit(1)
		}

		if _, found := nodeInfo.Labels[systemWorkerLabel]; found { // no need to delete the pvc - already on a system node
			return
		}

		client.GetClientset().CoreV1().PersistentVolumeClaims(namespace).Delete("data-runai-db-0", &metav1.DeleteOptions{})
		log.Debugf("Deleted PVC data-runai-db-0")
	}

	stsList, err := client.GetClientset().AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Debugf("Failed to list statefulsets in the %s namespace", namespace)
		return
	}

	for _, sts := range stsList.Items {
		client.GetClientset().AppsV1().StatefulSets(namespace).Delete(sts.Name, &metav1.DeleteOptions{})
		log.Debugf("Deleted Statefulset: %v", sts.Name)
	}
}

func updateRunaiConfigurations(client *client.Client, flags nodeRoleTypes, nodesInCluster map[string]v1.Node, withBackend bool) {
	log.Info("Updating Run:AI configurations")
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
	log.Debugf("Nodes with cpu or gpu workers already exist: %v", nodeWithRestrictSchedulingExist)
	log.Debugf("Nodes with runai system workers already exist: %v", nodeWithRestrictRunaiSystemExist)
	common.ScaleRunaiOperator(client, 0)
	updateDeploymentWithAffinity(client, flags, common.RunaiNamespace, common.RunaiOperatorDeploymentName, nodeWithRestrictRunaiSystemExist)
	updateRunaiConfigIfNeeded(client, flags, nodeWithRestrictSchedulingExist, nodeWithRestrictRunaiSystemExist)
	deleteResourcesIfNeeded(flags, client, nodesInCluster, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist, true, common.RunaiNamespace)
	common.ScaleRunaiOperator(client, 1)

	if withBackend {
		common.ScaleRunaiBackendOperator(client, 0)
		updateDeploymentWithAffinity(client, flags, common.RunaiBackendNamespace, common.RunaiBackendOperatorDeploymentName, nodeWithRestrictRunaiSystemExist)
		updateHelmReleaseIfNeeded(client, flags, nodeWithRestrictRunaiSystemExist)
		deleteResourcesIfNeeded(flags, client, nodesInCluster, nodeWithRestrictRunaiSystemExist, nodeWithRestrictSchedulingExist, false, common.RunaiBackendNamespace)
		common.ScaleRunaiBackendOperator(client, 1)
	}
}

func updateDeploymentWithAffinity(client *client.Client, flags nodeRoleTypes, namespace, deploymentName string, nodeWithRestrictRunaiSystemExist bool) {
	if !flags.RunaiSystemWorker {
		return
	}

	var err error
	var deployment *appsv1.Deployment

	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		deployment, err = client.GetClientset().AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			log.Infof("Failed to get runai-operator, error: %v", err)
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
		_, err = client.GetClientset().AppsV1().Deployments(namespace).Update(deployment)
		if err != nil {
			log.Debugf("Failed to update the %s, attempt: %v error: %v", deploymentName, i, err)
			continue
		}
		break
	}
	if err != nil {
		log.Infof("Failed to update the %s, error: %v", deploymentName, err)
		os.Exit(1)
	}

	log.Debugf("Updated %s to have node affinity and scaled to 0 replicas", deploymentName)
}

func updateRunaiConfigIfNeeded(client *client.Client, flags nodeRoleTypes, nodeWithRestrictSchedulingExist, nodeWithRestrictRunaiSystemExist bool) {
	runaiconfigResource := schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}
	var error error
	var runaiConfig *unstructured.Unstructured
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		runaiConfig, error = client.GetDynamicClient().Resource(runaiconfigResource).Namespace(common.RunaiNamespace).Get("runai", metav1.GetOptions{})
		if error != nil {
			fmt.Println("Failed to get RunaiConfig, Run:AI is not installed on the cluster")
			os.Exit(1)
		}
		nodeAffinityMapOldValues, _, err := unstructured.NestedMap(runaiConfig.Object, "spec", "global", "nodeAffinity")
		log.Debugf("RunaiConfig old values of nodeAffinityMap: %v", nodeAffinityMapOldValues)

		nodeAffinityMap := map[string]interface{}{}
		for key, val := range nodeAffinityMapOldValues {
			nodeAffinityMap[key] = val
		}
		if err != nil {
			fmt.Printf("Failed to get nodeAffinityMap from runaiConfig, error: %v", err)
			os.Exit(1)
		}

		if flags.CpuWorker || flags.GpuWorker {
			nodeAffinityMap["restrictScheduling"] = nodeWithRestrictSchedulingExist
		}

		if flags.RunaiSystemWorker {
			nodeAffinityMap["restrictRunaiSystem"] = nodeWithRestrictRunaiSystemExist
		}

		if !reflect.DeepEqual(nodeAffinityMap, nodeAffinityMapOldValues) {
			log.Debugf("Updating RunaiConfig with nodeAffinityMap: %v", nodeAffinityMap)
			err = unstructured.SetNestedMap(runaiConfig.Object, nodeAffinityMap, "spec", "global", "nodeAffinity")
			_, error = client.GetDynamicClient().Resource(runaiconfigResource).Namespace(common.RunaiNamespace).Update(runaiConfig, metav1.UpdateOptions{})
			if error != nil {
				log.Debugf("Failed to update runaiconfig, attempt: %v, error: %v", i, error)
				continue
			}
		}
		break
	}

	if error != nil {
		log.Infof("Failed to update runaiconfig, error: %v", error)
		os.Exit(1)
	}
}

func updateHelmReleaseIfNeeded(client *client.Client, flags nodeRoleTypes, nodeWithRestrictRunaiSystemExist bool) {
	helmReleaseResource := schema.GroupVersionResource{Group: "helm.fluxcd.io", Version: "v1", Resource: "HelmRelease"}
	var error error
	var runaiBackendHelmRelease *unstructured.Unstructured
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		runaiBackendHelmRelease, error = client.GetDynamicClient().Resource(helmReleaseResource).Namespace(common.RunaiBackendNamespace).Get("runai-backend", metav1.GetOptions{})
		if error != nil {
			fmt.Println("Failed to get HelmRelease, Run:AI Backend is not installed on the cluster")
			os.Exit(1)
		}
		nodeAffinityMapOldValues, _, err := unstructured.NestedMap(runaiBackendHelmRelease.Object, "spec", "global", "nodeAffinity")
		log.Debugf("HelmRelease old values of nodeAffinityMap: %v", nodeAffinityMapOldValues)

		nodeAffinityMap := map[string]interface{}{}
		for key, val := range nodeAffinityMapOldValues {
			nodeAffinityMap[key] = val
		}
		if err != nil {
			fmt.Printf("Failed to get nodeAffinityMap from runaiBackendHelmRelease, error: %v", err)
			os.Exit(1)
		}

		if flags.RunaiSystemWorker {
			nodeAffinityMap["restrictRunaiSystem"] = nodeWithRestrictRunaiSystemExist
		}

		if !reflect.DeepEqual(nodeAffinityMap, nodeAffinityMapOldValues) {
			log.Debugf("Updating HelmRelease with nodeAffinityMap: %v", nodeAffinityMap)
			err = unstructured.SetNestedMap(runaiBackendHelmRelease.Object, nodeAffinityMap, "spec", "global", "nodeAffinity")
			_, error = client.GetDynamicClient().Resource(helmReleaseResource).Namespace(common.RunaiNamespace).Update(runaiBackendHelmRelease, metav1.UpdateOptions{})
			if error != nil {
				log.Debugf("Failed to update HelmRelease, attempt: %v, error: %v", i, error)
				continue
			}
		}
		break
	}

	if error != nil {
		log.Infof("Failed to update HelmRelease, error: %v", error)
		os.Exit(1)
	}
}

func labelNodesWithRolesAndGetNodesInCluster(client *client.Client, flags nodeRoleTypes, args []string, shouldEnableLabel bool) map[string]v1.Node {
	log.Info("Updating nodes with roles")

	allNodeClusters := map[string]v1.Node{}
	nodesInCluster, err := client.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil || len(nodesInCluster.Items) == 0 {
		fmt.Println("Failed to list nodes in cluster")
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

func updateLabelsSingleNode(nodeInfo *v1.Node, flags nodeRoleTypes, client *client.Client, shouldEnableLabel bool) {
	var err error
	for i := 0; i < common.NumberOfRetiresForApiServer; i++ {
		if nodeInfo.Labels == nil {
			nodeInfo.Labels = map[string]string{}
		}
		if flags.GpuWorker {
			if shouldEnableLabel {
				nodeInfo.Labels[gpuWorkerLabel] = ""
			} else {
				delete(nodeInfo.Labels, gpuWorkerLabel)
			}
		}
		if flags.CpuWorker {
			if shouldEnableLabel {
				nodeInfo.Labels[cpuWorkerLabel] = ""
			} else {
				delete(nodeInfo.Labels, cpuWorkerLabel)
			}
		}
		if flags.RunaiSystemWorker {
			if shouldEnableLabel {
				nodeInfo.Labels[systemWorkerLabel] = ""
			} else {
				delete(nodeInfo.Labels, systemWorkerLabel)
			}
		}
		_, err = client.GetClientset().CoreV1().Nodes().Update(nodeInfo)
		if err == nil {
			break
		}
		nodeInfo, _ = client.GetClientset().CoreV1().Nodes().Get(nodeInfo.Name, metav1.GetOptions{})

		log.Debugf("Failed to update node, attempt: %v, error: %v", i, err)
	}
	if err != nil {
		log.Infof("Failed to update node: %v, ", err)
		os.Exit(1)
	}
}

func Remove() *cobra.Command {
	flags := nodeRoleTypes{}
	withBackend := false
	var command = &cobra.Command{
		Use:     "node-role NODE_NAME",
		Aliases: []string{"node-roles"},
		Short:   "Remove node with roles",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && !flags.AllNodes {
				fmt.Println("No nodes were selected")
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			client := client.GetClient()
			nodesInCluster := labelNodesWithRolesAndGetNodesInCluster(client, flags, args, false)
			updateRunaiConfigurations(client, flags, nodesInCluster, withBackend)
			log.Infof("Successfully updated nodes with roles")
		},
	}

	command.Flags().BoolVar(&withBackend, "with-backend", false, "Update backend pods (In Air-gapped environment)")
	command.Flags().BoolVar(&flags.AllNodes, "all", false, "Set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "Set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "Set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "Set nodes with node-role of Run:AI System Worker.")
	return command
}
