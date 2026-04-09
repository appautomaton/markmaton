package quality

import "testing"

func TestAnalyzeQuality(t *testing.T) {
	result := Analyze(
		"# Title\n\nFirst paragraph.\n\nSecond paragraph.",
		"Title",
		1,
		1,
		true,
		false,
	)

	if result.TextLength == 0 {
		t.Fatalf("expected text length to be positive")
	}
	if !result.TitlePresent {
		t.Fatalf("expected title to be present")
	}
	if result.QualityScore <= 0 {
		t.Fatalf("expected quality score to be positive")
	}
}

func TestNeedsFallback(t *testing.T) {
	weak := Analyze("tiny", "", 0, 0, true, false)
	if !NeedsFallback(weak) {
		t.Fatalf("expected weak content to require fallback")
	}
}

func TestAnalyzePenalizesShellHeavyOpenings(t *testing.T) {
	shellHeavy := Analyze(
		"## Search the blog\n\nSearch docs\n\n### Suggested\n\nPrimary navigation\n\nYou signed in with another tab or window. Reload to refresh your session.\n\nRepository files navigation",
		"Blog | OpenAI Developers",
		25,
		0,
		true,
		false,
	)
	readable := Analyze(
		"# Launch notes for the Responses API\n\nImportant product updates for developers building production agents.\n\nThis release improves reliability and tool use.",
		"Launch notes for the Responses API",
		3,
		1,
		true,
		false,
	)

	if shellHeavy.QualityScore >= readable.QualityScore {
		t.Fatalf("expected shell-heavy output to score lower than readable output")
	}
	if shellHeavy.QualityScore >= 0.8 {
		t.Fatalf("expected shell-heavy output to lose meaningful score, got %f", shellHeavy.QualityScore)
	}
}

func TestAnalyzePenalizesTableHeavyOpening(t *testing.T) {
	tableHeavy := Analyze(
		"|     |     |     |\n| --- | --- | --- |\n| [Hacker News](https://news.ycombinator.com) | [new](https://news.ycombinator.com/newest) | [past](https://news.ycombinator.com/front) |\n\nA useful comment eventually appears here.",
		"HN thread",
		10,
		0,
		true,
		false,
	)
	readable := Analyze(
		"# Useful thread title\n\nA readable opening paragraph appears before any structured content.\n\nA second paragraph explains the topic.",
		"Useful thread title",
		3,
		0,
		true,
		false,
	)

	if tableHeavy.QualityScore >= readable.QualityScore {
		t.Fatalf("expected table-heavy opening to score lower than readable output")
	}
}

func TestAnalyzePenalizesTopStoriesAndCookieShell(t *testing.T) {
	shellHeavy := Analyze(
		"## Top Stories:\n\n- Story one\n- Story two\n\nCareer Site Cookie Settings\n\nLoading...\n\nActual article starts much later.",
		"Article title",
		12,
		0,
		true,
		false,
	)

	if shellHeavy.QualityScore >= 0.9 {
		t.Fatalf("expected shell-heavy article opening to lose score, got %f", shellHeavy.QualityScore)
	}
}
