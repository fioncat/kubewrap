package kubectl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type cmdKubectl struct {
	name string
	args []string
}

func NewCommand(name string, args []string) Kubectl {
	return &cmdKubectl{name: name, args: args}
}

func (k *cmdKubectl) CheckNode(name string) error {
	nodes, err := k.ListNodes()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Name == name {
			return nil
		}
	}

	return newNotFoundError("node", name)
}

func (k *cmdKubectl) ListNodes() ([]*Node, error) {
	lines, err := k.lines("get", "nodes", "--no-headers")
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		desc := strings.TrimPrefix(line, name)
		nodes = append(nodes, &Node{
			Name:        name,
			Description: strings.TrimSpace(desc),
		})
	}
	return nodes, nil
}

func (k *cmdKubectl) CheckNamespace(name string) error {
	namespaces, err := k.ListNamespaces()
	if err != nil {
		return err
	}
	for _, ns := range namespaces {
		if ns == name {
			return nil
		}
	}
	return newNotFoundError("namespace", name)
}

func (k *cmdKubectl) ListNamespaces() ([]string, error) {
	output, err := k.output(nil, "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}
	return strings.Fields(output), nil
}

func (k *cmdKubectl) Apply(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := k.output(buf, "apply", "-f", "-")
	if err != nil {
		return err
	}
	return nil
}

func (k *cmdKubectl) DeletePod(namespace, name string) error {
	_, err := k.output(nil, "delete", "-n", namespace, "pod", name)
	return err
}

func (k *cmdKubectl) GetPodStatus(namespace, name string) (string, error) {
	status, err := k.output(nil, "get", "-n", namespace, "pod", name, "-o", "jsonpath={.status.phase}")
	if err != nil {
		return "", err
	}
	if len(status) == 0 {
		return "", errors.New("status returned by kubectl is empty, this is not expected")
	}
	return status, nil
}

func (k *cmdKubectl) Exec(namespace, name string, cmd []string) error {
	args := []string{"exec", "-it", "-n", namespace, name, "--"}
	args = append(args, cmd...)
	return k.exec(args, true, nil, nil)
}

func (k *cmdKubectl) Copy(namespace, src, dest string) error {
	args := []string{"cp", "-n", namespace, src, dest}
	_, err := k.output(nil, args...)
	return err
}

func (k *cmdKubectl) lines(args ...string) ([]string, error) {
	output, err := k.output(nil, args...)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (k *cmdKubectl) output(in io.Reader, args ...string) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := k.exec(args, false, in, buf)
	if err != nil {
		return "", err
	}
	output := buf.String()
	return strings.TrimSpace(output), nil
}

func (k *cmdKubectl) exec(args []string, tty bool, in io.Reader, out io.Writer) error {
	if len(k.args) > 0 {
		args = append(k.args, args...)
	}
	cmd := exec.Command(k.name, args...)
	cmd.Stderr = os.Stderr
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdin = in
		cmd.Stdout = out
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("kubectl command exited with bad status: %s %s", k.name, strings.Join(args, " "))
	}

	return nil
}
