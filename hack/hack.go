package hack

import (
	_ "embed"
	"strings"
)

//go:embed kw.sh
var bash string

func GetBash(name string) string {
	return strings.ReplaceAll(bash, "{{name}}", name)
}
