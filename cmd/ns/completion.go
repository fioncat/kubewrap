package ns

import (
	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/fioncat/kubewrap/pkg/kubectl"
	"github.com/spf13/cobra"
)

func CompletionFunc(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg := cmd.GetCompleteConfig(c)

	mgr, err := kubeconfig.NewManager(cfg.KubeConfig.Root, cfg.KubeConfig.Alias)
	if err != nil {
		cmd.WriteCompleteLogs("init kubeconfig manager failed: %v", err)
		return nil, cobra.ShellCompDirectiveError
	}

	cur, ok := mgr.Current()
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	kubectl := kubectl.NewCommand(cfg.Kubectl.Name, cfg.Kubectl.Args)

	items, err := listNamespaces(cfg, kubectl, cur.Name)
	if err != nil {
		cmd.WriteCompleteLogs("list namespaces: %v", err)
		return nil, cobra.ShellCompDirectiveError
	}

	return items, cobra.ShellCompDirectiveNoFileComp
}
