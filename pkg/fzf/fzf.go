package fzf

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrCanceled = errors.New("fzf: canceled by user")

const ExitCodeCanceled = 130

func Search(items []string) (int, error) {
	var inputBuf bytes.Buffer
	inputBuf.Grow(len(items))
	for _, item := range items {
		inputBuf.WriteString(item + "\n")
	}

	var outputBuf bytes.Buffer
	cmd := exec.Command("fzf")
	cmd.Stdin = &inputBuf
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outputBuf

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			code := exitError.ExitCode()
			switch code {
			case ExitCodeCanceled:
				return 0, ErrCanceled

			default:
				return 0, fmt.Errorf("fzf exited with code %d", code)
			}
		}
		return 0, fmt.Errorf("fzf exited with error: %w", err)
	}

	result := outputBuf.String()
	result = strings.TrimSpace(result)
	for idx, item := range items {
		if item == result {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("fzf: cannot find %q", result)
}
