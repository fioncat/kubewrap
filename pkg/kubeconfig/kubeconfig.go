package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envName      = "KUBECONFIG_NAME"
	envPath      = "KUBECONFIG" // this is used by `kubectl`
	envNamespace = "KUBECONFIG_NAMESPACE"
)

const sourceTemplate = `
export %s="%s"
export %s="%s"
export %s="%s"
alias k='kubectl%s'
`

const unsetTemplate = `
export %s=""
export %s=""
export %s=""
alias k='kubectl'
`

type Manager interface {
	Put(name string, data []byte) (*KubeConfig, error)
	Delete(name string) error
	DeleteAll() error

	Current() (*KubeConfig, bool)
	Get(name string) (*KubeConfig, bool)
	List() []*KubeConfig
}

type KubeConfig struct {
	root string

	Name  string
	Alias string
}

func (c *KubeConfig) Path() string {
	name := c.Name
	if c.Alias != "" {
		name = c.Alias
	}
	return filepath.Join(c.root, name)
}

func (c *KubeConfig) GenerateSource(ns string) string {
	var nsSet string
	if len(ns) > 0 {
		nsSet = fmt.Sprintf(" -n %s", ns)
	}
	source := fmt.Sprintf(sourceTemplate, envName, c.Name, envPath, c.Path(), envNamespace, ns, nsSet)
	return strings.TrimSpace(source)
}

func (c *KubeConfig) String() string {
	s := c.Name
	if c.Alias != "" {
		s = fmt.Sprintf("%s (alias to %s)", s, c.Alias)
	}
	return s
}

func UnsetSource() string {
	source := fmt.Sprintf(unsetTemplate, envName, envPath, envNamespace)
	return strings.TrimSpace(source)
}

func GetCurrentNamespace() string {
	return os.Getenv(envNamespace)
}
