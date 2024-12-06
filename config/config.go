package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Command   string    `json:"cmd" toml:"cmd"`
	Kubectl   Kubectl   `json:"kubectl" toml:"kubectl"`
	NodeShell NodeShell `json:"nodeshell" toml:"nodeshell"`
}

type Kubectl struct {
	Name string   `json:"name" toml:"name"`
	Args []string `json:"args" toml:"args"`
}

type NodeShell struct {
	Namespace string   `json:"namespace" toml:"namespace"`
	Image     string   `json:"image" toml:"image"`
	Shell     []string `json:"shell" toml:"shell"`
}

//go:embed defaults.toml
var defaultsData []byte

var defaults = func() *Config {
	var cfg Config
	err := toml.Unmarshal(defaultsData, &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}()

func Load(path string, useDefault bool) (*Config, error) {
	if useDefault {
		return defaults, nil
	}

	var cfg Config
	if path != "" {
		_, err := toml.DecodeFile(path, &cfg)
		if err != nil {
			return nil, err
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(homeDir, ".config", "kubewrap", "config.toml")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return defaults, nil
			}
			return nil, err
		}

		err = toml.Unmarshal(data, &cfg)
		if err != nil {
			return nil, err
		}
	}

	err := cfg.normalize()
	if err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}

	return &cfg, nil

}

func (c *Config) normalize() error {
	if len(c.Command) == 0 {
		c.Command = defaults.Command
	}

	if len(c.Kubectl.Name) == 0 {
		c.Kubectl.Name = defaults.Kubectl.Name
	}

	if len(c.NodeShell.Namespace) == 0 {
		c.NodeShell.Namespace = defaults.NodeShell.Namespace
	}
	if len(c.NodeShell.Image) == 0 {
		c.NodeShell.Image = defaults.NodeShell.Image
	}
	if len(c.NodeShell.Shell) == 0 {
		c.NodeShell.Shell = defaults.NodeShell.Shell
	}

	return nil
}
