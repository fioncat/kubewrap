package dirs

import (
	"fmt"
	"os"
)

func EnsureCreate(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("mkdir dir: %w", err)
			}
			return nil
		}
		return fmt.Errorf("check dir stat: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}
