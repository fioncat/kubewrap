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
	resouceType string
	name        string
}

func newNotFoundError(resouceType, name string) error {
	return &NotFoundError{resouceType: resouceType, name: name}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("cannot find %s %q", e.resouceType, e.name)
}

func IsNotFound(err error) bool {
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}
