package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fioncat/kubewrap/pkg/nodeshell"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

type NodeShellOptions interface {
	Validator
	Node() string
	Run(cmdctx *Context, nodeshell *nodeshell.NodeShell) error
}

func BuildNodeShell(c *cobra.Command, opts NodeShellOptions) *cobra.Command {
	nsOpts := &nodeShellOptions{opts: opts}
	c.Flags().StringVarP(&nsOpts.namespace, "namespace", "n", "", "namespace of the shell pod, default will use option from config file")
	c.Flags().StringVarP(&nsOpts.image, "image", "i", "", "image of the shell pod, default will use option from config file")
	c.Flags().StringVarP(&nsOpts.shell, "shell", "s", "", "shell command to run, default will use option from config file")
	return Build(c, nsOpts)
}

type nodeShellOptions struct {
	opts NodeShellOptions

	namespace string
	image     string
	shell     string
}

func (o *nodeShellOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.opts.Validate(cmd, args)
}

func (o *nodeShellOptions) Run(cmdctx *Context) error {
	if len(o.namespace) == 0 {
		o.namespace = cmdctx.Config.NodeShell.Namespace
	}
	if len(o.image) == 0 {
		o.image = cmdctx.Config.NodeShell.Image
	}
	var shell []string
	if len(o.shell) == 0 {
		shell = cmdctx.Config.NodeShell.Shell
	} else {
		shell = strings.Fields(o.shell)
	}

	term.PrintHint("Spawning shell pod on %q", o.opts.Node())
	ns, err := nodeshell.New(cmdctx.Kubectl, o.opts.Node(), o.namespace, o.image, shell)
	if err != nil {
		return err
	}
	defer func() {
		term.PrintHint("Deleting shell pod on %q", o.opts.Node())
		closeErr := ns.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to delete shell pod on %q, you may need to delete it manually\n", o.opts.Node())
		}
	}()

	return o.opts.Run(cmdctx, ns)
}
