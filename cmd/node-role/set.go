package node_role

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type nodeRoleTypes struct {
	CpuWorker         bool
	GpuWorker         bool
	RunaiSystemWorker bool
}

func Set() *cobra.Command {
	o := nodeRoleTypes{}
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

			for _, nodeToLabel := range args {
				for _, nodeInfo := range nodesInCluster.Items {
					if nodeToLabel != nodeInfo.Name {
						continue
					}
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

					_, err = client.GetClientset().CoreV1().Nodes().Update(&nodeInfo)
					if err != nil {
						log.Infof("Failed to update node: %v, ", nodeInfo)
					}
				}
			}

			log.Infof("Successfully update nodes with roles")
		},
	}

	command.Flags().BoolVar(&o.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&o.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&o.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}

func Remove() *cobra.Command {
	o := nodeRoleTypes{}
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

			for _, nodeToLabel := range args {
				for _, nodeInfo := range nodesInCluster.Items {
					if nodeToLabel != nodeInfo.Name {
						continue
					}
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

					_, err = client.GetClientset().CoreV1().Nodes().Update(&nodeInfo)
					if err != nil {
						log.Infof("Failed to update node: %v, ", nodeInfo)
					}
				}
			}

			log.Infof("Successfully update nodes with roles")
		},
	}

	command.Flags().BoolVar(&o.CpuWorker, "cpu-worker", false, "set nodes with node-role of CPU Worker.")
	command.Flags().BoolVar(&o.GpuWorker, "gpu-worker", false, "set nodes with node-role of GPU Worker.")
	command.Flags().BoolVar(&o.RunaiSystemWorker, "runai-system-worker", false, "set nodes with node-role of Runai System Worker.")
	return command
}
