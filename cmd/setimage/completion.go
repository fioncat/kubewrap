package setimage

import (
	"github.com/fioncat/kubewrap/cmd"
	"github.com/spf13/cobra"
)

func CompletionFunc(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return cmd.CompleteContainer(c, toComplete)
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
