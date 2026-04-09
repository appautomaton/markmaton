package postprocess

import (
	"regexp"
	"strings"
)

var skipToContentRegex = regexp.MustCompile(`(?im)^\[skip to content\]\(#[^)]+\)\s*$`)
var labelDateCollisionRegex = regexp.MustCompile(`([[:alpha:]])((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},\s+\d{4})`)

var genericListControlPhrases = []string{
	"switch cards to show media",
	"switch cards to hide media",
	"filter",
	"sort",
}

func Markdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = skipToContentRegex.ReplaceAllString(markdown, "")
	markdown = strings.ReplaceAll(markdown, ")[![", ")\n\n[![")
	markdown = labelDateCollisionRegex.ReplaceAllString(markdown, "$1\\\n\\\n$2")

	lines := strings.Split(markdown, "\n")
	lines = removeDuplicateStandaloneCardImages(lines)
	lines = removeGenericListControls(lines)
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	lines = collapseBlankLines(lines)
	markdown = strings.Join(lines, "\n")
	return strings.TrimSpace(markdown)
}

func removeDuplicateStandaloneCardImages(lines []string) []string {
	var out []string

	for i := 0; i < len(lines); i++ {
		current := strings.TrimSpace(lines[i])
		if current == "" || !strings.HasPrefix(current, "![") {
			out = append(out, lines[i])
			continue
		}

		nextIndex := i + 1
		for nextIndex < len(lines) && strings.TrimSpace(lines[nextIndex]) == "" {
			nextIndex++
		}

		if nextIndex < len(lines) {
			next := strings.TrimSpace(lines[nextIndex])
			if strings.HasPrefix(next, "["+current) {
				continue
			}
		}

		out = append(out, lines[i])
	}

	return out
}

func removeGenericListControls(lines []string) []string {
	var out []string

	for _, line := range lines {
		if isGenericControlOnlyLine(line) {
			continue
		}
		out = append(out, line)
	}

	return out
}

func normalizeControlLine(line string) string {
	normalized := strings.ToLower(strings.TrimSpace(line))
	normalized = strings.ReplaceAll(normalized, "\\", " ")
	normalized = strings.NewReplacer("*", "", "_", "", "#", "", "`", "").Replace(normalized)
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func isGenericControlOnlyLine(line string) bool {
	normalized := normalizeControlLine(line)
	if normalized == "" {
		return false
	}

	candidate := normalized
	for _, phrase := range genericListControlPhrases {
		candidate = strings.ReplaceAll(candidate, phrase, "")
	}
	candidate = strings.Join(strings.Fields(candidate), " ")
	return candidate == ""
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
