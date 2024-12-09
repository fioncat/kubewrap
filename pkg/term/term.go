package term

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/fioncat/kubewrap/pkg/fzf"
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

func Confirm(skip bool, format string, args ...any) error {
	if skip {
		return nil
	}
	hint := fmt.Sprintf(format, args...)
	fmt.Printf("%s? (y/n) ", hint)

	var resp string
	fmt.Scanf("%s", &resp)

	if resp == "y" {
		return nil
	}

	return fzf.ErrCanceled
}

func FormatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05")
}
