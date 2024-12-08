package source

import (
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/source"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "source",
		Short: "Print source content (please don't use directly)",
		Args:  cobra.NoArgs,
	}

	c.Flags().BoolVarP(&opts.noDelete, "no-delete", "", false, "don't delete source file")

	return cmd.Build(c, &opts)
}

type Options struct {
	noDelete bool
}

func (o *Options) Validate(_ *cobra.Command, _ []string) error { return nil }

func (o *Options) Run(cmdctx *cmd.Context) error {
	src, err := source.Get(cmdctx.Config, o.noDelete)
	if err != nil {
		return err
	}
	fmt.Println(src)
	return nil
}
