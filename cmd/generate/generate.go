package generate

import "github.com/spf13/cobra"

func Command() *cobra.Command {

	var command = &cobra.Command{
		Use:   "generate",
		Short: "Generate files and configs.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(KubeConfigGenerateCommand())

	return command
}
