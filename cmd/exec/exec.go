package exec

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
		Use:   "exec <NODE> -- <CMD>",
		Short: "Execute a command on a node",

		ValidArgsFunction: cmd.SingleNodeCompletionFunc,
	}
	return cmd.BuildNodeShell(c, &opts)
}

type Options struct {
	node string

	cmd []string
}

func (o *Options) Validate(c *cobra.Command, args []string) error {
	o.node = args[0]
	if len(o.node) == 0 {
		return errors.New("node is required")
	}

	argsAtDash := c.ArgsLenAtDash()
	if argsAtDash > -1 {
		o.cmd = args[argsAtDash:]
	}
	if len(o.cmd) == 0 {
		return errors.New("command is required")
	}

	return nil
}

func (o *Options) Node() string {
	return o.node
}

func (o *Options) Run(cmdctx *cmd.Context, nodeshell *nodeshell.NodeShell) error {
	term.PrintHint("Running command on %q", o.node)
	return nodeshell.Exec(o.cmd)
}
