package config

import (
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/spf13/cobra"
)

func CompletionFunc(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	mgr := cmd.GetCompleteKubeconfigManager(c)
	if mgr == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var curName string
	cur, ok := mgr.Current()
	if ok {
		curName = cur.Name
	}

	kcs := mgr.List()
	items := make([]string, 0, len(kcs))
	for _, kc := range kcs {
		if curName != "" && kc.Name == curName {
			continue
		}
		item := kc.Name
		if kc.Alias != "" {
			item = fmt.Sprintf("%s\talias to %s", kc.Name, kc.Alias)
		}
		items = append(items, item)
	}
	return items, cobra.ShellCompDirectiveNoFileComp
}
