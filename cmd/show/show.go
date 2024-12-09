package show

import (
	"errors"
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "show",
		Short: "Print current selected kubeconfig and namespace",
		Args:  cobra.NoArgs,
	}
	return cmd.Build(c, &opts)
}

type Options struct{}

func (o *Options) Validate(_ *cobra.Command, _ []string) error { return nil }

func (o *Options) Run(cmdctx *cmd.Context) error {
	cfg := cmdctx.Config
	mgr, err := kubeconfig.NewManager(cfg.KubeConfig.Root, cfg.KubeConfig.Alias)
	if err != nil {
		return err
	}

	cur, ok := mgr.Current()
	if !ok {
		return errors.New("no current selected kubeconfig")
	}

	ns := kubeconfig.GetCurrentNamespace()

	str := cur.String()
	if ns != "" {
		str = fmt.Sprintf("%s -> %s", str, ns)
	}

	fmt.Println(str)
	return nil
}
