package secret

import (
	"fmt"
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nodeRoleTypes struct {
	ClusterWide bool
}

const (
	clusterWideSecretLabel = "runai/cluster-wide"
)

func Set() *cobra.Command {
	flags := nodeRoleTypes{}
	var command = &cobra.Command{
		Use:     "secret SECRET_NAME",
		Aliases: []string{"secrets"},
		Short:   "Set Secret resource",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().NFlag() == 0 {
				fmt.Println("No flags were provided")
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			client := client.GetClient()
			if flags.ClusterWide {
				updateSecrets(client, args, true)
				fmt.Println("Successfully set cluster wide settings to secrets")
			}
		},
	}

	command.Flags().BoolVar(&flags.ClusterWide, "cluster-wide", false, "set Secret as cluster wide")
	return command
}

func Remove() *cobra.Command {
	flags := nodeRoleTypes{}
	var command = &cobra.Command{
		Use:     "secret SECRET_NAME",
		Aliases: []string{"secrets"},
		Short:   "Remove Secret resource",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().NFlag() == 0 {
				fmt.Println("No flags were provided")
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			client := client.GetClient()
			if flags.ClusterWide {
				updateSecrets(client, args, false)
				fmt.Println("Successfully removed cluster wide settings from secrets")
			}
		},
	}

	command.Flags().BoolVar(&flags.ClusterWide, "cluster-wide", false, "set Secret as cluster wide")
	return command
}

func updateSecrets(client *client.Client, args []string, shouldAddSecret bool) {
	secretList, err := client.GetClientset().CoreV1().Secrets("runai").List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list all secrets from RunAi Namespace, error: %v", err)
		os.Exit(1)
	}

	secretsToUpdateMap := map[string]bool{}
	for _, secretsToLabel := range args {
		secretsToUpdateMap[secretsToLabel] = false
	}
	for _, secretInfo := range secretList.Items {
		if _, found := secretsToUpdateMap[secretInfo.Name]; found {
			if secretInfo.Labels == nil {
				secretInfo.Labels = map[string]string{}
			}
			delete(secretInfo.Labels, clusterWideSecretLabel)
			if shouldAddSecret {
				secretInfo.Labels[clusterWideSecretLabel] = "true"
			} else {
				delete(secretInfo.Labels, clusterWideSecretLabel)
			}
			secretsToUpdateMap[secretInfo.Name] = true
			_, err = client.GetClientset().CoreV1().Secrets("runai").Update(&secretInfo)
			log.Debugf("Updated secret: %v", secretInfo.Name)
		}
	}

	for secretName, value := range secretsToUpdateMap {
		if !value {
			log.Infof("Secret: %v doesn't exist", secretName)
		}
	}
}
