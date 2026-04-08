package postprocess

import (
	"regexp"
	"strings"
)

var skipToContentRegex = regexp.MustCompile(`(?im)^\[skip to content\]\(#[^)]+\)\s*$`)

func Markdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = skipToContentRegex.ReplaceAllString(markdown, "")

	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	lines = collapseBlankLines(lines)
	markdown = strings.Join(lines, "\n")
	return strings.TrimSpace(markdown)
}

func collapseBlankLines(lines []string) []string {
	var out []string
	blankCount := 0
	insideFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			insideFence = !insideFence
			blankCount = 0
			out = append(out, line)
			continue
		}

		if !insideFence && trimmed == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
			out = append(out, "")
			continue
		}

		blankCount = 0
		out = append(out, line)
	}

	return out
}
