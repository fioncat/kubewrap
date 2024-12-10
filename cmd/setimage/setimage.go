package setimage

import (
	"errors"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "set-image <QUERY> <IMAGE>",
		Short: "Set the image of a container",
		Args:  cobra.ExactArgs(2),

		ValidArgsFunction: CompletionFunc,
	}
	return cmd.Build(c, &opts)
}

type Options struct {
	query string
	image string
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	o.query = args[0]

	o.image = args[1]
	if len(o.image) == 0 {
		return errors.New("image is required")
	}

	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	c, err := cmd.SelectContainer(cmdctx.Kubectl, o.query)
	if err != nil {
		return err
	}
	err = cmdctx.Kubectl.SetImage(c, o.image)
	if err != nil {
		return err
	}

	term.PrintHint("Set image of %v to %q", c, o.image)
	return nil
}
