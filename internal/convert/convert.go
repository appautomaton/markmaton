package convert

import (
	"strings"

	md "github.com/firecrawl/html-to-markdown"
	"github.com/firecrawl/html-to-markdown/plugin"
)

func ToMarkdown(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return "", nil
	}

	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	converter.Use(plugin.RobustCodeBlock())

	return converter.ConvertString(html)
}
