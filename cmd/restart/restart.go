package restart

import (
	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "restart <QUERY>",
		Short: "Restart a resource",
		Args:  cobra.ExactArgs(1),

		ValidArgsFunction: CompletionFunc,
	}
	return cmd.Build(c, &opts)
}

type Options struct {
	query string
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	o.query = args[0]
	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	r, err := cmd.SelectResource(cmdctx.Kubectl, o.query)
	if err != nil {
		return err
	}
	err = cmdctx.Kubectl.RolloutRestart(r)
	if err != nil {
		return err
	}

	term.PrintHint("Restarted %v", r)
	return nil
}
