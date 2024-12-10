package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/kubewrap/config"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/fioncat/kubewrap/pkg/kubectl"
	"github.com/spf13/cobra"
)

func GetCompleteConfig(c *cobra.Command) *config.Config {
	configPath := c.Flags().Lookup("config").Value.String()
	useDefaultConfig := c.Flags().Lookup("default-config").Value.String() == "true"
	cfg, err := config.Load(configPath, useDefaultConfig)
	if err != nil {
		WriteCompleteLogs("Load config failed: %v", err)
		return nil
	}
	return cfg
}

func GetCompleteKubeconfigManager(c *cobra.Command) kubeconfig.Manager {
	cfg := GetCompleteConfig(c)
	if cfg == nil {
		return nil
	}

	mgr, err := kubeconfig.NewManager(cfg.KubeConfig.Root, cfg.KubeConfig.Alias)
	if err != nil {
		WriteCompleteLogs("Create kubeconfig manager failed: %v", err)
		return nil
	}

	return mgr
}

func getCompleteKubectl(c *cobra.Command) kubectl.Kubectl {
	printConfig := c.Flags().Lookup("print-config").Value.String() == "true"
	if printConfig {
		WriteCompleteLogs("In print config mode, skip completion")
		return nil
	}

	cfg := GetCompleteConfig(c)
	if cfg == nil {
		return nil
	}

	return kubectl.NewCommand(cfg.Kubectl.Name, cfg.Kubectl.Args)
}

func CompleteNodeItems(c *cobra.Command) ([]string, bool) {
	nodes, ok := CompleteNodes(c)
	if !ok {
		return nil, false
	}
	items := make([]string, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, fmt.Sprintf("%s\t%s", node.Name, node.Description))
	}
	return items, true
}

func CompleteNodes(c *cobra.Command) ([]*kubectl.Node, bool) {
	kubectl := getCompleteKubectl(c)
	if kubectl == nil {
		return nil, false
	}

	nodes, err := kubectl.ListNodes()
	if err != nil {
		WriteCompleteLogs("List nodes failed: %v", err)
		return nil, false
	}
	return nodes, true
}

var resourceTypeCompletionList = []string{
	"deploy/", "sts/", "ds/", "job/", "cronjob/",
}

func CompleteResource(c *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResource(c, toComplete, false)
}

func completeResource(c *cobra.Command, toComplete string, fromContainer bool) ([]string, cobra.ShellCompDirective) {
	fields := strings.Split(toComplete, "/")
	switch len(fields) {
	case 0, 1:
		return resourceTypeCompletionList, cobra.ShellCompDirectiveNoSpace

	case 2:
		resourceType := fields[0]
		namespace := getCurrentNamespace()
		k := getCompleteKubectl(c)
		if k == nil {
			return nil, cobra.ShellCompDirectiveError
		}

		rs, err := k.ListResources(resourceType, namespace)
		if err != nil {
			WriteCompleteLogs("List resources failed: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}

		items := make([]string, 0, len(rs))
		for _, r := range rs {
			item := fmt.Sprintf("%s/%s", resourceType, r.Name)
			if fromContainer {
				item = item + "/"
			}
			items = append(items, item)
		}

		flag := cobra.ShellCompDirectiveNoFileComp
		if fromContainer {
			flag = cobra.ShellCompDirectiveNoSpace
		}

		return items, flag
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func CompleteContainer(c *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	fields := strings.Split(toComplete, "/")
	switch len(fields) {
	case 0, 1, 2:
		return completeResource(c, toComplete, true)

	case 3:
		r := &kubectl.Resource{
			Type:      fields[0],
			Namespace: getCurrentNamespace(),
			Name:      fields[1],
		}
		k := getCompleteKubectl(c)
		if k == nil {
			return nil, cobra.ShellCompDirectiveError
		}

		cs, err := k.ListContainers(r)
		if err != nil {
			WriteCompleteLogs("List containers failed: %v", err)
			return nil, cobra.ShellCompDirectiveError
		}

		items := make([]string, 0, len(cs))
		for _, container := range cs {
			items = append(items, fmt.Sprintf("%s/%s/%s", r.Type, r.Name, container.ContainerName))
		}

		return items, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func SingleNodeCompletionFunc(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	items, ok := CompleteNodeItems(c)
	if !ok {
		return nil, cobra.ShellCompDirectiveError
	}

	return items, cobra.ShellCompDirectiveNoFileComp
}

func WriteCompleteLogs(format string, args ...any) {
	logs := fmt.Sprintf(format+"\n", args...)
	path := filepath.Join(os.TempDir(), "kubewrap_complete.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(logs)
}
