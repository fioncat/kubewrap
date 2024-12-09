package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fioncat/kubewrap/config"
	"github.com/fioncat/kubewrap/pkg/dirs"
)

func Apply(cfg *config.Config, src string) error {
	path := cfg.SourceFilePath
	err := dirs.EnsureCreate(filepath.Dir(path))
	if err != nil {
		return err
	}

	err = os.WriteFile(path, []byte(src), 0644)
	if err != nil {
		return fmt.Errorf("write source file: %w", err)
	}

	return nil
}

func Get(cfg *config.Config, noDelete bool) (string, error) {
	path := cfg.SourceFilePath

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read source file: %w", err)
	}

	if !noDelete {
		err = os.Remove(path)
		if err != nil {
			return "", fmt.Errorf("delete source file: %w", err)
		}
	}

	return string(data), nil
}
