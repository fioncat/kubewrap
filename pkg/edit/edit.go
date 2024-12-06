package edit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fioncat/kubewrap/config"
)

func Edit(cfg *config.Config) ([]byte, error) {
	path, err := createEditFile()
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

	return data, nil
}

func createEditFile() (string, error) {
	path := filepath.Join(os.TempDir(), "kubewrap_edit.yaml")
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create edit file: %w", err)
	}
	return path, file.Close()
}
