package cmd

import (
	"fmt"

	"github.com/fioncat/kubewrap/config"
	"github.com/fioncat/kubewrap/pkg/kubectl"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

type Context struct {
	Command *cobra.Command
	Config  *config.Config
	Kubectl kubectl.Kubectl
}

type Validator interface {
	Validate(c *cobra.Command, args []string) error
}

type Options interface {
	Validator
	Run(cmdctx *Context) error
}

func Build(c *cobra.Command, opts Options) *cobra.Command {
	var (
		printConfig      bool
		configPath       string
		useDefaultConfig bool
	)

	c.RunE = func(cmd *cobra.Command, args []string) error {
		err := opts.Validate(cmd, args)
		if err != nil {
			return fmt.Errorf("validate command args: %w", err)
		}

		cfg, err := config.Load(configPath, useDefaultConfig)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if printConfig {
			return term.PrintJson(cfg)
		}

		kubectl := kubectl.NewCommand(cfg.Kubectl.Name, cfg.Kubectl.Args)
		cmdctx := &Context{
			Command: cmd,
			Config:  cfg,
			Kubectl: kubectl,
		}

		return opts.Run(cmdctx)
	}

	c.Flags().StringVarP(&configPath, "config", "", "", "config file path")
	c.Flags().BoolVarP(&useDefaultConfig, "default-config", "", false, "force to use default config")
	c.Flags().BoolVarP(&printConfig, "print-config", "", false, "print the config and exit (skip main process), useful for debug")

	return c
}
