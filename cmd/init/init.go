package init

import (
	"errors"
	"fmt"
	"os"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/hack"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "init <SHELL>",
		Short: "Print init script, you should source this in the profile",
		Args:  cobra.ExactArgs(1),
	}
	return cmd.Build(c, &opts)
}

type Options struct {
	shell   string
	cmdName string
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	o.shell = args[0]
	if len(o.shell) == 0 {
		return errors.New("shell is required")
	}
	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	name := cmdctx.Config.Command
	if len(o.cmdName) > 0 {
		name = o.cmdName
	}

	root := cmdctx.Command.Root()
	root.Use = name
	fmt.Println(hack.GetBash(name))

	switch o.shell {
	case "bash", "sh":
		return root.GenBashCompletionV2(os.Stdout, true)

	case "zsh":
		return root.GenZshCompletion(os.Stdout)

	case "fish":
		return root.GenFishCompletion(os.Stdout, true)

	default:
		return fmt.Errorf("unknown shell type: %s", o.shell)
	}
}
