package cp

import (
	"errors"
	"strings"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/nodeshell"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "cp <SRC> <DEST>",
		Short: "Use nodeshell to copy files between local and remote nodes",
		Args:  cobra.ExactArgs(2),

		ValidArgsFunction: CompletionFunc,
	}
	return cmd.BuildNodeShell(c, &opts)
}

type Options struct {
	node string
	src  nodeshell.CopyPath
	dest nodeshell.CopyPath
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	src, ok := o.parseCopy(args[0])
	if !ok {
		return errors.New("invalid source path")
	}

	dest, ok := o.parseCopy(args[1])
	if !ok {
		return errors.New("invalid destination path")
	}

	if len(o.node) == 0 {
		return errors.New("require at least one remote copy path")
	}

	if src.Remote && dest.Remote {
		return errors.New("cannot copy between two remote nodes")
	}

	o.src = src
	o.dest = dest

	return nil
}

func (o *Options) Node() string {
	return o.node
}

func (o *Options) Run(cmdctx *cmd.Context, nodeshell *nodeshell.NodeShell) error {
	term.PrintHint("Copying between host and %q", o.node)
	return nodeshell.Copy(o.src, o.dest)
}

func (o *Options) parseCopy(arg string) (nodeshell.CopyPath, bool) {
	fields := strings.Split(arg, ":")
	if len(fields) == 0 {
		return nodeshell.CopyPath{}, false
	}
	if len(fields) == 1 {
		return nodeshell.CopyPath{Path: arg}, true
	}

	node := fields[0]
	if len(node) == 0 {
		return nodeshell.CopyPath{}, false
	}
	o.node = node

	path := strings.TrimPrefix(arg, node+":")
	if len(path) == 0 {
		return nodeshell.CopyPath{}, false
	}

	return nodeshell.CopyPath{
		Path:   path,
		Remote: true,
	}, true
}
