package kubectl

import "fmt"

type Kubectl interface {
	CheckNode(name string) error
	ListNodes() ([]*Node, error)

	CheckNamespace(name string) error
	ListNamespaces() ([]string, error)

	Apply(data []byte) error
	DeletePod(namespace, name string) error

	GetPodStatus(namespace, name string) (string, error)

	Exec(namespace, name string, cmd []string) error
	Copy(namespace, src, dest string) error
}

type Node struct {
	Name        string
	Description string
}

type NotFoundError struct {
	resourceType string
	name         string
}

func newNotFoundError(resourceType, name string) error {
	return &NotFoundError{resourceType: resourceType, name: name}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("cannot find %s %q", e.resourceType, e.name)
}

func IsNotFound(err error) bool {
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}
