package quality

import (
	"regexp"
	"strings"

	"github.com/appautomaton/markmaton/internal/model"
)

var linkRegex = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

var shellSignalPhrases = []string{
	"skip to content",
	"skip to main content",
	"search the blog",
	"search docs",
	"primary navigation",
	"repository files navigation",
	"you signed in with another tab or window",
	"you signed out in another tab or window",
	"you switched accounts on another tab or window",
	"please reload this page",
	"dismiss alert",
	"open more actions menu",
	"go to file",
	"top stories",
	"career site cookie settings",
	"loading",
	"birthday mode",
	"notifications",
	"new issue",
}

func Analyze(markdown string, title string, linkCount, imageCount int, usedMainContent, fallbackUsed bool) model.Quality {
	text := visibleText(markdown)
	textLength := len([]rune(text))
	paragraphCount := countParagraphs(markdown)

	linkDensity := 0.0
	if textLength > 0 {
		linkDensity = float64(linkCount) / float64(textLength)
	}

	score := 0.0
	if strings.TrimSpace(title) != "" {
		score += 0.2
	}
	if textLength >= 800 {
		score += 0.4
	} else {
		score += 0.4 * (float64(textLength) / 800.0)
	}
	if paragraphCount >= 3 {
		score += 0.2
	} else {
		score += 0.2 * (float64(paragraphCount) / 3.0)
	}
	switch {
	case linkDensity <= 0.01:
		score += 0.1
	case linkDensity <= 0.02:
		score += 0.07
	case linkDensity <= 0.03:
		score += 0.03
	}
	if imageCount > 0 {
		score += 0.1
	}
	score -= shellSignalPenalty(markdown)
	score -= tableOpeningPenalty(markdown)
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return model.Quality{
		TextLength:      textLength,
		ParagraphCount:  paragraphCount,
		LinkCount:       linkCount,
		ImageCount:      imageCount,
		TitlePresent:    strings.TrimSpace(title) != "",
		LinkDensity:     linkDensity,
		QualityScore:    score,
		UsedMainContent: usedMainContent,
		FallbackUsed:    fallbackUsed,
	}
}

func NeedsFallback(quality model.Quality) bool {
	return quality.TextLength < 80 || quality.ParagraphCount < 1 || quality.QualityScore < 0.18
}

func shellSignalPenalty(markdown string) float64 {
	opening := strings.ToLower(openingText(markdown, 14))
	if opening == "" {
		return 0
	}

	matches := 0
	for _, phrase := range shellSignalPhrases {
		if strings.Contains(opening, phrase) {
			matches++
		}
	}

	penalty := 0.0
	if matches > 0 {
		penalty += 0.06 * float64(matches)
	}
	if strings.Count(opening, "search") >= 2 {
		penalty += 0.04
	}
	if strings.Count(opening, "navigation") >= 1 {
		penalty += 0.04
	}
	if penalty > 0.4 {
		penalty = 0.4
	}

	return penalty
}

func tableOpeningPenalty(markdown string) float64 {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	nonEmpty := make([]string, 0, 12)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		nonEmpty = append(nonEmpty, line)
		if len(nonEmpty) == 12 {
			break
		}
	}

	if len(nonEmpty) == 0 {
		return 0
	}

	tableLines := 0
	pipeCount := 0
	for _, line := range nonEmpty {
		pipeCount += strings.Count(line, "|")
		if strings.HasPrefix(line, "|") || strings.Contains(line, "| --- |") {
			tableLines++
		}
	}

	switch {
	case tableLines >= 3 || pipeCount >= 16:
		return 0.18
	case tableLines >= 2 || pipeCount >= 8:
		return 0.1
	default:
		return 0
	}
}

func visibleText(markdown string) string {
	text := strings.ReplaceAll(markdown, "\r\n", "\n")
	text = linkRegex.ReplaceAllString(text, "$1")
	replacer := strings.NewReplacer(
		"`", "",
		"*", "",
		"_", "",
		"#", "",
		">", "",
	)
	text = replacer.Replace(text)
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func countParagraphs(markdown string) int {
	parts := strings.Split(strings.TrimSpace(markdown), "\n\n")
	count := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}

func openingText(markdown string, maxLines int) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	collected := make([]string, 0, maxLines)

	for _, line := range lines {
		normalized := strings.TrimSpace(visibleText(line))
		if normalized == "" {
			continue
		}
		collected = append(collected, normalized)
		if len(collected) == maxLines {
			break
		}
	}

	return strings.Join(collected, "\n")
}
