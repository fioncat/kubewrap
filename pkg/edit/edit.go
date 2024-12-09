package edit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/fioncat/kubewrap/config"
)

func Edit(cfg *config.Config, initData []byte) ([]byte, error) {
	path, err := createEditFile(initData)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cfg.Editor, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("run editor: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read edit file: %w", err)
	}

	err = os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("remove edit file: %w", err)
	}

	if reflect.DeepEqual(data, initData) {
		return nil, errors.New("edit content not changed")
	}

	return data, nil
}

func createEditFile(initData []byte) (string, error) {
	path := filepath.Join(os.TempDir(), "kubewrap_edit.yaml")
	err := os.WriteFile(path, initData, 0644)
	if err != nil {
		return "", fmt.Errorf("write edit file: %w", err)
	}
	return path, nil
}
