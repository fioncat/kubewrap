package ns

import (
	"errors"
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/config"
	"github.com/fioncat/kubewrap/pkg/fzf"
	"github.com/fioncat/kubewrap/pkg/history"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/fioncat/kubewrap/pkg/kubectl"
	"github.com/fioncat/kubewrap/pkg/source"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "ns [NAME]",
		Short: "Switch to namespace",
		Args:  cobra.MaximumNArgs(1),

		ValidArgsFunction: CompletionFunc,
	}

	c.Flags().BoolVarP(&opts.unuse, "unuse", "u", false, "unuse namespace")
	c.Flags().BoolVarP(&opts.list, "list", "l", false, "list namespaces")
	c.Flags().BoolVarP(&opts.listHistory, "list-history", "H", false, "show namespace history")

	return cmd.Build(c, &opts)
}

type Options struct {
	namespace string

	unuse bool

	list        bool
	listHistory bool
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.namespace = args[0]
	}
	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	cfg := cmdctx.Config
	configMgr, err := kubeconfig.NewManager(cfg.KubeConfig.Root, cfg.KubeConfig.Alias)
	if err != nil {
		return err
	}

	histMgr, err := history.NewManager(cfg.History.Path, cfg.History.Max)
	if err != nil {
		return err
	}

	cur, ok := configMgr.Current()
	if !ok {
		return errors.New("no kubeconfig selected, cannot perform ns operations, please select one first")
	}

	if o.list {
		var nsList []string
		nsList, err = listNamespaces(cmdctx.Config, cmdctx.Kubectl, cur.Name)
		if err != nil {
			return err
		}
		for _, ns := range nsList {
			fmt.Println(ns)
		}
		return nil
	}

	if o.listHistory {
		records := histMgr.List()
		for _, record := range records {
			if record.Name != cur.Name {
				continue
			}
			if record.Namespace == "" {
				continue
			}
			fmt.Printf("[%s] %s\n", term.FormatTimestamp(record.Timestamp), record.Namespace)
		}
		return nil
	}

	if o.unuse {
		curNs := kubeconfig.GetCurrentNamespace()
		if curNs == "" {
			return errors.New("no current namespace used, cannot unuse")
		}
		term.PrintHint("Unuse current namespace %q", curNs)
		return source.Apply(cfg, cur.GenerateSource(""))
	}

	ns, err := o.selectNs(cmdctx, cur.Name, histMgr)
	if err != nil {
		return err
	}

	term.PrintHint("Switch to namespace %q", ns)
	err = source.Apply(cfg, cur.GenerateSource(ns))
	if err != nil {
		return err
	}

	histMgr.Add(cur.Name, ns)
	return histMgr.Save()
}

func (o *Options) selectNs(cmdctx *cmd.Context, curName string, histMgr history.Manager) (string, error) {
	if o.namespace == "-" {
		curNamespace := kubeconfig.GetCurrentNamespace()
		lastNsPtr := histMgr.GetLastNamespace(curName, curNamespace)
		if lastNsPtr == nil {
			return "", errors.New("no last namespace selected")
		}
		return *lastNsPtr, nil
	}

	if o.namespace != "" {
		return o.namespace, nil
	}

	items, err := listNamespaces(cmdctx.Config, cmdctx.Kubectl, curName)
	if err != nil {
		return "", err
	}
	idx, err := fzf.Search(items)
	if err != nil {
		return "", err
	}

	return items[idx], nil
}

func listNamespaces(cfg *config.Config, kubectl kubectl.Kubectl, curName string) ([]string, error) {
	nsList, err := listNamespacesRaw(cfg, kubectl, curName)
	if err != nil {
		return nil, err
	}

	newNsList := make([]string, 0, len(nsList))
	curNamespace := kubeconfig.GetCurrentNamespace()
	for _, ns := range nsList {
		if curNamespace != "" && ns == curNamespace {
			continue
		}
		newNsList = append(newNsList, ns)
	}
	return newNsList, nil
}

func listNamespacesRaw(cfg *config.Config, kubectl kubectl.Kubectl, curName string) ([]string, error) {
	for _, nsAlias := range cfg.NamespaceAlias {
		for _, name := range nsAlias.Configs {
			if name == curName {
				return nsAlias.Namespaces, nil
			}
		}
	}

	return kubectl.ListNamespaces()
}
