package convert

import (
	"regexp"
	"strings"

	md "github.com/firecrawl/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

var (
	linkOnlyLineRegex   = regexp.MustCompile(`^\[[^\]]+\]\([^)]+\)$`)
	multiLinkLineRegex  = regexp.MustCompile(`^(?:\[[^\]]+\]\([^)]+\)(?:\s+|$))+$`)
	pureNumberLineRegex = regexp.MustCompile(`^\d+$`)
	repoChromeLineRegex = regexp.MustCompile(`^\[[^\]]+\]\([^)]+\)/\s+\*\*\[[^\]]+\]\([^)]+\)\*\*(?:\s+\w+)?$`)
	linkLabelRegex      = regexp.MustCompile(`^\[([^\]]+)\]\([^)]+\)$`)
	linkLabelsRegex     = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
)

var standaloneControlLines = map[string]struct{}{
	"add a comment":                  {},
	"copy":                           {},
	"follow":                         {},
	"highest score (default)":        {},
	"improve this answer":            {},
	"improve this question":          {},
	"new issue":                      {},
	"reset to default":               {},
	"share":                          {},
	"sorted by:":                     {},
	"trending (recent votes count more)": {},
	"date modified (newest first)":   {},
	"date created (oldest first)":    {},
}

func DefaultBeforeHookRegistrations() []BeforeHookRegistration {
	return []BeforeHookRegistration{
		newBeforeHookRegistration("drop_button_like_elements", dropButtonLikeElements()),
	}
}

func DefaultAfterHookRegistrations() []AfterHookRegistration {
	return []AfterHookRegistration{
		newAfterHookRegistration("trim_opening_shell_controls", trimOpeningShellControls()),
		newAfterHookRegistration("drop_standalone_control_lines", dropStandaloneControlLines()),
		newAfterHookRegistration("drop_standalone_metric_lines", dropStandaloneMetricLines()),
		newAfterHookRegistration("drop_redundant_opening_heading_echoes", dropRedundantOpeningHeadingEchoes()),
		newAfterHookRegistration("collapse_adjacent_duplicate_lines", collapseAdjacentDuplicateLines()),
	}
}

func newBeforeHookRegistration(name string, hook md.BeforeHook) BeforeHookRegistration {
	return BeforeHookRegistration{
		Name: name,
		Hook: hook,
	}
}

func newAfterHookRegistration(name string, hook md.Afterhook) AfterHookRegistration {
	return AfterHookRegistration{
		Name: name,
		Hook: hook,
	}
}

func dropButtonLikeElements() md.BeforeHook {
	return func(selec *goquery.Selection) {
		selec.Find("button, [role='button']").Each(func(_ int, s *goquery.Selection) {
			s.Remove()
		})
	}
}

func trimOpeningShellControls() md.Afterhook {
	return func(markdown string) string {
		lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
		start := 0
		inFence := false

		for i := 0; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			if isFenceLine(trimmed) {
				inFence = !inFence
			}
			if inFence {
				break
			}
			if trimmed == "" {
				start = i + 1
				continue
			}
			if isBrokenMetricLine(trimmed, nextNonEmptyLine(lines, i+1)) {
				start = nextNonEmptyIndex(lines, i+1) + 1
				i = nextNonEmptyIndex(lines, i+1)
				continue
			}
			if isOpeningShellLine(trimmed) {
				start = i + 1
				continue
			}
			break
		}

		if start == 0 {
			return markdown
		}

		return strings.TrimLeft(strings.Join(lines[start:], "\n"), "\n")
	}
}

func dropStandaloneControlLines() md.Afterhook {
	return func(markdown string) string {
		lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
		filtered := make([]string, 0, len(lines))
		inFence := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if isFenceLine(trimmed) {
				inFence = !inFence
				filtered = append(filtered, line)
				continue
			}
			if !inFence && (isStandaloneControlLine(trimmed) || isStandalonePaginationLine(trimmed)) {
				continue
			}
			filtered = append(filtered, line)
		}

		return strings.Join(filtered, "\n")
	}
}

func dropStandaloneMetricLines() md.Afterhook {
	return func(markdown string) string {
		lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
		filtered := make([]string, 0, len(lines))
		inFence := false

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if isFenceLine(trimmed) {
				inFence = !inFence
				filtered = append(filtered, line)
				continue
			}
			if inFence {
				filtered = append(filtered, line)
				continue
			}
			if isStandaloneMetricLine(trimmed, previousNonEmptyLine(lines, i-1), nextNonEmptyLine(lines, i+1)) {
				continue
			}
			filtered = append(filtered, line)
		}

		return strings.Join(filtered, "\n")
	}
}

func dropRedundantOpeningHeadingEchoes() md.Afterhook {
	return func(markdown string) string {
		lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
		filtered := make([]string, 0, len(lines))
		inFence := false
		headingNormalized := ""
		nonEmptyAfterHeading := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if isFenceLine(trimmed) {
				inFence = !inFence
				filtered = append(filtered, line)
				continue
			}
			if inFence {
				filtered = append(filtered, line)
				continue
			}
			if headingNormalized == "" && strings.HasPrefix(trimmed, "#") {
				headingNormalized = normalizeComparableLine(trimmed)
				filtered = append(filtered, line)
				continue
			}
			if headingNormalized != "" && trimmed != "" {
				nonEmptyAfterHeading++
				if nonEmptyAfterHeading <= 8 &&
					normalizeComparableLine(trimmed) == headingNormalized &&
					strings.Contains(trimmed, "](") {
					continue
				}
			}
			filtered = append(filtered, line)
		}

		return strings.Join(filtered, "\n")
	}
}

func collapseAdjacentDuplicateLines() md.Afterhook {
	return func(markdown string) string {
		lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
		collapsed := make([]string, 0, len(lines))
		inFence := false
		lastNormalized := ""
		blankRun := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if isFenceLine(trimmed) {
				inFence = !inFence
				collapsed = append(collapsed, line)
				lastNormalized = ""
				blankRun = 0
				continue
			}
			if inFence {
				collapsed = append(collapsed, line)
				continue
			}
			if trimmed == "" {
				collapsed = append(collapsed, line)
				blankRun++
				continue
			}
			normalized := normalizeComparableLine(trimmed)
			if normalized != "" && normalized == lastNormalized && blankRun <= 1 && shouldCollapseSpacedDuplicate(trimmed) {
				continue
			}
			collapsed = append(collapsed, line)
			lastNormalized = normalized
			blankRun = 0
		}

		return strings.Join(collapsed, "\n")
	}
}

func isFenceLine(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

func isOpeningShellLine(line string) bool {
	if isStandaloneControlLine(line) {
		return true
	}
	if pureNumberLineRegex.MatchString(line) {
		return true
	}
	if repoChromeLineRegex.MatchString(line) {
		return true
	}
	if linkOnlyLineRegex.MatchString(line) && len([]rune(line)) <= 120 {
		return true
	}
	if strings.HasPrefix(line, "- ") {
		content := strings.TrimSpace(strings.TrimPrefix(line, "- "))
		if isStandaloneControlLine(content) {
			return true
		}
		if linkOnlyLineRegex.MatchString(content) && len([]rune(content)) <= 120 {
			return true
		}
	}
	return false
}

func isStandaloneControlLine(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	if normalized == "" {
		return false
	}
	if _, ok := standaloneControlLines[normalized]; ok {
		return true
	}
	if strings.HasPrefix(normalized, "[timeline](") {
		return true
	}
	if match := linkLabelRegex.FindStringSubmatch(strings.TrimSpace(line)); len(match) == 2 {
		label := strings.ToLower(strings.TrimSpace(match[1]))
		if _, ok := standaloneControlLines[label]; ok {
			return true
		}
	}
	return false
}

func isStandalonePaginationLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !multiLinkLineRegex.MatchString(trimmed) {
		return false
	}
	labels := extractLinkLabels(trimmed)
	if len(labels) == 0 {
		return false
	}
	for _, label := range labels {
		label = strings.ToLower(strings.TrimSpace(label))
		if label == "" {
			return false
		}
		if pureNumberLineRegex.MatchString(label) {
			continue
		}
		switch label {
		case "next", "previous", "prev", "older", "newer":
			continue
		default:
			return false
		}
	}
	return true
}

func isStandaloneMetricLine(line, prevLine, nextLine string) bool {
	if !pureNumberLineRegex.MatchString(strings.TrimSpace(line)) {
		return false
	}
	prevTrimmed := strings.TrimSpace(prevLine)
	nextTrimmed := strings.TrimSpace(nextLine)
	if prevTrimmed == "" && nextTrimmed == "" {
		return false
	}
	contexts := []string{prevTrimmed, nextTrimmed}
	for _, context := range contexts {
		if context == "" {
			continue
		}
		if isStandaloneControlLine(context) || isStandalonePaginationLine(context) {
			return true
		}
		if linkOnlyLineRegex.MatchString(context) || multiLinkLineRegex.MatchString(context) {
			return true
		}
		if strings.HasPrefix(context, "#") {
			return true
		}
	}
	return false
}

func isBrokenMetricLine(line, nextLine string) bool {
	if !strings.HasPrefix(line, "- [") {
		return false
	}
	if strings.Contains(line, "](") {
		return false
	}
	nextLine = strings.TrimSpace(nextLine)
	if nextLine == "" {
		return false
	}
	if !strings.Contains(nextLine, "](") {
		return false
	}
	return true
}

func previousNonEmptyLine(lines []string, start int) string {
	for i := start; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

func nextNonEmptyLine(lines []string, start int) string {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

func nextNonEmptyIndex(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return i
		}
	}
	return start
}

func normalizeComparableLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "#") {
		trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
	}
	if linkLabelsRegex.MatchString(trimmed) {
		trimmed = linkLabelsRegex.ReplaceAllString(trimmed, "$1")
	}
	trimmed = strings.ReplaceAll(trimmed, `\`, "")
	var normalized strings.Builder
	lastSpace := false
	for _, r := range strings.ToLower(trimmed) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			normalized.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			normalized.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(normalized.String())
}

func shouldCollapseSpacedDuplicate(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return false
	}
	return len([]rune(trimmed)) <= 120
}

func extractLinkLabels(line string) []string {
	matches := linkLabelsRegex.FindAllStringSubmatch(line, -1)
	labels := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) == 2 {
			labels = append(labels, match[1])
		}
	}
	return labels
}
