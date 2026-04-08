package quality

import (
	"regexp"
	"strings"

	"github.com/appautomaton/markmaton/internal/model"
)

var linkRegex = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

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
	if linkDensity <= 0.03 {
		score += 0.1
	}
	if imageCount > 0 {
		score += 0.1
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
