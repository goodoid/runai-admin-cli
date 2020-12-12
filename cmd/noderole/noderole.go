package noderole

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type nodeRoleTypes struct {
	CpuWorker         bool
	All               bool
	GpuWorker         bool
	RunaiSystemWorker bool
}

func Set() *cobra.Command {
	flags := nodeRoleTypes{}
	var command = &cobra.Command{
		Use:     "node-role NODE_NAME",
		Aliases: []string{"node-roles"},
		Short:   "Set node with roles",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := client.GetClient()
			nodesInCluster, err := client.GetClientset().CoreV1().Nodes().List(metav1.ListOptions{})
			if err != nil || len(nodesInCluster.Items) == 0 {
				fmt.Println("Failed to list nodesInCluster")
				os.Exit(1)
			}

			if flags.All {
				for _, nodeInfo := range nodesInCluster.Items {
					setLabelsSingleNode(&nodeInfo, flags, client)
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

					setLabelsSingleNode(&nodeInfo, flags, client)
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

	command.Flags().BoolVar(&flags.All, "all", false, "set all nodes")
	command.Flags().BoolVar(&flags.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&flags.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&flags.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}

func setLabelsSingleNode(nodeInfo *v1.Node, o nodeRoleTypes, client *client.Client) {
	if nodeInfo.Labels == nil {
		nodeInfo.Labels = map[string]string{}
	}
	if o.GpuWorker {
		nodeInfo.Labels["node-role.kubernetes.io/runai-gpu-worker"] = ""
	}
	if o.CpuWorker {
		nodeInfo.Labels["node-role.kubernetes.io/runai-cpu-worker"] = ""
	}
	if o.RunaiSystemWorker {
		nodeInfo.Labels["node-role.kubernetes.io/runai-system-worker"] = ""
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

			if flags.All {
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

	command.Flags().BoolVar(&flags.All, "all", false, "set all nodes")
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
		delete(nodeInfo.Labels, "node-role.kubernetes.io/runai-gpu-worker")
	}
	if flags.CpuWorker {
		delete(nodeInfo.Labels, "node-role.kubernetes.io/runai-cpu-worker")
	}
	if flags.RunaiSystemWorker {
		delete(nodeInfo.Labels, "node-role.kubernetes.io/runai-system-worker")
	}

	_, err := client.GetClientset().CoreV1().Nodes().Update(&nodeInfo)
	if err != nil {
		log.Infof("Failed to update node: %v, ", err)
		os.Exit(1)
	}
}
