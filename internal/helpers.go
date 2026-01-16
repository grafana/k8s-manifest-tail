package internal

import (
	"fmt"
	"strings"
)

func FormatQuotedList(elements []string) string {
	if len(elements) == 0 {
		return ""
	}

	quoted := make([]string, len(elements))
	for i, element := range elements {
		quoted[i] = fmt.Sprintf("%q", element)
	}

	if len(quoted) == 1 {
		return quoted[0]
	}
	if len(quoted) == 2 {
		return fmt.Sprintf("%s or %s", quoted[0], quoted[1])
	}
	return fmt.Sprintf("%s, or %s", strings.Join(quoted[:len(quoted)-1], ", "), quoted[len(quoted)-1])
}
