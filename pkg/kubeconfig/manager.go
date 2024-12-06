package kubeconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fioncat/kubewrap/pkg/dirs"
)

type manager struct {
	root string

	current *KubeConfig
	configs map[string]*KubeConfig
}

func NewManager(root string, alias map[string]string) (Manager, error) {
	mgr := &manager{
		root:    root,
		current: nil,
		configs: make(map[string]*KubeConfig),
	}
	err := mgr.init(alias)
	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func (m *manager) init(alias map[string]string) error {
	err := filepath.Walk(m.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		name, err := filepath.Rel(m.root, path)
		if err != nil {
			// This error should not happen in normal case. So we add a prefix to mark it.
			// If this happens, it means there is a bug in the code.
			return fmt.Errorf("[Internal] bad kubeconfig path %q, not in expected position", path)
		}

		m.configs[name] = &KubeConfig{
			root: m.root,

			Name:  name,
			Alias: "",
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read kubeconfig root: %w", err)
	}

	for alias, target := range alias {
		_, ok := m.configs[alias]
		if ok {
			return fmt.Errorf("alias %q is already used by a kubeconfig", alias)
		}
		_, ok = m.configs[target]
		if !ok {
			return fmt.Errorf("alias %q target %q not found", alias, target)
		}
		m.configs[alias] = &KubeConfig{
			root: m.root,

			Name:  alias,
			Alias: target,
		}
	}

	currentName := os.Getenv(envName)
	if currentName != "" {
		currentConfig, ok := m.configs[currentName]
		if !ok {
			return fmt.Errorf("current kubeconfig %q not found, please unuse it", currentName)
		}
		m.current = currentConfig
	}

	return nil
}

func (m *manager) Put(name string, data []byte) error {
	config, ok := m.configs[name]
	if !ok {
		config = &KubeConfig{
			root: m.root,
			Name: name,
		}
		m.configs[name] = config
	}
	path := config.Path()
	err := dirs.EnsureCreate(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("ensure kubeconfig dir: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("write kubeconfig file: %w", err)
	}
	return nil
}

func (m *manager) Delete(name string) error {
	if m.current != nil && m.current.Name == name {
		return errors.New("cannot delete current kubeconfig, please unuse it first")
	}
	for _, kubeconfig := range m.configs {
		if kubeconfig.Alias == name {
			return errors.New("this kubeconfig is used by an alias, please delete the alias in config file first")
		}
	}

	config, ok := m.configs[name]
	if !ok {
		return fmt.Errorf("kubeconfig %q not found", name)
	}

	if config.Alias != "" {
		return errors.New("cannot delete an alias kubeconfig, please delete it from config file")
	}

	path := config.Path()
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete kubeconfig file: %w", err)
	}

	var dir string
	for {
		dir = filepath.Dir(path)
		if dir == m.root {
			break
		}

		ents, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(ents) > 0 {
			break
		}

		err = os.Remove(dir)
		if err != nil {
			return err
		}
	}

	delete(m.configs, name)
	return nil
}

func (m *manager) DeleteAll() error {
	if m.current != nil {
		return errors.New("cannot delete all kubeconfigs, please unuse the current kubeconfig first")
	}
	var hasAlias bool
	for _, kubeconfig := range m.configs {
		if kubeconfig.Alias != "" {
			hasAlias = true
			break
		}
	}
	if hasAlias {
		return errors.New("cannot delete all kubeconfigs, please delete the alias kubeconfigs first")
	}

	m.configs = nil
	return os.RemoveAll(m.root)
}

func (m *manager) Current() (*KubeConfig, bool) {
	return m.current, m.current != nil
}

func (m *manager) Get(name string) (*KubeConfig, bool) {
	config, ok := m.configs[name]
	return config, ok
}

func (m *manager) List() []*KubeConfig {
	var list []*KubeConfig
	for _, config := range m.configs {
		list = append(list, config)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}
