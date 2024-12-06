package main

import (
	"fmt"
	"os"

	"github.com/fioncat/kubewrap/cmd/cp"
	"github.com/fioncat/kubewrap/cmd/exec"
	initcmd "github.com/fioncat/kubewrap/cmd/init"
	"github.com/fioncat/kubewrap/cmd/login"
	"github.com/spf13/cobra"
)

var (
	Version     string = "N/A"
	BuildType   string = "N/A"
	BuildCommit string = "N/A"
	BuildTime   string = "N/A"
)

func newCommand() *cobra.Command {
	var printBuildInfo bool

	c := &cobra.Command{
		Use:   "kubewrap",
		Short: "A wrapper for kubectl, to add more tools",

		Version: Version,

		SilenceErrors: true,
		SilenceUsage:  true,

		// Completion is impletemented by `init` command, so disable this
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if printBuildInfo {
				fmt.Printf("version: %s\n", Version)
				fmt.Printf("type:    %s\n", BuildType)
				fmt.Printf("commit:  %s\n", BuildCommit)
				fmt.Printf("time:    %s\n", BuildTime)
				return nil

			}
			return cmd.Usage()
		},
	}

	c.Flags().BoolVarP(&printBuildInfo, "build", "b", false, "print build information and exit")

	return c
}

func main() {
	c := newCommand()

	c.AddCommand(cp.New())
	c.AddCommand(exec.New())
	c.AddCommand(initcmd.New())
	c.AddCommand(login.New())

	err := c.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
