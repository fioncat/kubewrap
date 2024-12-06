package nodeshell

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fioncat/kubewrap/pkg/kubectl"
)

//go:embed nodeshell.yaml
var yamlData []byte

const (
	checkPodReadyInterval = time.Millisecond * 300
	checkPodReadyTimeout  = time.Second * 10
)

func generateYAML(namespace, name, node, image string) []byte {
	data := bytes.ReplaceAll(yamlData, []byte("{{name}}"), []byte(name))
	data = bytes.ReplaceAll(data, []byte("{{namespace}}"), []byte(namespace))
	data = bytes.ReplaceAll(data, []byte("{{node}}"), []byte(node))
	data = bytes.ReplaceAll(data, []byte("{{image}}"), []byte(image))
	return data
}

type NodeShell struct {
	node string

	podName      string
	podNamespace string

	image string

	shell []string

	kubectl kubectl.Kubectl
}

type CopyPath struct {
	Path   string
	Remote bool
}

func New(kubectl kubectl.Kubectl, node, namespace, image string, shell []string) (*NodeShell, error) {
	err := kubectl.CheckNode(node)
	if err != nil {
		return nil, err
	}

	err = kubectl.CheckNamespace(namespace)
	if err != nil {
		return nil, err
	}

	safeNode := strings.ReplaceAll(node, ".", "-")
	podName := fmt.Sprintf("nodeshell-%s-%s", safeNode, genRandomName(5))

	ns := &NodeShell{
		node:         node,
		podName:      podName,
		podNamespace: namespace,
		image:        image,
		shell:        shell,
		kubectl:      kubectl,
	}
	err = ns.start()
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (n *NodeShell) start() error {
	yaml := generateYAML(n.podNamespace, n.podName, n.node, n.image)
	err := n.kubectl.Apply(yaml)
	if err != nil {
		return fmt.Errorf("nodeshell: create pod: %w", err)
	}

	checkInterval := time.NewTicker(checkPodReadyInterval)
	checkTimeout := time.NewTimer(checkPodReadyTimeout)

	var status string = "Unknown"
	for {
		select {
		case <-checkInterval.C:
			status, err = n.kubectl.GetPodStatus(n.podNamespace, n.podName)
			if err != nil {
				return fmt.Errorf("nodeshell check ready: get pod status: %w", err)
			}

			if status == "Running" {
				return nil
			}

		case <-checkTimeout.C:
			err = n.Close()
			if err != nil {
				return fmt.Errorf("delete pod after wait nodeshell pod ready timeout: %w", err)
			}

			return fmt.Errorf("wait nodeshell pod ready timeout after %v, please check its status (the last status is %q)", checkPodReadyTimeout, status)
		}
	}
}

func (n *NodeShell) Login() error {
	return n.kubectl.Exec(n.podNamespace, n.podName, n.shell)
}

func (n *NodeShell) Exec(cmd []string) error {
	return n.kubectl.Exec(n.podNamespace, n.podName, cmd)
}

func (n *NodeShell) Copy(src, dest CopyPath) error {
	if src.Remote && dest.Remote {
		return errors.New("copy: both src and dest are remote")
	}
	if !src.Remote && !dest.Remote {
		return errors.New("copy: both src and dest are local")
	}

	srcPath := src.Path
	if src.Remote {
		srcPath = fmt.Sprintf("%s:%s", n.podName, src.Path)
	}

	destPath := dest.Path
	if dest.Remote {
		destPath = fmt.Sprintf("%s:%s", n.podName, dest.Path)
	}

	return n.kubectl.Copy(n.podNamespace, srcPath, destPath)
}

func (n *NodeShell) Close() error {
	return n.kubectl.DeletePod(n.podNamespace, n.podName)
}

const randomCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

func genRandomName(length int) string {
	s := make([]byte, length)
	for i := 0; i < length; i++ {
		s[i] = randomCharset[rand.Intn(len(randomCharset))]
	}

	return string(s)
}
