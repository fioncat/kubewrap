package login

import (
	"errors"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/nodeshell"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "login <NODE>",
		Short: "Use nodeshell to login to a node",
		Args:  cobra.ExactArgs(1),

		ValidArgsFunction: cmd.SingleNodeCompletionFunc,
	}
	return cmd.BuildNodeShell(c, &opts)
}

type Options struct {
	node string
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	o.node = args[0]
	if len(o.node) == 0 {
		return errors.New("node is required")
	}

	return nil
}

func (o *Options) Node() string {
	return o.node
}

func (o *Options) Run(cmdctx *cmd.Context, nodeshell *nodeshell.NodeShell) error {
	term.PrintHint("Login to %q", o.Node())
	return nodeshell.Login()
}
