package cluster

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/util/command"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func runSetCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.HelpFunc()(cmd, args)
		return nil
	} else if len(args) > 1 {
		return fmt.Errorf("Accepts 1 argument, received %d", len(args))
	}
	clusterName := args[0]

	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}

	config.CurrentContext = clusterName

	err = clientcmd.ModifyConfig(configAccess, *config, true)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}

	fmt.Printf("Set current cluster to %s \n", clusterName)
	return nil

}

func newSetClusterCommand() *cobra.Command {
	commandWrapper := command.NewCommandWrapper(runSetCommand)
	var command = &cobra.Command{
		Use:   "set [cluster]",
		Short: "Set current cluster",
		Run:   commandWrapper.Run,
	}

	return command
}
