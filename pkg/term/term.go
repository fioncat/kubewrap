package term

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

func PrintJson(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func PrintHint(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	hint := color.New(color.Bold).Sprint(s)
	prefix := color.New(color.Bold, color.FgGreen).Sprint("==>")
	fmt.Println(prefix, hint)
}
