package convert

import (
	"strings"
)

func ToMarkdown(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return "", nil
	}

	converter := DefaultBuilder("").Build()
	return converter.ConvertString(html)
}
