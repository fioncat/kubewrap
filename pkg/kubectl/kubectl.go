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

	ListResources(resourceType, namespace string) ([]*Resource, error)
	ListContainers(r *Resource) ([]*Container, error)

	SetImage(c *Container, image string) error
	Scale(r *Resource, replicas int) error
	RolloutRestart(r *Resource) error
}

type Node struct {
	Name        string
	Description string
}

type Resource struct {
	Type      string
	Namespace string
	Name      string
}

func (r *Resource) String() string {
	return fmt.Sprintf("%s %s/%s", r.Type, r.Namespace, r.Name)
}

type Container struct {
	Resource
	ContainerName string
}

func (c *Container) String() string {
	return fmt.Sprintf("%s %s/%s/%s", c.Type, c.Namespace, c.Name, c.ContainerName)
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
