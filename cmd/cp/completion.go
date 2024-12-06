package cp

import (
	"fmt"
	"strings"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/spf13/cobra"
)

func CompletionFunc(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	src := args[0]
	if strings.Contains(src, ":") {
		return nil, cobra.ShellCompDirectiveDefault
	}

	if strings.Contains(toComplete, ":") {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	nodes, ok := cmd.CompleteNodes(c)
	if !ok {
		return nil, cobra.ShellCompDirectiveError
	}

	items := make([]string, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, fmt.Sprintf("%s:/\t%s", node.Name, node.Description))
	}
	return items, cobra.ShellCompDirectiveNoSpace
}
