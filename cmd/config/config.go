package config

import (
	"errors"
	"fmt"

	"github.com/fioncat/kubewrap/cmd"
	"github.com/fioncat/kubewrap/pkg/edit"
	"github.com/fioncat/kubewrap/pkg/fzf"
	"github.com/fioncat/kubewrap/pkg/history"
	"github.com/fioncat/kubewrap/pkg/kubeconfig"
	"github.com/fioncat/kubewrap/pkg/source"
	"github.com/fioncat/kubewrap/pkg/term"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts Options
	c := &cobra.Command{
		Use:   "config [NAME]",
		Short: "Manage kube config files",
		Args:  cobra.MaximumNArgs(1),

		ValidArgsFunction: CompletionFunc,
	}

	c.Flags().BoolVarP(&opts.edit, "edit", "e", false, "edit kubeconfig file")
	c.Flags().BoolVarP(&opts.delete, "delete", "d", false, "delete kubeconfig file")
	c.Flags().BoolVarP(&opts.deleteAll, "delete-all", "D", false, "delete all kubeconfig files")
	c.Flags().BoolVarP(&opts.list, "list", "l", false, "list kubeconfig files")
	c.Flags().BoolVarP(&opts.listHistory, "list-history", "H", false, "show kubeconfig history")
	c.Flags().BoolVarP(&opts.unuse, "unuse", "u", false, "unuse current kubeconfig")

	c.Flags().BoolVarP(&opts.skipConfirm, "noconfirm", "y", false, "skip confirm")

	return cmd.Build(c, &opts)
}

type Options struct {
	name string

	edit      bool
	delete    bool
	deleteAll bool

	list        bool
	listHistory bool

	unuse bool

	skipConfirm bool

	configMgr  kubeconfig.Manager
	historyMgr history.Manager
}

func (o *Options) Validate(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.name = args[0]
	}

	opts := []bool{
		o.edit, o.delete, o.deleteAll, o.list, o.listHistory, o.unuse,
	}
	var hasMode bool
	for _, opt := range opts {
		if opt {
			hasMode = true
			continue
		}
		if hasMode && opt {
			return errors.New("mode cannot duplicate")
		}
	}

	return nil
}

func (o *Options) Run(cmdctx *cmd.Context) error {
	err := o.prepare(*cmdctx)
	if err != nil {
		return err
	}

	switch {
	case o.edit:
		return o.handleEdit(cmdctx)
	case o.delete:
		return o.handleDelete()
	case o.deleteAll:
		return o.configMgr.DeleteAll()
	case o.list:
		return o.handleList()
	case o.listHistory:
		return o.handleListHistory()
	case o.unuse:
		return o.handleUnuse(cmdctx)
	default:
		return o.handleUse(cmdctx)
	}
}

func (o *Options) prepare(cmdctx cmd.Context) error {
	cfg := cmdctx.Config
	configMgr, err := kubeconfig.NewManager(cfg.KubeConfig.Root, cfg.KubeConfig.Alias)
	if err != nil {
		return err
	}

	histMgr, err := history.NewManager(cfg.History.Path, cfg.History.Max)
	if err != nil {
		return err
	}

	o.configMgr = configMgr
	o.historyMgr = histMgr

	return nil
}

func (o *Options) handleUse(cmdctx *cmd.Context) error {
	kc, err := o.selectUse(cmdctx)
	if err != nil {
		return err
	}

	return o.use(cmdctx, kc)
}

func (o *Options) selectUse(cmdctx *cmd.Context) (*kubeconfig.KubeConfig, error) {
	var curName string
	cur, ok := o.configMgr.Current()
	if ok {
		curName = cur.Name
	}

	if o.name == "-" {
		lastNamePtr := o.historyMgr.GetLastName(curName)
		if lastNamePtr == nil {
			return nil, errors.New("no last kubeconfig selected")
		}

		name := *lastNamePtr
		kc, ok := o.configMgr.Get(name)
		if !ok {
			return nil, fmt.Errorf("cannot find last kubeconfig %q in history, you should remove history records", name)
		}
		return kc, nil
	}

	if o.name != "" {
		kc, ok := o.configMgr.Get(o.name)
		if !ok {
			err := term.Confirm(o.skipConfirm, "kubeconfig %q not found, do you want to create it", o.name)
			if err != nil {
				return nil, err
			}

			data, err := edit.Edit(cmdctx.Config)
			if err != nil {
				return nil, err
			}

			return o.configMgr.Put(o.name, data)
		}
		return kc, nil
	}

	return o.selectOne(curName)
}

func (o *Options) handleEdit(cmdctx *cmd.Context) error {
	name, err := o.selectEdit()
	if err != nil {
		return err
	}

	data, err := edit.Edit(cmdctx.Config)
	if err != nil {
		return err
	}

	kc, err := o.configMgr.Put(name, data)
	if err != nil {
		return err
	}

	cur, ok := o.configMgr.Current()
	if ok && cur.Name == name {
		return nil
	}

	return o.use(cmdctx, kc)
}

func (o *Options) selectEdit() (string, error) {
	if o.name != "" {
		return o.name, nil
	}

	cur, ok := o.configMgr.Current()
	if ok {
		return cur.Name, nil
	}

	kc, err := o.selectOne("")
	if err != nil {
		return "", err
	}

	return kc.Name, nil
}

func (o *Options) handleDelete() error {
	name, err := o.selectDelete()
	if err != nil {
		return err
	}

	o.historyMgr.DeleteByName(name)
	err = o.historyMgr.Save()
	if err != nil {
		return err
	}

	term.PrintHint("Delete kubeconfig %q", name)
	return o.configMgr.Delete(name)
}

func (o *Options) selectDelete() (string, error) {
	if o.name != "" {
		return o.name, nil
	}

	var curName string
	cur, ok := o.configMgr.Current()
	if ok {
		curName = cur.Name
	}

	kc, err := o.selectOne(curName)
	if err != nil {
		return "", err
	}
	return kc.Name, nil
}

func (o *Options) handleList() error {
	kc, err := o.selectGet()
	if err != nil {
		return err
	}
	if kc != nil {
		fmt.Println(kc.String())
		return nil
	}

	kcs := o.configMgr.List()
	for _, kc := range kcs {
		fmt.Println(kc.String())
	}
	return nil
}

func (o *Options) handleListHistory() error {
	records := o.historyMgr.List()

	kc, err := o.selectGet()
	if err != nil {
		return err
	}
	if kc != nil {
		newRecords := make([]*history.Record, 0)
		for _, record := range records {
			if record.Name == kc.Name {
				newRecords = append(newRecords, record)
			}
		}
		records = newRecords
	}

	for _, record := range records {
		if record.Namespace != "" {
			continue
		}
		fmt.Printf("[%s] %s\n", term.FormatTimestamp(record.Timestamp), record.Name)
	}

	return nil
}

func (o *Options) selectGet() (*kubeconfig.KubeConfig, error) {
	if o.name == "" {
		return nil, nil
	}
	kc, ok := o.configMgr.Get(o.name)
	if !ok {
		return nil, fmt.Errorf("cannot find kubeconfig %q to show", o.name)
	}
	return kc, nil
}

func (o *Options) use(cmdctx *cmd.Context, kc *kubeconfig.KubeConfig) error {
	term.PrintHint("Switch to kubeconfig %q", kc.Name)
	src := kc.GenerateSource("")
	err := source.Apply(cmdctx.Config, src)
	if err != nil {
		return err
	}

	o.historyMgr.Add(kc.Name, "")
	return o.historyMgr.Save()
}

func (o *Options) selectOne(exclude string) (*kubeconfig.KubeConfig, error) {
	kcs := o.configMgr.List()
	items := make([]string, 0, len(kcs))
	for _, kc := range kcs {
		if exclude != "" && kc.Name == exclude {
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

func (o *Options) handleUnuse(cmdctx *cmd.Context) error {
	cur, ok := o.configMgr.Current()
	if !ok {
		return errors.New("no current kubeconfig used, cannot unuse")
	}

	term.PrintHint("Unuse current kubeconfig %q", cur.Name)
	src := kubeconfig.UnsetSource()
	return source.Apply(cmdctx.Config, src)
}
