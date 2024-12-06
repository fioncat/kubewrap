package config

import (
	"errors"
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/fzf"
	"github.com/fioncat/kubewrap/pkg/history"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
)

type Options struct {
	name string

	edit      bool
	delete    bool
	deleteAll bool

	skipConfirm bool

	configMgr  kubeconfig.Manager
	historyMgr history.Manager
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	kc, err := o.selectKubeconfig()
	if err != nil {
		return err
	}

	return nil
}

func (o *Options) selectKubeconfig() (*kubeconfig.KubeConfig, error) {
	if o.name == "" {
		if o.edit {
			kc, ok := o.configMgr.Current()
			if !ok {
				return nil, errors.New("no kubeconfig is selected to edit")
			}
			return kc, nil
		}

		current, _ := o.configMgr.Current()
		kcs := o.configMgr.List()
		items := make([]string, 0, len(kcs))
		for _, kc := range kcs {
			if current != nil && kc.Name == current.Name {
				continue
			}
			items = append(items, kc.Name)
		}
		if len(items) == 0 {
			return nil, errors.New("no kubeconfig to select")
		}

		idx, err := fzf.Search(items)
		if err != nil {
			return nil, err
		}

		return kcs[idx], nil
	}

	if o.name == "-" {
		var current string
		if cur, ok := o.configMgr.Current(); ok {
			current = cur.Name
		}
		lastName := o.historyMgr.GetLastName(current)
		if lastName == nil {
			return nil, errors.New("no last selected kubeconfig")
		}
		name := *lastName
		kc, ok := o.configMgr.Get(name)
		if !ok {
			return nil, fmt.Errorf("last selected kubeconfig %q is not found, please clear history file", name)
		}
		return kc, nil
	}

	kc, _ := o.configMgr.Get(o.name)
	return kc, nil
}
