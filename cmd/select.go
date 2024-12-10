package cmd

import (
	"fmt"
	"strings"

	"github.com/fioncat/kubewrap/pkg/fzf"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/fioncat/kubewrap/pkg/kubectl"
)

func SelectResource(k kubectl.Kubectl, query string) (*kubectl.Resource, error) {
	fields := strings.Split(query, "/")
	if len(fields) != 1 && len(fields) != 2 {
		return nil, fmt.Errorf("invalid resource query %q, should be '<type>[/name]'", query)
	}

	resourceType := fields[0]
	if resourceType == "" {
		return nil, fmt.Errorf("invalid resource query %q, type is required", query)
	}

	namespace := getCurrentNamespace()

	var name string
	if len(fields) == 2 {
		name = fields[1]
	}
	if name == "" {
		rs, err := k.ListResources(resourceType, namespace)
		if err != nil {
			return nil, err
		}

		items := make([]string, 0, len(rs))
		for _, r := range rs {
			items = append(items, r.Name)
		}
		var idx int
		idx, err = fzf.Search(items)
		if err != nil {
			return nil, err
		}

		name = items[idx]
	}

	return &kubectl.Resource{
		Type:      resourceType,
		Namespace: namespace,
		Name:      name,
	}, nil
}

type selectContainerItem struct {
	key string

	container *kubectl.Container
}

func SelectContainer(k kubectl.Kubectl, query string) (*kubectl.Container, error) {
	fields := strings.Split(query, "/")
	if len(fields) != 1 && len(fields) != 2 && len(fields) != 3 {
		return nil, fmt.Errorf("invalid container query %q, should be '<type>[/<name>/<container>]'", query)
	}

	resourceType := fields[0]

	var name string
	if len(fields) > 1 {
		name = fields[1]
	}

	namespace := getCurrentNamespace()

	if name == "" {
		return selectContainerByResourceType(k, resourceType, namespace)
	}

	var containerName string
	if len(fields) == 3 {
		containerName = fields[2]
	}
	if containerName != "" {
		return &kubectl.Container{
			Resource: kubectl.Resource{
				Type:      resourceType,
				Namespace: namespace,
				Name:      name,
			},
			ContainerName: containerName,
		}, nil
	}

	r := &kubectl.Resource{
		Type:      resourceType,
		Namespace: namespace,
		Name:      name,
	}
	cs, err := k.ListContainers(r)
	if err != nil {
		return nil, err
	}
	if len(cs) == 0 {
		return nil, fmt.Errorf("no containers for %s %s/%s", r.Type, namespace, r.Name)
	}
	if len(cs) == 1 {
		return cs[0], nil
	}

	items := make([]string, 0, len(cs))
	for _, c := range cs {
		items = append(items, c.ContainerName)
	}
	idx, err := fzf.Search(items)
	if err != nil {
		return nil, err
	}
	return cs[idx], nil
}

func selectContainerByResourceType(k kubectl.Kubectl, resourceType, namespace string) (*kubectl.Container, error) {
	rs, err := k.ListResources(resourceType, namespace)
	if err != nil {
		return nil, err
	}
	citems := make([]*selectContainerItem, 0, len(rs))
	for _, r := range rs {
		var cs []*kubectl.Container
		cs, err = k.ListContainers(r)
		if err != nil {
			return nil, err
		}
		if len(cs) == 0 {
			continue
		}
		if len(cs) == 1 {
			citems = append(citems, &selectContainerItem{
				key:       r.Name,
				container: cs[0],
			})
			continue
		}
		for _, c := range cs {
			citems = append(citems, &selectContainerItem{
				key:       fmt.Sprintf("%s/%s", r.Name, c.ContainerName),
				container: c,
			})
		}
	}

	items := make([]string, 0, len(citems))
	for _, citem := range citems {
		items = append(items, citem.key)
	}
	idx, err := fzf.Search(items)
	if err != nil {
		return nil, err
	}

	return citems[idx].container, nil
}

func getCurrentNamespace() string {
	namespace := kubeconfig.GetCurrentNamespace()
	if namespace == "" {
		return "default"
	}
	return namespace
}
