package scale

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "scale <QUERY> <REPLICAS>",
		Short: "Scale the replicas of a resource",
		Args:  cobra.ExactArgs(2),

		ValidArgsFunction: CompletionFunc,
	}
	return cmd.Build(c, &opts)
}

type Options struct {
	query    string
	replicas int
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	o.query = args[0]

	var err error
	o.replicas, err = strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("replicas must be an integer: %w", err)
	}
	if o.replicas < 0 {
		return errors.New("replicas must be greater than or equal to 0")
	}

	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	r, err := cmd.SelectResource(cmdctx.Kubectl, o.query)
	if err != nil {
		return err
	}
	return cmdctx.Kubectl.Scale(r, o.replicas)
}
