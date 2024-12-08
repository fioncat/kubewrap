package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	maxConfigHistoryMax = 500
	minConfigHistoryMax = 5
)

type Config struct {
	Command string `json:"cmd" toml:"cmd"`

	Editor string `json:"editor" toml:"editor"`

	SourceFilePath string `json:"source_file_path" toml:"source_file_path"`

	Kubectl    Kubectl    `json:"kubectl" toml:"kubectl"`
	NodeShell  NodeShell  `json:"nodeshell" toml:"nodeshell"`
	KubeConfig KubeConfig `json:"kubeconfig" toml:"kubeconfig"`
	History    History    `json:"history" toml:"history"`

	NamespaceAlias []NamespaceAlias `json:"namespace_alias" toml:"namespace_alias"`
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

type KubeConfig struct {
	Root  string            `json:"root" toml:"root"`
	Alias map[string]string `json:"alias" toml:"alias"`
}

type History struct {
	Path string `json:"path" toml:"path"`
	Max  int    `json:"max" toml:"max"`
}

type NamespaceAlias struct {
	Configs    []string `json:"configs" toml:"configs"`
	Namespaces []string `json:"namespaces" toml:"namespaces"`
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
	cfg, err := load(path, useDefault)
	if err != nil {
		return nil, err
	}

	err = cfg.normalize()
	if err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}

	return cfg, nil
}

func load(path string, useDefault bool) (*Config, error) {
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

	return &cfg, nil
}

func (c *Config) normalize() error {
	if len(c.Command) == 0 {
		c.Command = defaults.Command
	}

	if len(c.Editor) == 0 {
		c.Editor = defaults.Editor
	}
	c.Editor = os.ExpandEnv(c.Editor)
	if c.Editor == "" {
		c.Editor = "vim"
	}

	if len(c.SourceFilePath) == 0 {
		c.SourceFilePath = defaults.SourceFilePath
	}
	c.SourceFilePath = os.ExpandEnv(c.SourceFilePath)
	if !filepath.IsAbs(c.SourceFilePath) {
		return errors.New("`source_file_path` is not absolute")
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

	if len(c.KubeConfig.Root) == 0 {
		c.KubeConfig.Root = defaults.KubeConfig.Root
	}
	c.KubeConfig.Root = os.ExpandEnv(c.KubeConfig.Root)
	if !filepath.IsAbs(c.KubeConfig.Root) {
		return errors.New("`kubeconfig.root` is not absolute")
	}

	if len(c.History.Path) == 0 {
		c.History.Path = defaults.History.Path
	}
	c.History.Path = os.ExpandEnv(c.History.Path)

	if c.History.Max <= 0 {
		c.History.Max = defaults.History.Max
	}
	if c.History.Max < minConfigHistoryMax {
		return fmt.Errorf("`history.max` is too small, should be >= %d", minConfigHistoryMax)
	}
	if c.History.Max > maxConfigHistoryMax {
		return fmt.Errorf("`history.max` is too large, should be <= %d", maxConfigHistoryMax)
	}

	return nil
}
